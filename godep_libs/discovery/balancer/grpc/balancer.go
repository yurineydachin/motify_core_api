package grpc

import (
	"sync"
	"time"

	"errors"
	"motify_core_api/godep_libs/discovery/balancer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type notifier interface {
	// Notify registers a channel. This channel will be pushed for any updates of available addresses
	Notify(chan struct{})
}

var _ grpc.Balancer = (*Balancer)(nil)

// Balancer implements grpc.Balancer interface
// It's wrapper which makes balancer.ILoadBalancer compatible with grpc.Balancer
type Balancer struct {
	balancer  balancer.ILoadBalancer
	conns     map[string]bool
	connsMx   sync.RWMutex
	addrCh    chan []grpc.Address
	updatesCh chan struct{}
	closeCh   chan struct{}
	closeOnce sync.Once
	startOnce sync.Once
}

// NewBalancer creates new grpc.Balancer from balancer.ILoadBalancer
func NewBalancer(b balancer.ILoadBalancer) *Balancer {
	return &Balancer{
		balancer:  b,
		addrCh:    make(chan []grpc.Address),
		updatesCh: make(chan struct{}),
		closeCh:   make(chan struct{}),
		conns:     make(map[string]bool),
	}
}

// Start waits balancer.ILoadBalancer ready and subscribes for nodes updates
func (b *Balancer) Start(target string, config grpc.BalancerConfig) error {
	s := b.balancer.(notifier)
	s.Notify(b.updatesCh)
	b.startOnce.Do(func() {
		go b.runUpdater()
	})
	return nil
}

// Up informs the Balancer that gRPC has a connection to the server at addr
// It returns func which is called once the connection to addr gets lost or closed
func (b *Balancer) Up(addr grpc.Address) func(error) {
	b.connsMx.Lock()
	b.conns[addr.Addr] = true
	b.connsMx.Unlock()
	return func(error) {
		b.connsMx.Lock()
		delete(b.conns, addr.Addr)
		b.connsMx.Unlock()
	}
}

// addrIsActive returns bool (true if gRPC has a connection to the server at addr) and int (number of active connections)
func (b *Balancer) addrIsActive(addr string) (bool, int) {
	b.connsMx.RLock()
	defer b.connsMx.RUnlock()
	if up, ok := b.conns[addr]; ok && up {
		return true, len(b.conns)
	}
	return false, len(b.conns)
}

// Get returns Next available node
func (b *Balancer) Get(ctx context.Context, opts grpc.BalancerGetOptions) (a grpc.Address, put func(), err error) {
	for {
		a.Addr, err = b.balancer.Next()
		if err == nil {
			ok, n := b.addrIsActive(a.Addr)
			if ok {
				return
			}
			if n > 0 {
				continue
			}
		}
		if !opts.BlockingWait {
			return // Here inactive address will be returned or error from balancer
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-time.After(50 * time.Millisecond):
			// Wait active connection
		}
	}
}

// Notify returns chan in which Balancer sends whole list of available nodes
func (b *Balancer) Notify() <-chan []grpc.Address {
	return b.addrCh
}

// Close Balancer. This method needs only for satisfy grpc.Balancer. For real stopping balancer use Stop method
func (b *Balancer) Close() error {
	// This method is denied because we want to reuse Balancer in grpc.Dial method,
	// but it's not possible because grpc.Dial closes balancer in case of fail
	return errors.New("denied")
}

// Stop Balancer
func (b *Balancer) Stop() {
	b.closeOnce.Do(func() {
		b.balancer.Stop()
		s := b.balancer.(notifier)
		s.Notify(nil)
		close(b.closeCh)
	})
}

// Stats returns node statistics
func (b *Balancer) Stats() (nodes []balancer.NodeStat) {
	stats := b.balancer.Stats()
	b.connsMx.RLock()
	defer b.connsMx.RUnlock()
	for _, row := range stats {
		up, ok := b.conns[row.Value]
		row.Connected = ok && up
		nodes = append(nodes, row)
	}
	return nodes
}

// runUpdater waits the updates from balancer.ILoadBalancer and gets address list from balancer.ILoadBalancer.Stats
func (b *Balancer) runUpdater() {
	for {
		select {
		case <-b.updatesCh:
			b.addrCh <- b.getAddressesFromStats()
		case <-b.closeCh:
			close(b.addrCh)
			close(b.updatesCh)
			return
		}
	}
}

// getAddressesFromStats returns address list from balancer.ILoadBalancer.Stats
func (b *Balancer) getAddressesFromStats() (addresses []grpc.Address) {
	for _, row := range b.balancer.Stats() {
		addresses = append(addresses, grpc.Address{Addr: row.Value})
	}
	return
}
