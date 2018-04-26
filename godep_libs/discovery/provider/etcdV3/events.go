package etcdV3

import (
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"motify_core_api/godep_libs/discovery/provider"
)

var eventTypes = map[mvccpb.Event_EventType]provider.EventType{
	mvccpb.PUT:    provider.EventPut,
	mvccpb.DELETE: provider.EventDelete,
}

var validNamespaces = map[string]bool{
	provider.NamespaceRollout:          true,
	provider.NamespaceExportedEntities: true,
}

func discoveryEventType(t mvccpb.Event_EventType) provider.EventType {
	return eventTypes[t]
}

func extractNamespace(key string) (string, bool) {
	parts := strings.SplitN(key, "/", 3)
	if len(parts) <= 1 {
		return "", false
	}
	// key with namespace should start with "/"
	if parts[0] != "" {
		return "", false
	}
	ns := parts[1]
	if validNamespaces[ns] {
		return ns, true
	}
	return "", false
}

func discoveryKV(kv *mvccpb.KeyValue) (provider.KV, error) {
	rawKey := string(kv.Key)
	key, err := parseServiceKey(rawKey)
	if err != nil {
		// We have some namespaces where keys are not service-specific, but we still want
		// to work with them via provider.
		// And I don't want to hide the parse error in other cases.
		// So we will check this namespaces manually.
		ns, ok := extractNamespace(rawKey)
		if !ok {
			return provider.KV{}, err
		}
		return provider.KV{
			Namespace: ns,
			RawKey:    rawKey,
			Value:     string(kv.Value),
		}, nil
	}

	resp := provider.KV{
		Namespace: key.Namespace,
		Service:   key.Service,
		RawKey:    rawKey,
		Value:     string(kv.Value),
	}
	return resp, nil
}

func newEventFromKVs(etcdKVs []*mvccpb.KeyValue) (*provider.Event, error) {
	kvs := make([]provider.KV, 0, len(etcdKVs))
	for _, ectdKV := range etcdKVs {
		kv, err := discoveryKV(ectdKV)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv)
	}
	resp := &provider.Event{
		Type: provider.EventPut,
		KVs:  kvs,
	}
	return resp, nil
}

func watchResponse(event *clientv3.Event) (*provider.Event, error) {
	if event.Kv == nil {
		return nil, fmt.Errorf("empty event.Kv")
	}

	kv, err := discoveryKV(event.Kv)
	if err != nil {
		return nil, err
	}

	resp := &provider.Event{
		Type: discoveryEventType(event.Type),
		KVs:  []provider.KV{kv},
	}
	return resp, nil
}
