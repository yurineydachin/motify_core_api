package locator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"motify_core_api/godep_libs/discovery/provider"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TestLocatorSuit struct{}

var _ = Suite(&TestLocatorSuit{})

func (s *TestLocatorSuit) TestGet(c *C) {
	c.Logf("locator.Get should return parsed endpoint value")

	p := &provider.Mock{
		GetCallback: func(ctx context.Context, filter provider.KeyFilter) ([]provider.KV, error) {
			return []provider.KV{
				provider.KV{
					Service: provider.Service{
						Name:        "test",
						Type:        provider.TypeSystem,
						RolloutType: filter.RolloutType,
						Owner:       filter.Owner,
						ClusterType: filter.ClusterType,
					},
					Value: `{"endpoint_main": "go1.dc:8080"}`,
				},
			}, nil
		},
	}
	locator := New(p, nil)
	expected := []Location{
		Location{
			Service: provider.Service{
				Name: "test",
				Type: provider.TypeSystem,
			},
			Endpoint: "go1.dc:8080",
		},
	}

	locations, err := locator.Get(context.Background(), "test", TypeSystem, nil)
	c.Assert(err, IsNil)
	c.Assert(locations, DeepEquals, expected)
}

func (s *TestLocatorSuit) TestGet_InvalidValue(c *C) {
	c.Logf("locator.Get should fail if value can't be parsed")

	p := &provider.Mock{
		GetCallback: func(ctx context.Context, filter provider.KeyFilter) ([]provider.KV, error) {
			return []provider.KV{
				provider.KV{
					Value: "go1.dc:8080",
				},
			}, nil
		},
	}
	locator := New(p, nil)

	locations, err := locator.Get(context.Background(), "test", TypeUnknown, nil)
	c.Assert(err, NotNil)
	c.Assert(locations, IsNil)
}

func (s *TestLocatorSuit) TestGet_Error(c *C) {
	c.Logf("locator.Get should fail if provider returns error")

	p := &provider.Mock{
		GetCallback: func(ctx context.Context, filter provider.KeyFilter) ([]provider.KV, error) {
			return nil, fmt.Errorf("error")
		},
	}
	locator := New(p, nil)

	locations, err := locator.Get(context.Background(), "test", TypeUnknown, nil)
	c.Assert(err, NotNil)
	c.Assert(locations, IsNil)
}

func (s *TestLocatorSuit) TestWatch(c *C) {
	c.Logf("locator.Watch should process provider.Watch events")

	discoveryEvents := make(chan *provider.Event, 1)
	p := &provider.Mock{
		WatchCallback: func(ctx context.Context, filter provider.KeyFilter) <-chan *provider.Event {
			return discoveryEvents
		},
	}
	locator := New(p, nil)

	events := locator.Watch(context.Background(), "test", TypeAppMain, nil)
	c.Assert(events, NotNil)

	discoveryEvents <- &provider.Event{
		Type: provider.EventDelete,
		KVs:  []provider.KV{provider.KV{Service: provider.Service{Name: "test"}}},
	}

	expected := &Event{
		Type: provider.EventDelete,
		Locations: []Location{
			Location{
				Service: provider.Service{Name: "test"},
			},
		},
	}

	select {
	case <-time.After(time.Second):
		c.Fatalf("watch timeout")
	case event := <-events:
		c.Assert(event, DeepEquals, expected)
	}
}

func (s *TestLocatorSuit) TestWatchCancel(c *C) {
	discoveryEvents := make(chan *provider.Event)
	p := &provider.Mock{
		WatchCallback: func(ctx context.Context, filter provider.KeyFilter) <-chan *provider.Event {
			return discoveryEvents
		},
	}
	locator := New(p, nil)

	ctx, cancel := context.WithCancel(context.Background())
	events := locator.Watch(ctx, "test", TypeAppMain, nil)
	c.Assert(events, NotNil)

	c.Logf("locator.Watch should be canceled via context and read-out all data from provider")
	cancel()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			discoveryEvents <- &provider.Event{Type: provider.EventDelete}
		}
		close(discoveryEvents)
		close(done)
	}()

	select {
	case <-time.After(time.Second):
		c.Fatalf("discovery events are not readed out from channel")
	case <-done:
	}

	c.Logf("There should be no data in event channel, the channel should be closed.")
	c.Assert(len(events), Equals, 0)
	select {
	case <-time.After(time.Second):
		c.Fatalf("events channel is not closed")
	case <-events:
		// It doesn't matter if the channel is closed or the event is set.
		// Watcher just attempts to send the event and skips if the reciever is not ready.
		// Because we can't check if the channel is closed without reading it, we may eventually
		// recieve event value here.
	}
}
