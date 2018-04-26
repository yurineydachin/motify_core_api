package balancer

import (
	"context"
	"sync"
	"time"

	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/locator"
	"motify_core_api/godep_libs/discovery/provider"
)

// RoundRobin is a ILoadBalancer implementing round-robin balancing algorythm
type RoundRobin struct {
	serviceName  string
	endpointType locator.EndpointType
	filter       locator.KeyFilter

	nodes      Nodes
	nodesByKey NodesByKey
	// Index of current rolling service used for round-robin balancing
	index  int
	mx     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	readyOnce sync.Once
	ready     chan struct{}
	startedAt time.Time

	logger  discovery.ILogger
	logName string

	updates chan struct{}
	upmx    sync.Mutex
}

var _ ILoadBalancer = (*RoundRobin)(nil)

// NewRoundRobinEtcd2 returns legacy round robin balancer from IServiceLocator2
// TODO: purge after dropping etcd2 support
func NewRoundRobinEtcd2(l locator.IServiceLocator2, logger discovery.ILogger, info locator.LocationInfo) *RoundRobin {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}

	return NewRoundRobin(locator.NewAdapter(l, info, logger), logger, LoadBalancerOptions{ServiceName: info.ServiceName})
}

// NewRoundRobin returns new RoundRobin balancer instance.
func NewRoundRobin(l locator.ILocator, logger discovery.ILogger, opts LoadBalancerOptions) *RoundRobin {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}
	opts.defaultize()

	ctx, cancel := context.WithCancel(context.Background())
	b := &RoundRobin{
		logger:       logger,
		serviceName:  opts.ServiceName,
		filter:       *opts.Filter,
		endpointType: opts.EndpointType,
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
func (b *RoundRobin) String() string {
	return b.logName
}

// Next returns next address for balancing
func (b *RoundRobin) Next() (string, error) {
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

func (b *RoundRobin) nextIndex() (int, error) {
	count := len(b.nodes)
	if count == 0 {
		return 0, newErrNoServiceAvailable(b.String())
	}
	i := b.index
	i = (i + 1) % count
	b.index = i
	b.nodes[i].count++
	return i, nil
}

func (b *RoundRobin) waitForInitializing() error {
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
func (b *RoundRobin) Stats() []NodeStat {
	var stats []NodeStat

	b.mx.RLock()
	if len(b.nodes) > 0 {
		p := 100. / float64(len(b.nodes))
		for _, v := range b.nodes {
			stats = append(stats, NodeStat{
				Key:            v.Key,
				Value:          v.URL,
				Healthy:        true,
				HitCount:       v.count,
				HitProbability: p,
				RTT:            0,
				RTTAverage:     0,
			})
		}
	}
	b.mx.RUnlock()

	return stats
}

// Stop stops the balancer's discovery loop
func (b *RoundRobin) Stop() {
	b.cancel()
}

func (b *RoundRobin) consume(locator locator.ILocator) {
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

func (b *RoundRobin) handleEvent(event *locator.Event) {
	defer b.update()

	// Any event leads data access, so makes sense to lock here
	b.mx.Lock()
	switch event.Type {
	case provider.EventPut:
		for _, location := range event.Locations {
			b.putNode(location)
		}
		b.readyOnce.Do(func() {
			b.logger.Infof("%q balancer init done in %s", b, time.Since(b.startedAt))
			close(b.ready)
		})
	case provider.EventDelete:
		for _, location := range event.Locations {
			b.deleteNode(location)
		}
	}
	b.mx.Unlock()
}

func (b *RoundRobin) putNode(location locator.Location) {
	key := location.Service.InstanceKey()
	node := newNode(key, location.Endpoint)

	if oldNode := b.nodesByKey[key]; oldNode != nil {
		// No need to update the node if URL's are equal
		if oldNode.URL != node.URL {
			b.logger.Debugf("%q updating %s", b, location.Endpoint)
			b.updateNode(node)
			b.nodesByKey[node.Key] = node
		}
	} else {
		b.logger.Debugf("%q appending %s", b, location.Endpoint)
		b.nodes = append(b.nodes, node)
		b.nodesByKey[node.Key] = node
	}
}

func (b *RoundRobin) updateNode(node *Node) {
	i := b.nodes.indexOf(node.Key)
	if i >= 0 {
		b.nodes[i] = node
	} else {
		// append node if not found, but it should not actually happen
		b.logger.Infof("node %s is indexed but not found", node.Key)
		b.nodes = append(b.nodes, node)
	}
}

func (b *RoundRobin) deleteNode(location locator.Location) {
	key := location.Service.InstanceKey()
	b.logger.Debugf("%q deleting node: %s", b, key)

	if b.nodesByKey[key] == nil {
		b.logger.Infof("node %s is already deleted", key)
		return
	}

	delete(b.nodesByKey, key)
	if i := b.nodes.indexOf(key); i != -1 {
		b.nodes = b.nodes.removeByIndex(i)
	}
}

// Notify registers a channel. This channel will be pushed for any updates of available addresses
func (b *RoundRobin) Notify(updates chan struct{}) {
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

// ServiceName return service name that is balancer is pointing to
func (b *RoundRobin) ServiceName() string {
	return b.serviceName
}

func (b *RoundRobin) update() {
	b.upmx.Lock()
	defer b.upmx.Unlock()
	if b.updates != nil {
		b.updates <- struct{}{}
	}
}
