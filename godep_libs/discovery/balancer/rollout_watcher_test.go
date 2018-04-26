package balancer

import (
	"context"
	"testing"
	"time"

	"godep.lzd.co/discovery/provider"
)

func checkRolloutType(rollout string, err error, expected string, t *testing.T) {
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if rollout != expected {
		t.Fatalf("got %q, expected %q", rollout, expected)
	}
}

// A dirty hack again
func watchProviderReady(p *rolloutProvider, t *testing.T) {
	d := 100 * time.Millisecond
	deadline := time.After(d)
	for {
		p.mu.RLock()
		l := len(p.watches[rolloutPrefix])
		p.mu.RUnlock()
		if l >= 1 {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("no watch called for %s", d)
		case <-time.After(1 * time.Millisecond):
			// poll watches
		}
	}
}

func TestRolloutWatcher_NewRolloutWatcher(t *testing.T) {
	r := NewRolloutWatcher(newRolloutProvider(), nil)

	if r == nil {
		t.Fatalf("watcher is nil")
	}
	if _, ok := r.(*rolloutWatcher); !ok {
		t.Fatalf("unexpected NewRolloutWatcher core type")
	}
	r.Stop()
}

func TestRolloutWatcher_GetRolloutType_Error(t *testing.T) {
	r := NewRolloutWatcher(newRolloutProvider(), nil)
	defer r.Stop()

	_, err := r.GetRolloutType("")
	if err == nil {
		t.Errorf("Error should be returned if segregationID is empty")
	}

	_, err = r.GetRolloutType("ololo")
	if err == nil {
		t.Errorf("Error should be returned if segregationID is unparsable to int")
	}
}

func TestRolloutWatcher_GetRolloutType_NotFound(t *testing.T) {
	r := NewRolloutWatcher(newRolloutProvider(), nil)
	defer r.Stop()

	rolloutType, err := r.GetRolloutType("1")
	checkRolloutType(rolloutType, err, provider.RolloutTypeStable, t)
}

func TestRolloutWatcher_GetRolloutType(t *testing.T) {
	p := newRolloutProvider()
	r := NewRolloutWatcher(p, nil)
	defer r.Stop()
	watchProviderReady(p, t)

	t.Logf("Int format should not make any difference")
	p.setKV(rolloutPrefix+"001", "unstable1")
	p.setKV(rolloutPrefix+"2", "unstable1")
	waitForRolloutsChange(r, 2)

	rolloutType, err := r.GetRolloutType("1")
	checkRolloutType(rolloutType, err, "unstable1", t)

	rolloutType, err = r.GetRolloutType("2")
	checkRolloutType(rolloutType, err, "unstable1", t)
}

func TestRolloutWatcher_Stop(t *testing.T) {
	watchCanceled := make(chan bool)
	p := &provider.Mock{
		WatchCallback: func(ctx context.Context, filter provider.KeyFilter) <-chan *provider.Event {
			go func() {
				<-ctx.Done()
				close(watchCanceled)
			}()
			return make(chan *provider.Event)
		},
	}
	r := NewRolloutWatcher(p, nil)

	r.Stop()
	select {
	case <-watchCanceled:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Stop() did not cancel context")
	}

	t.Logf("Consequent Stop() calls should not panic")
	r.Stop()
}
