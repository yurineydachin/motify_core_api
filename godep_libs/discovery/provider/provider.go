package provider

import (
	"context"
)

// Namespaces used for registering data in discovery provider
const (
	NamespaceAdmin            = "admin"
	NamespaceDiscovery        = "discovery"
	NamespaceMetrics          = "metrics"
	NamespaceRollout          = "rollout"
	NamespaceExportedEntities = "exported_entities"
)

// IProvider is a common interface for all service discovery providers
type IProvider interface {
	// RegisterValues sets the values after validation.
	// Values are automatically keeped alive and deleted after context cancellation.
	RegisterValues(ctx context.Context, kvs ...KV) error

	// Get returns all values matching the given key filter
	Get(ctx context.Context, filter KeyFilter) ([]KV, error)

	// Watch waits for updates on given keys defined via key filter and sends
	// them to events channel.
	// Watch is canceled via context cancellation.
	Watch(ctx context.Context, filter KeyFilter) <-chan *Event
}

// EventType is event type
type EventType uint8

// Event types enum
const (
	EventUnknown EventType = iota
	EventPut
	EventDelete
)

// String returns string representation of type
func (e EventType) String() string {
	switch e {
	case EventPut:
		return "PUT"
	case EventDelete:
		return "DELETE"
	}
	return "UNKNOWN"
}

// KeyFilter contains key-defining filter fields for further search.
// Fields can be extended in future without major verion release.
type KeyFilter struct {
	// Prefix defines the raw key prefix to search
	Prefix string
	// Namespace defines the namespace to search in (e.g. "discovery").
	Namespace string
	// Type - if set, clarifies the service type (e.g. "app").
	Type ServiceType
	// Name - if set, clarifies the service name (e.g. "customer_api").
	Name string
	// RolloutType - if set, clarifies the rollout type (e.g. "stable").
	RolloutType string
	// Owner - if set, clarifies the service owner (e.g. "shared")
	Owner string
	// ClusterType - if set, clarifies the cluster type (e.g. "common")
	ClusterType string
}

// KV contains stored value and service key (Namespace and Service)
type KV struct {
	Namespace string
	Service   Service
	RawKey    string
	Value     string
}

// Event represents Watch event
type Event struct {
	Type EventType
	KVs  []KV
}
