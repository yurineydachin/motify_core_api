package balancer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/provider"
)

const (
	// rolloutPrefix is a key prefix for watching rollout updates
	rolloutPrefix = "/" + provider.NamespaceRollout + "/segregation/"
	// maxSegregationID represents maximum count of segregation ID buckets to allocate map length once
	maxSegregationID = 1000
)

// IRolloutWatcher is an interface for watching changes of segregationID -> rolloutType mapping in discovery provider.
type IRolloutWatcher interface {
	// GetRolloutType parses segregationID and returns corresponding rolloutType string
	GetRolloutType(segregationID string) (string, error)
	// Stop stops watch process
	Stop()
}

type rolloutWatcher struct {
	logger discovery.ILogger

	mu           sync.RWMutex
	rolloutTypes map[int64]string
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewRolloutWatcher returns new IRolloutWatcher instance.
func NewRolloutWatcher(p provider.IProvider, logger discovery.ILogger) IRolloutWatcher {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &rolloutWatcher{
		logger:       logger,
		rolloutTypes: make(map[int64]string, maxSegregationID),
		ctx:          ctx,
		cancel:       cancel,
	}

	go r.watch(p)
	return r
}

// GetRolloutType parses segregationID and returns corresponding rolloutType
func (r *rolloutWatcher) GetRolloutType(segregationID string) (string, error) {
	if segregationID == "" {
		return "", fmt.Errorf("empty segregationID")
	}

	i, err := strconv.ParseInt(segregationID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("can't parse segregationID %q: %s", segregationID, err)
	}

	r.mu.RLock()
	t, ok := r.rolloutTypes[i]
	r.mu.RUnlock()
	if !ok {
		return provider.RolloutTypeStable, nil
	}
	return t, nil
}

// Stop stops watching rollout types updates
func (r *rolloutWatcher) Stop() {
	r.cancel()
}

// watch is a consuming loop for rollout mapping events.
// It starts watching for changes in rollot deploy and updates internal map segregationID->rolloutType
func (r *rolloutWatcher) watch(p provider.IProvider) {
	filter := provider.KeyFilter{
		Prefix: rolloutPrefix,
	}
	events := p.Watch(r.ctx, filter)
	for {
		select {
		case event, ok := <-events:
			if !ok {
				r.logger.Errorf("rollout watch channel closed unexpectedly")
				return
			}
			r.handleEvent(event)
		case <-r.ctx.Done():
			r.logger.Debugf("rollout watch is stopped")
			return
		}
	}
}

// handleEvent processes Watch event, adding or deleting segregationID->rolloutType mappings
func (r *rolloutWatcher) handleEvent(event *provider.Event) {
	// Any event leads data access, so makes sense to lock here
	r.mu.Lock()
	switch event.Type {
	case provider.EventPut:
		for _, kv := range event.KVs {
			id, err := parseSegregationID(kv.RawKey)
			if err != nil {
				r.logger.Warningf("%s", err)
				continue
			}
			r.rolloutTypes[id] = kv.Value
		}
	case provider.EventDelete:
		for _, kv := range event.KVs {
			id, err := parseSegregationID(kv.RawKey)
			if err != nil {
				r.logger.Warningf("%s", err)
				continue
			}
			delete(r.rolloutTypes, id)
		}
	}
	r.mu.Unlock()
}

// parseSegregationID parses key string into segregation ID value
func parseSegregationID(k string) (int64, error) {
	if !strings.HasPrefix(k, rolloutPrefix) {
		return 0, fmt.Errorf("can't parse key '%s': rollout prefix not found", k)
	}

	segregationID := strings.TrimPrefix(k, rolloutPrefix)
	i, err := strconv.ParseInt(segregationID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse segregationID in key '%s': %s", k, err)
	}

	return i, nil
}
