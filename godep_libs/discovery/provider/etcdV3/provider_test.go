package etcdV3

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/stretchr/testify/mock"
	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/provider"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
)

const (
	chanTimeout = 100 * time.Millisecond
)

type kv struct {
	k string
	v string
}

type callConfig struct {
	// opCount is a number of etcd options for etcd.Get call. Needed for mocking
	opCount int
	rev     int64
}

type callOption func(*callConfig)

func withOpCount(c int) callOption { return func(cfg *callConfig) { cfg.opCount = c } }
func withRev(rev int64) callOption { return func(cfg *callConfig) { cfg.rev = rev } }

type TestProviderSuit struct{}

var _ = Suite(&TestProviderSuit{})

func (s *TestProviderSuit) newServiceKey() serviceKey {
	return serviceKey{
		Namespace: "test",
		Service: provider.Service{
			Name:         "test_service",
			Type:         provider.TypeApp,
			Owner:        provider.DefaultOwner,
			RolloutType:  provider.RolloutTypeStable,
			ClusterType:  provider.DefaultClusterType,
			InstanceName: provider.NewInstanceNameFromString("host:port"),
		},
	}
}

func (s *TestProviderSuit) newInstanceKey(instance string) serviceKey {
	return serviceKey{
		Namespace: "test",
		Service: provider.Service{
			Name:         "test_service",
			Type:         provider.TypeApp,
			Owner:        provider.DefaultOwner,
			RolloutType:  provider.RolloutTypeStable,
			ClusterType:  provider.DefaultClusterType,
			InstanceName: provider.NewInstanceNameFromString(instance),
		},
	}
}

func (s *TestProviderSuit) getResponse(kvs []kv, rev int64) *clientv3.GetResponse {
	r := &clientv3.GetResponse{
		Kvs: make([]*mvccpb.KeyValue, 0, len(kvs)),
		Header: &etcdserverpb.ResponseHeader{
			Revision: rev,
		},
	}
	for _, kv := range kvs {
		r.Kvs = append(r.Kvs, &mvccpb.KeyValue{
			Key:   []byte(kv.k),
			Value: []byte(kv.v),
		})
	}
	return r
}

func (s *TestProviderSuit) watchEvent(t mvccpb.Event_EventType, k, v string) *clientv3.Event {
	return &clientv3.Event{
		Type: t,
		Kv: &mvccpb.KeyValue{
			Key:   []byte(k),
			Value: []byte(v),
		},
	}
}

func (s *TestProviderSuit) sendWatchEvent(c *C, wc chan clientv3.WatchResponse, event *clientv3.Event, rev int64) {
	resp := clientv3.WatchResponse{}
	if rev != 0 {
		resp.Header = etcdserverpb.ResponseHeader{Revision: rev}
	}
	if event != nil {
		resp.Events = []*clientv3.Event{event}
	}
	select {
	case wc <- resp:
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}
}

func (s *TestProviderSuit) prepareGetCall(kvAPI *mockKvAPI, opts ...callOption) *mock.Call {
	cfg := &callConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.opCount == 0 {
		cfg.opCount = 2
	}
	args := []interface{}{
		mock.Anything, mock.AnythingOfType("string"),
	}
	for i := 0; i < cfg.opCount; i++ {
		args = append(args, mock.AnythingOfType("clientv3.OpOption"))
	}
	return kvAPI.On("Get", args...)

}

func (s *TestProviderSuit) setGet(kvAPI *mockKvAPI, kvs []kv, opts ...callOption) *mockKvAPI {
	cfg := &callConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	s.prepareGetCall(kvAPI, opts...).Return(s.getResponse(kvs, cfg.rev), nil)
	return kvAPI
}

func (s *TestProviderSuit) setGetError(kvAPI *mockKvAPI, err error, opts ...callOption) *mockKvAPI {
	s.prepareGetCall(kvAPI, opts...).Return(nil, err)
	return kvAPI
}

func (s *TestProviderSuit) prepareWatchCall(w *mockWatcher, opts ...callOption) *mock.Call {
	cfg := &callConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.opCount == 0 {
		// 4 options by default
		cfg.opCount = 4
	}
	args := []interface{}{
		mock.Anything, mock.AnythingOfType("string"),
	}
	for i := 0; i < cfg.opCount; i++ {
		args = append(args, mock.AnythingOfType("clientv3.OpOption"))
	}
	return w.On("Watch", args...)
}

// setWatch sets up basic watch, closing the result channel on context cancel and auto-setting revision header if not set
func (s *TestProviderSuit) setWatch(w *mockWatcher, in chan clientv3.WatchResponse, opts ...callOption) *mock.Call {
	out := make(chan clientv3.WatchResponse)
	return s.prepareWatchCall(w, opts...).
		Run(func(args mock.Arguments) {
			go func() {
				ctx := args.Get(0).(context.Context)
				rev := int64(0)
				for {
					rev++
					select {
					case resp := <-in:
						if resp.Header.Revision == 0 {
							resp.Header.Revision = rev
						}
						out <- resp
					case <-ctx.Done():
						// there is no way to set ctx error inside the response
						out <- clientv3.WatchResponse{
							Canceled: true,
							Header: etcdserverpb.ResponseHeader{
								Revision: rev,
							},
						}
						close(out)
					}
				}
			}()
		}).
		Return(clientv3.WatchChan(out))
}

func (s *TestProviderSuit) TestGet_ValidData(c *C) {
	sk := s.newServiceKey()
	kvAPI := s.setGet(&mockKvAPI{}, []kv{{sk.storageKey(), "foo"}})
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	res, err := p.Get(context.Background(), provider.KeyFilter{})
	c.Assert(err, IsNil)
	c.Assert(len(res), Equals, 1)
	expected := provider.KV{
		Namespace: sk.Namespace,
		Service:   sk.Service,
		RawKey:    sk.storageKey(),
		Value:     "foo",
	}
	c.Assert(res[0], Equals, expected)
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestGet_InvalidData(c *C) {
	kvAPI := s.setGet(&mockKvAPI{}, []kv{{"foo", "bar"}})
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	_, err := p.Get(context.Background(), provider.KeyFilter{})
	c.Assert(err, NotNil)
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestGet_Error(c *C) {
	expectedErr := fmt.Errorf("error")
	kvAPI := s.setGetError(&mockKvAPI{}, expectedErr)
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	_, err := p.Get(context.Background(), provider.KeyFilter{})
	c.Assert(err, NotNil)
	c.Assert(err, Equals, expectedErr)
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestGet_DefaultTimeout(c *C) {
	c.Logf("Default timeout should fire if no timeout is set in context")
	tmp := defaultGetTimeout
	defaultGetTimeout = time.Microsecond
	defer func() { defaultGetTimeout = tmp }()

	kvAPI := &mockKvAPI{}
	s.prepareGetCall(kvAPI).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			<-ctx.Done()
		}).
		Return(nil, context.DeadlineExceeded)
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	done := make(chan bool)
	go func() {
		defer close(done)
		_, err := p.Get(context.Background(), provider.KeyFilter{})
		c.Assert(err, NotNil)
		c.Assert(err, Equals, context.DeadlineExceeded)
	}()

	select {
	case <-done:
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestGet_CtxTimeout(c *C) {
	c.Logf("Parent ctx timeout should fire even if it's longer that default")
	tmp := defaultGetTimeout
	defaultGetTimeout = time.Nanosecond
	defer func() { defaultGetTimeout = tmp }()

	kvAPI := &mockKvAPI{}
	s.prepareGetCall(kvAPI).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			<-ctx.Done()
		}).
		Return(nil, context.DeadlineExceeded)
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	done := make(chan time.Duration)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		start := time.Now()
		_, err := p.Get(ctx, provider.KeyFilter{})
		elapsed := time.Now().Sub(start)
		c.Assert(err, NotNil)
		c.Assert(err, Equals, context.DeadlineExceeded)
		done <- elapsed
		close(done)
	}()

	select {
	case elapsed := <-done:
		if elapsed < time.Millisecond {
			c.Fatalf("Context is closed early, in %s", elapsed)
		}
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestFetchExisting_WithData(c *C) {
	host1 := s.newInstanceKey("h1")
	host2 := s.newInstanceKey("h2")
	hosts := []serviceKey{host1, host2}
	kvAPI := s.setGet(&mockKvAPI{}, []kv{{host1.storageKey(), "h1"}, {host2.storageKey(), "h2"}}, withRev(666), withOpCount(1))
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	events := make(chan *provider.Event, 2)
	done := make(chan bool)
	go func() {
		defer close(done)
		rev := p.fetchExisting(context.Background(), "/", events)
		close(events)
		c.Assert(rev, Equals, int64(666))
	}()
	select {
	case <-done:
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}

	c.Assert(len(events), Equals, 1)
	event := <-events
	c.Assert(event.Type, Equals, provider.EventPut)
	c.Assert(len(event.KVs), Equals, len(hosts))
	for i, sk := range hosts {
		expected := provider.KV{
			Namespace: sk.Namespace,
			Service:   sk.Service,
			RawKey:    sk.storageKey(),
			Value:     string(sk.Service.InstanceName),
		}
		c.Assert(event.KVs[i], Equals, expected)
	}

	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestFetchExisting_NoData(c *C) {
	c.Logf("If no data is available - no events will occur, revision is returned")
	kvAPI := s.setGet(&mockKvAPI{}, []kv{}, withRev(666), withOpCount(1))
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	events := make(chan *provider.Event, 2)
	done := make(chan bool)
	go func() {
		defer close(done)
		rev := p.fetchExisting(context.Background(), "/", events)
		close(events)
		c.Assert(rev, Equals, int64(666))
	}()
	select {
	case <-done:
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}

	c.Assert(len(events), Equals, 0)
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestFetchExisting_ErrorData(c *C) {
	c.Logf("If data is unparsable, fetch should pass anyway")
	kvAPI := s.setGet(&mockKvAPI{}, []kv{{"foo", "bar"}}, withRev(666), withOpCount(1))
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	events := make(chan *provider.Event, 2)
	done := make(chan bool)
	go func() {
		defer close(done)
		rev := p.fetchExisting(context.Background(), "/", events)
		close(events)
		c.Assert(rev, Equals, int64(666))
	}()
	select {
	case <-done:
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}

	c.Assert(len(events), Equals, 0)
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestFetchExisting_Timeout(c *C) {
	c.Logf("If error occured, retry should be done after timeout")
	tmp := watchRetryAfter
	watchRetryAfter = time.Millisecond
	defer func() { watchRetryAfter = tmp }()

	called := make(chan bool)
	kvAPI := &mockKvAPI{}
	s.prepareGetCall(kvAPI, withOpCount(1)).
		Run(func(args mock.Arguments) {
			close(called)
		}).
		Return(nil, fmt.Errorf("error")).
		Once()
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	events := make(chan *provider.Event, 2)
	done := make(chan time.Duration)
	go func() {
		start := time.Now()
		rev := p.fetchExisting(context.Background(), "/", events)
		elapsed := time.Now().Sub(start)
		close(events)
		c.Assert(rev, Equals, int64(666))
		done <- elapsed
		close(done)
	}()

	sk := s.newServiceKey()
	select {
	case <-called:
		s.setGet(kvAPI, []kv{{sk.storageKey(), "foo"}}, withRev(666), withOpCount(1))
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}

	select {
	case elapsed := <-done:
		if elapsed < watchRetryAfter {
			c.Fatalf("fetch is done too early, in %s", elapsed)
		}
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}

	expected := provider.KV{
		Namespace: sk.Namespace,
		Service:   sk.Service,
		RawKey:    sk.storageKey(),
		Value:     "foo",
	}
	c.Assert(len(events), Equals, 1)
	event := <-events
	c.Assert(event.Type, Equals, provider.EventPut)
	c.Assert(len(event.KVs), Equals, 1)
	c.Assert(event.KVs[0], Equals, expected)
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestFetchExisting_Cancel(c *C) {
	c.Logf("If error occured and context is canceled, fetch must exit")
	kvAPI := &mockKvAPI{}
	s.prepareGetCall(kvAPI, withOpCount(1)).
		Return(nil, fmt.Errorf("error")).
		Once()
	p := &providerEtcd{
		logger: discovery.NewNilLogger(),
		kvAPI:  kvAPI,
	}

	events := make(chan *provider.Event, 2)
	done := make(chan time.Duration)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		rev := p.fetchExisting(ctx, "/", events)
		close(events)
		c.Assert(rev, Equals, int64(0))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}

	c.Assert(len(events), Equals, 0)
	kvAPI.AssertExpectations(c)
}

func (s *TestProviderSuit) TestDoWatch(c *C) {
	w := &mockWatcher{}
	watchEvents := make(chan clientv3.WatchResponse)
	s.setWatch(w, watchEvents)
	p := &providerEtcd{
		logger:  discovery.NewNilLogger(),
		watcher: w,
	}

	events := make(chan *provider.Event, 5)
	done := make(chan bool)
	go func() {
		expectedRevision := int64(1)
		rev, err := p.doWatch(context.Background(), "/", 0, events)
		close(events)
		c.Assert(rev, Equals, expectedRevision)
		c.Assert(err, NotNil)
		close(done)
	}()

	sk := s.newServiceKey()
	s.sendWatchEvent(c, watchEvents, s.watchEvent(mvccpb.PUT, sk.storageKey(), "foo"), 0)

	expected := provider.KV{
		Namespace: sk.Namespace,
		Service:   sk.Service,
		RawKey:    sk.storageKey(),
		Value:     "foo",
	}
	select {
	case event := <-events:
		c.Assert(event.Type, Equals, provider.EventPut)
		c.Assert(len(event.KVs), Equals, 1)
		c.Assert(event.KVs[0], Equals, expected)
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}

	watchEvents <- clientv3.WatchResponse{Canceled: true}
	close(watchEvents)

	select {
	case <-done:
	case <-time.After(chanTimeout):
		c.Fatalf("Test timeouted!")
	}
	w.AssertExpectations(c)
}

func (s *TestProviderSuit) TestDoWatch_ProgressNotify(c *C) {
	tmp := progressReportTimeout
	progressReportTimeout = 50 * time.Millisecond
	defer func() { progressReportTimeout = tmp }()

	w := &mockWatcher{}
	watchEvents := make(chan clientv3.WatchResponse)
	s.setWatch(w, watchEvents)
	p := &providerEtcd{
		logger:  discovery.NewNilLogger(),
		watcher: w,
	}

	events := make(chan *provider.Event, 5)
	done := make(chan bool)
	go func() {
		expectedRevision := int64(100)
		rev, err := p.doWatch(context.Background(), "/", 0, events)
		close(events)
		c.Assert(rev, Equals, expectedRevision)
		c.Assert(err, NotNil)
		c.Assert(err, ErrorMatches, "ProgressNotify timeout")
		close(done)
	}()

	for i := 1; i <= 100; i++ {
		s.sendWatchEvent(c, watchEvents, nil, int64(i))
		<-time.After(500 * time.Microsecond)
	}
	c.Assert(len(events), Equals, 0)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		c.Fatalf("Test timeouted!")
	}
	w.AssertExpectations(c)
}

func (s *TestProviderSuit) TestDoWatch_CreatedNotify(c *C) {
	tmp := watchCreateTimeout
	watchCreateTimeout = 50 * time.Millisecond
	defer func() { watchCreateTimeout = tmp }()

	w := &mockWatcher{}
	watchEvents := make(chan clientv3.WatchResponse)
	s.setWatch(w, watchEvents)
	p := &providerEtcd{
		logger:  discovery.NewNilLogger(),
		watcher: w,
	}

	events := make(chan *provider.Event, 5)
	done := make(chan bool)
	go func() {
		defer close(done)
		rev, err := p.doWatch(context.Background(), "/", 0, events)
		c.Assert(rev, Equals, int64(0))
		c.Assert(err, NotNil)
		c.Assert(err, ErrorMatches, "CreatedNotify timeout")
	}()

	c.Assert(len(events), Equals, 0)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		c.Fatalf("Test timeouted!")
	}
	w.AssertExpectations(c)
}

func (s *TestProviderSuit) TestDoWatch_FrosenWatch(c *C) {
	tmp := watchCreateTimeout
	watchCreateTimeout = 50 * time.Millisecond
	defer func() { watchCreateTimeout = tmp }()

	finish := make(chan bool)
	w := &mockWatcher{}
	s.prepareWatchCall(w).
		Run(func(args mock.Arguments) {
			// Slow watch creation
			<-finish
		}).
		Return(nil)

	p := &providerEtcd{
		logger:  discovery.NewNilLogger(),
		watcher: w,
	}

	events := make(chan *provider.Event, 5)
	done := make(chan bool)
	go func() {
		defer close(done)
		rev, err := p.doWatch(context.Background(), "/", 0, events)
		c.Assert(rev, Equals, int64(0))
		c.Assert(err, NotNil)
		c.Assert(err, ErrorMatches, "failed to initialize Watch client")
	}()

	c.Assert(len(events), Equals, 0)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		c.Fatalf("Test timeouted!")
	}
	close(finish)
	w.AssertExpectations(c)
}
