package balancer

import (
	"context"
	"sort"
	"sync"
	"time"

	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/locator"
	"motify_core_api/godep_libs/discovery/provider"
)

const (
	initTimeout         = time.Millisecond * 500
	healthcheckInterval = time.Millisecond * 1000
	healthcheckTimeout  = time.Millisecond * 500
	maxResponseTime     = time.Minute
)

// CheckerFunc is a function to check node's health
type CheckerFunc func(address string, timeout time.Duration) error

// WeightedRoundRobin is a ILoadBalancer implementing round-robin balancing algorithm
// with additional healthcheck-based priority scheme
type WeightedRoundRobin struct {
	serviceName  string
	endpointType locator.EndpointType
	filter       locator.KeyFilter

	checker CheckerFunc

	nodes      Nodes
	nodesByKey NodesByKey

	// Index of current rolling service used for round-robin balancing
	index int
	// current weight
	cw        uint64
	gcdVal    uint64
	maxWeight uint64
	sumWeight uint64

	readyOnce sync.Once
	ready     chan struct{}
	startedAt time.Time
	mx        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc

	logger  discovery.ILogger
	logName string

	updates chan struct{}
	upmx    sync.Mutex
}

var _ ILoadBalancer = (*WeightedRoundRobin)(nil)

// NewWeightedRoundRobinEtcd2 returns legacy weighted round robin balancer from IServiceLocator2
// TODO: purge after dropping etcd2 support
func NewWeightedRoundRobinEtcd2(l locator.IServiceLocator2, logger discovery.ILogger, info locator.LocationInfo) *WeightedRoundRobin {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}

	return NewWeightedRoundRobin(locator.NewAdapter(l, info, logger), logger, LoadBalancerOptions{ServiceName: info.ServiceName})
}

// NewWeightedRoundRobin returns new NewWeightedRoundRobin balancer instance.
func NewWeightedRoundRobin(l locator.ILocator, logger discovery.ILogger, opts LoadBalancerOptions) *WeightedRoundRobin {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}
	opts.defaultize()

	ctx, cancel := context.WithCancel(context.Background())
	b := &WeightedRoundRobin{
		logger:       logger,
		serviceName:  opts.ServiceName,
		filter:       *opts.Filter,
		endpointType: opts.EndpointType,
		checker:      DialCheck,
		index:        -1,
		ready:        make(chan struct{}),
		nodesByKey:   make(NodesByKey),
		ctx:          ctx,
		cancel:       cancel,
	}

	// We want to log not only service name, but also requested RolloutType
	b.logName = b.serviceName + " " + b.filter.RolloutType

	go b.consume(l)
	return b
}

// String returns string representation of balancer for logging
func (b *WeightedRoundRobin) String() string {
	return b.logName
}

func (b *WeightedRoundRobin) consume(locator locator.ILocator) {
	b.startedAt = time.Now()
	time.AfterFunc(initTimeout, func() {
		b.readyOnce.Do(func() {
			b.logger.Debugf("%q balancer failed to receive updates during %s timeout", b, initTimeout)
			close(b.ready)
		})
	})

	events := locator.Watch(b.ctx, b.serviceName, b.endpointType, &b.filter)
	for {
		select {
		case event, ok := <-events:
			if !ok {
				b.logger.Error("locator channel closed unexpectedly")
				return
			}
			b.handleEvent(event)
		case <-b.ctx.Done():
			b.logger.Debugf("%q consumer is stopped.", b)
			return
		}
	}
}

func (b *WeightedRoundRobin) handleEvent(event *locator.Event) {
	defer b.update()

	// Any event leads data access, so makes sense to lock here
	b.mx.Lock()
	switch event.Type {
	case provider.EventPut:
		b.putNodes(event.Locations)

		b.readyOnce.Do(func() {
			b.logger.Infof("%q balancer init done in %s", b, time.Since(b.startedAt))
			go b.healthchecker()
			close(b.ready)
		})
	case provider.EventDelete:
		b.deleteNodes(event.Locations)
	}
	b.mx.Unlock()
}

func (b *WeightedRoundRobin) putNodes(locations []locator.Location) {
	modified := false
	for _, location := range locations {
		if b.putNode(location) {
			modified = true
		}
	}
	if modified {
		b.updateWeights()
	}
}

func (b *WeightedRoundRobin) putNode(location locator.Location) bool {
	key := location.Service.InstanceKey()
	node := newNode(key, location.Endpoint)

	if oldNode := b.nodesByKey[key]; oldNode != nil {
		// No need to update the node if URL's are equal
		if oldNode.URL != node.URL {
			b.logger.Debugf("%q updating %s", b, location.Endpoint)
			b.updateNode(node)
			b.nodesByKey[node.Key] = node
			return true
		}
	} else {
		b.logger.Debugf("%q appending %s", b, location.Endpoint)
		b.nodes = append(b.nodes, node)
		b.nodesByKey[node.Key] = node
		return true
	}

	return false
}

func (b *WeightedRoundRobin) updateNode(node *Node) {
	i := b.nodes.indexOf(node.Key)
	if i >= 0 {
		// shift left and add new node to the end because it has max response time by default
		b.nodes = b.nodes.removeByIndex(i)
		b.nodes = append(b.nodes, node)
	} else {
		// append node if not found, but it should not actually happen
		b.logger.Infof("node %s is indexed but not found", node.Key)
		b.nodes = append(b.nodes, node)
	}
}

func (b *WeightedRoundRobin) deleteNodes(locations []locator.Location) {
	modified := false
	for _, location := range locations {
		if b.deleteNode(location) {
			modified = true
		}
	}
	if modified {
		b.updateWeights()
	}
}

func (b *WeightedRoundRobin) deleteNode(location locator.Location) bool {
	key := location.Service.InstanceKey()
	b.logger.Debugf("%q deleting node: %s", b, key)

	if b.nodesByKey[key] == nil {
		b.logger.Infof("node %s is already deleted", key)
		return false
	}

	delete(b.nodesByKey, key)
	if i := b.nodes.indexOf(key); i != -1 {
		b.nodes = b.nodes.removeByIndex(i)
		return true
	}
	return false
}

type healthcheckStat struct {
	RTT   time.Duration
	State nodeState
}

func (b *WeightedRoundRobin) healthchecker() {
	for {
		var copiedNodes Nodes
		var stats []healthcheckStat
		var wg sync.WaitGroup

		b.mx.RLock()
		count := len(b.nodes)
		if count == 0 {
			b.mx.RUnlock()
			goto SLEEP
		}
		copiedNodes = make(Nodes, count)
		copy(copiedNodes, b.nodes)
		b.mx.RUnlock()

		stats = make([]healthcheckStat, count)
		wg.Add(count)
		for i, node := range copiedNodes {
			go func(url string, stat *healthcheckStat) {
				defer wg.Done()
				b.checkNode(url, stat)
			}(node.address, &stats[i])
		}

		go func() {
			wg.Wait()

			select {
			case <-b.ctx.Done():
				return
			default:
			}

			b.mx.Lock()
			defer b.mx.Unlock()

			for i, node := range copiedNodes {
				r := &stats[i]
				node.state = r.State
				node.stats.Add(r.RTT)
			}
			sort.Sort(b.nodes)
			b.updateWeights()
		}()

	SLEEP:
		select {
		case <-b.ctx.Done():
			return
		case <-time.After(healthcheckInterval):
		}
	}
}

func (b *WeightedRoundRobin) checkNode(address string, stat *healthcheckStat) {
	start := time.Now()
	if err := b.checker(address, healthcheckTimeout); err != nil {
		stat.RTT = maxResponseTime
		stat.State = Unhealthy
	} else {
		stat.RTT = time.Since(start)
		stat.State = Healthy
	}
}

func (b *WeightedRoundRobin) updateWeights() {
	if len(b.nodes) == 0 {
		b.index = -1
		b.cw = 0
		b.maxWeight = 0
		b.sumWeight = 0
		return
	}

	/*
		// 1.5 is empiracally chosen coefficient
		n := nextPowerOfTwo(3 * len(b.nodes) / 2)
		minRTT := b.nodes[0].stats.avg
		b.sumWeight = 0

		for _, node := range b.nodes {
			// sumRTT / nodeRTT / maxWeight where maxWeight = sumRTT / minRTT * numOfNodes
			node.weight = uint64(float64(minRTT) / float64(node.stats.avg) * float64(n))
			b.sumWeight += node.weight
		}

		b.maxWeight = b.nodes[0].weight
		if b.cw > b.maxWeight {
			b.index = -1
			b.cw = 0
		}
	*/

	var sumRTT uint64
	for _, node := range b.nodes {
		sumRTT += uint64(node.stats.avg)
	}

	var sumWeight uint64
	for _, node := range b.nodes {
		node.weight = uint64((float64(sumRTT) / float64(node.stats.avg)) + 0.5)
		sumWeight += node.weight
	}

	/*
		n := nextPowerOfTwo(len(b.nodes))
		k := float64(sumWeight) / float64(n)
		sumWeight = 0
		for _, node := range b.nodes {
			node.weight = uint64(float64(node.weight)/k + 0.5)
			sumWeight += node.weight
		}
	*/
	b.sumWeight = sumWeight

	b.maxWeight = b.nodes[0].weight
	if b.cw > b.maxWeight {
		b.index = -1
		b.cw = 0
	}
}

// Next returns next address for balancing
func (b *WeightedRoundRobin) Next() (string, error) {
	if err := b.waitForInitializing(); err != nil {
		return "", err
	}

	b.mx.Lock()
	defer b.mx.Unlock()

	i, err := b.nextIndex()
	if err != nil {
		return "", err
	}
	return b.nodes[i].URL, nil
}

func (b *WeightedRoundRobin) nextIndex() (int, error) {
	count := len(b.nodes)
	if count == 0 || b.maxWeight == 0 {
		return 0, newErrNoServiceAvailable(b.String())
	}

	i := b.index
	cw := b.cw

	i = (i + 1) % len(b.nodes)
	if i == 0 {
		if cw <= 1 {
			cw = b.maxWeight
		} else {
			cw--
		}
	}

	node := b.nodes[i]
	if node.weight < cw {
		// all weights are less than cw and we need to start over
		i = 0
		node = b.nodes[0]
		if cw <= 1 {
			cw = b.maxWeight
		} else {
			cw--
		}
	}

	b.index = i
	b.cw = cw
	node.count++
	return i, nil
}

func (b *WeightedRoundRobin) waitForInitializing() error {
	select {
	case <-b.ready:
		// Initialized. Ready to work.
	case <-b.ctx.Done():
		b.logger.Debugf("%q balancer is stopped.", b)
		return newErrNoServiceAvailable(b.String())
	}
	return nil
}

// Stats returns balancing node statistics
func (b *WeightedRoundRobin) Stats() []NodeStat {
	var stats []NodeStat

	b.mx.RLock()
	for _, v := range b.nodes {
		healthy := v.state == Healthy
		p := float64(v.weight) / float64(b.sumWeight)
		stats = append(stats, NodeStat{
			Key:            v.Key,
			Value:          v.URL,
			Healthy:        healthy,
			HitCount:       v.count,
			HitProbability: p,
			RTT:            v.stats.Current(),
			RTTAverage:     v.stats.avg,
		})
	}
	b.mx.RUnlock()

	return stats
}

// Stop stops the balancer's discovery loop
func (b *WeightedRoundRobin) Stop() {
	b.cancel()
}

// ServiceName return service name that is balancer is pointing to
func (b *WeightedRoundRobin) ServiceName() string {
	return b.serviceName
}

// Notify registers a channel. This channel will be pushed for any updates of available addresses
func (b *WeightedRoundRobin) Notify(updates chan struct{}) {
	b.upmx.Lock()
	defer b.upmx.Unlock()
	b.updates = updates

	go func() {
		// notify about updates after init
		// this is a fallback in case of Notify will be called after init
		b.waitForInitializing()
		b.update()
	}()
}

func (b *WeightedRoundRobin) update() {
	b.upmx.Lock()
	defer b.upmx.Unlock()
	if b.updates != nil {
		b.updates <- struct{}{}
	}
}

//func nextPowerOfTwo(n int) int {
//	n--
//	n |= n >> 1
//	n |= n >> 2
//	n |= n >> 4
//	n |= n >> 8
//	n |= n >> 16
//	n++
//	return n
//}
