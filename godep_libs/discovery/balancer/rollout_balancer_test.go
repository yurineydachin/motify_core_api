package balancer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"motify_core_api/godep_libs/discovery/provider"
)

const (
	serviceName = "testService"
)

// rolloutProvider is a test provider for rolloutBalancer
type rolloutProvider struct {
	provider.Mock

	mu      sync.RWMutex
	watches map[string][]chan *provider.Event

	watchesReady chan bool
	once         sync.Once
}

func prefixFromFilter(filter provider.KeyFilter) string {
	if filter.Prefix != "" {
		return filter.Prefix
	}

	prefix := fmt.Sprintf("/%s/%s/%s/%s/%s/%s/",
		filter.Namespace, filter.Type, filter.Name, filter.RolloutType, filter.Owner,
		filter.ClusterType)
	// Trim left any "//" in prefix - to clean-up empty parts
	parts := strings.Split(prefix, "//")
	if len(parts) > 1 {
		prefix = parts[0] + "/"
	}

	return prefix
}

func serviceKey(name, rollout, instance string) string {
	return fmt.Sprintf("/%s/%s/%s/%s/%s/%s/%s", provider.NamespaceDiscovery, provider.TypeApp, name, rollout,
		provider.DefaultOwner, provider.DefaultClusterType, instance)
}

func newRolloutProvider() *rolloutProvider {
	p := &rolloutProvider{
		watches:      make(map[string][]chan *provider.Event),
		watchesReady: make(chan bool),
	}
	p.WatchCallback = p.watch
	return p
}

func newRolloutBalancer(p provider.IProvider, opts RolloutBalancerOptions) IRolloutBalancer {
	return NewRolloutBalancer(p, NewRolloutWatcher(p, nil), nil, opts)
}

func (r *rolloutProvider) waitReady() bool {
	select {
	case <-r.watchesReady:
		return true
	case <-time.After(500 * time.Millisecond):
		panic("provider is not ready after 500 msec")
	}
}

func (r *rolloutProvider) watch(ctx context.Context, filter provider.KeyFilter) <-chan *provider.Event {
	prefix := prefixFromFilter(filter)

	r.mu.Lock()
	defer r.mu.Unlock()
	ch := make(chan *provider.Event, 1)
	channels, ok := r.watches[prefix]
	if !ok {
		channels = make([]chan *provider.Event, 0)
	}
	channels = append(channels, ch)
	r.watches[prefix] = channels
	if len(r.watches) >= unstableRolloutCount+2 {
		r.once.Do(func() { close(r.watchesReady) })
	}

	go func() {
		<-ctx.Done()

		r.mu.Lock()
		channels := r.watches[prefix]
		var idx int
		for i, out := range channels {
			if out == ch {
				idx = i
				break
			}
		}
		close(ch)
		channels = append(channels[:idx], channels[idx+1:]...)
		r.watches[prefix] = channels
		r.mu.Unlock()
	}()

	return ch
}

func (r *rolloutProvider) addService(name string, rolloutType string, endpoint string) {
	key := serviceKey(name, rolloutType, endpoint)
	value := fmt.Sprintf(`{"endpoint_main": "%s"}`, endpoint)

	r.mu.RLock()
	defer r.mu.RUnlock()
	for prefix, channels := range r.watches {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		kv := provider.KV{
			Value:  value,
			RawKey: key,
			Service: provider.Service{
				InstanceName: provider.NewInstanceNameFromString(endpoint),
			},
		}
		event := &provider.Event{
			Type: provider.EventPut,
			KVs:  []provider.KV{kv},
		}
		for _, ch := range channels {
			ch <- event
		}
		break
	}
}

func (r *rolloutProvider) deleteService(name string, rolloutType string, endpoint string) {
	key := serviceKey(name, rolloutType, endpoint)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for prefix, channels := range r.watches {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		kv := provider.KV{
			RawKey: key,
			Service: provider.Service{
				InstanceName: provider.NewInstanceNameFromString(endpoint),
			},
		}
		event := &provider.Event{
			Type: provider.EventDelete,
			KVs:  []provider.KV{kv},
		}
		for _, ch := range channels {
			ch <- event
		}
		break
	}
}

func (r *rolloutProvider) setKV(key, value string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for prefix, channels := range r.watches {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		kv := provider.KV{
			RawKey: key,
			Value:  value,
		}
		event := &provider.Event{
			Type: provider.EventPut,
			KVs:  []provider.KV{kv},
		}
		for _, ch := range channels {
			ch <- event
		}
		break
	}
}

// This is a fucking workaround, but no way of doing it for now...
func waitForStatsChange(b IRolloutBalancer, rolloutType string, wantedLen int) {
	deadline := time.After(500 * time.Millisecond)
	for {
		stats := b.Stats(StatsOptions{RolloutType: rolloutType})
		if len(stats) == wantedLen {
			return
		}
		select {
		case <-deadline:
			panic(fmt.Errorf("stats not changed! Last stats: %#v", stats))
		case <-time.After(1 * time.Millisecond):
			// poll stats
		}
	}
}

// Also a dirty hack
func balancerWaitForRolloutsChange(b IRolloutBalancer, wantedLen int) {
	r, ok := b.(*rolloutBalancer)
	if !ok {
		panic("rolloutBalancer instance is expected")
	}
	waitForRolloutsChange(r.watcher, wantedLen)
}

// Also a dirty hack
func waitForRolloutsChange(w IRolloutWatcher, wantedLen int) {
	r, ok := w.(*rolloutWatcher)
	if !ok {
		panic("rolloutWatcher instance is expected")
	}
	deadline := time.After(500 * time.Millisecond)
	for {
		r.mu.RLock()
		rollouts := r.rolloutTypes
		if len(rollouts) == wantedLen {
			return
		}
		r.mu.RUnlock()

		select {
		case <-deadline:
			r.mu.RLock()
			defer r.mu.RUnlock()
			panic(fmt.Errorf("rollouts not changed! Last: %#v", rollouts))
		case <-time.After(1 * time.Millisecond):
			// poll rollouts
		}
	}
}

func checkNext(addr string, err error, expected string, t *testing.T) {
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if addr != expected {
		t.Fatalf("got %q, expected %q", addr, expected)
	}
}

func TestRolloutBalancer_NewRolloutBalancer(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)

	if b == nil {
		t.Fatalf("balancer is nil")
	}
	if _, ok := b.(*rolloutBalancer); !ok {
		t.Fatalf("unexpected NewRolloutBalancer core type")
	}
	b.Stop()
}

func TestRolloutBalancer_NewRolloutBalancer_Fallback(t *testing.T) {
	p := newRolloutProvider()
	balancers := []ILoadBalancer{
		newInMemBalancer([]string{"1", "2", "3"}),
		newInMemBalancer([]string{"4", "5", "6"}),
	}
	fallbackBalancer := NewFallbackBalancer(nil, balancers...)

	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
		FallbackBalancer: fallbackBalancer,
	}
	b := newRolloutBalancer(p, opts)

	if b == nil {
		t.Fatalf("balancer is nil")
	}
	rb, ok := b.(*rolloutBalancer)
	if !ok {
		t.Fatalf("unexpected NewRolloutBalancer core type")
	}
	if _, ok := rb.stableBalancer.(*FallbackBalancer); !ok {
		t.Fatalf("unexpected stable balancer type")
	}
	b.Stop()
}

func TestRolloutBalancer_Next_Fallback(t *testing.T) {
	p := newRolloutProvider()
	balancers := []ILoadBalancer{
		newInMemBalancer([]string{"1", "2", "3"}),
		newInMemBalancer([]string{"4", "5", "6"}),
	}
	fallbackBalancer := NewFallbackBalancer(nil, balancers...)

	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
		FallbackBalancer: fallbackBalancer,
	}
	b := newRolloutBalancer(p, opts)
	defer b.Stop()

	t.Logf("Fallback balancer should be used")
	addr, err := b.Next(NextOptions{})
	checkNext(addr, err, "1", t)
	addr, err = b.Next(NextOptions{})
	checkNext(addr, err, "2", t)
}

// TODO: parallel
func TestRolloutBalancer_Next_NoData(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	defer b.Stop()

	_, err := b.Next(NextOptions{})
	if err == nil {
		t.Fatalf("error should be returned")
	}
}

func TestRolloutBalancer_Next_Stable(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	p.addService(serviceName, "stable", "host1")
	p.addService(serviceName, "stable", "host2")
	waitForStatsChange(b, "stable", 2)

	addr, err := b.Next(NextOptions{})
	checkNext(addr, err, "host1", t)

	addr, err = b.Next(NextOptions{})
	checkNext(addr, err, "host2", t)

	addr, err = b.Next(NextOptions{})
	checkNext(addr, err, "host1", t)
}

func TestRolloutBalancer_Next_UnstableNoSegregation(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	p.addService(serviceName, "stable", "host1")
	p.addService(serviceName, "stable", "host2")
	p.addService(serviceName, "unstable1", "unstableHost")
	waitForStatsChange(b, "", 3)

	t.Logf("Stable service should be returned, because no segregation ID data is added")
	addr, err := b.Next(NextOptions{})
	checkNext(addr, err, "host1", t)
}

func TestRolloutBalancer_Next_Unstable(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	p.addService(serviceName, "stable", "host1")
	p.addService(serviceName, "unstable1", "unstableHost")
	waitForStatsChange(b, "", 2)
	t.Logf("Int format should not make any difference")
	p.setKV(rolloutPrefix+"001", "unstable1")
	p.setKV(rolloutPrefix+"2", "unstable1")
	balancerWaitForRolloutsChange(b, 2)

	addr, err := b.Next(NextOptions{SegregationID: "1"})
	checkNext(addr, err, "unstableHost", t)
	addr, err = b.Next(NextOptions{SegregationID: "2"})
	checkNext(addr, err, "unstableHost", t)
}

// TODO: parallel
func TestRolloutBalancer_Next_UnstableIsUnavailable(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	p.addService(serviceName, "stable", "host1")
	waitForStatsChange(b, "", 1)
	p.setKV(rolloutPrefix+"001", "unstable1")
	balancerWaitForRolloutsChange(b, 1)

	t.Logf("If unstable service is unavailable - no error should occur and stable should be returned")
	addr, err := b.Next(NextOptions{SegregationID: "1"})
	checkNext(addr, err, "host1", t)
}

func TestRolloutBalancer_Next_UnstableBecomesUnavailable(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	p.addService(serviceName, "stable", "host1")
	p.addService(serviceName, "unstable20", "unstableHost")
	waitForStatsChange(b, "", 2)
	p.setKV(rolloutPrefix+"001", "unstable20")
	balancerWaitForRolloutsChange(b, 1)

	addr, err := b.Next(NextOptions{SegregationID: "1"})
	checkNext(addr, err, "unstableHost", t)

	p.deleteService(serviceName, "unstable20", "unstableHost")
	waitForStatsChange(b, "", 1)

	t.Logf("If unstable service is unavailable - no error should occur and stable should be returned")
	addr, err = b.Next(NextOptions{SegregationID: "1"})
	checkNext(addr, err, "host1", t)
}

func TestRolloutBalancer_Next_SegregationIDUnparsed(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	p.addService(serviceName, "stable", "host1")
	p.addService(serviceName, "unstable20", "unstableHost")
	waitForStatsChange(b, "", 2)
	p.setKV(rolloutPrefix+"foooo", "unstable20")
	p.setKV(rolloutPrefix+"1000", "stable")
	balancerWaitForRolloutsChange(b, 1)

	addr, err := b.Next(NextOptions{SegregationID: "1"})
	checkNext(addr, err, "host1", t)
}

func TestRolloutBalancer_Next_InvalidUnstableService(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}
	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	t.Logf("If RE release unstable service with invalid name - stable should be returned")
	p.addService(serviceName, "stable", "host1")
	p.addService(serviceName, "unstable2000", "unstableHost")
	waitForStatsChange(b, "", 1)
	p.setKV(rolloutPrefix+"1", "unstable2000")
	p.setKV(rolloutPrefix+"1000", "stable")
	balancerWaitForRolloutsChange(b, 2)

	addr, err := b.Next(NextOptions{SegregationID: "1"})
	checkNext(addr, err, "host1", t)
}

func TestRolloutBalancer_Stats(t *testing.T) {
	p := newRolloutProvider()
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: serviceName,
		},
	}

	b := newRolloutBalancer(p, opts)
	p.waitReady()
	defer b.Stop()

	p.addService(serviceName, "stable", "host1")
	p.addService(serviceName, "stable", "host2")
	p.addService(serviceName, "unstable1", "unstableHost1")
	p.addService(serviceName, "unstable2", "unstableHost2")
	waitForStatsChange(b, "", 4)

	stats := b.Stats(StatsOptions{})
	t.Logf("Total stats should be returned")
	if len(stats) != 4 {
		t.Fatalf("Unexpected stats: %#v", stats)
	}

	stats = b.Stats(StatsOptions{RolloutType: provider.RolloutTypeStable})
	t.Logf("stable stats should be returned")
	if len(stats) != 2 {
		t.Fatalf("Unexpected stats: %#v", stats)
	}
}
