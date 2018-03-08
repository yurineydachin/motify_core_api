package locator

import (
	"godep.lzd.co/discovery/provider"
)

// KeyFilter contains service fields, clarifying location search
//
// It contains subset of provider.KeyFilter fields, because we don't need
// defining namespace, service name and type via filter.
type KeyFilter struct {
	// RolloutType - if set, clarifies the rollout type (e.g. "stable").
	RolloutType string
	// Owner - if set, clarifies the service owner (e.g. "shared")
	Owner string
	// ClusterType - if set, clarifies the cluster type (e.g. "common")
	ClusterType string
}

func discoveryKeyFilter(serviceName string, t EndpointType, filter *KeyFilter) provider.KeyFilter {
	if filter == nil {
		filter = &KeyFilter{}
	}
	f := provider.KeyFilter{
		Namespace:   provider.NamespaceDiscovery,
		Type:        t.ServiceType(),
		Name:        serviceName,
		RolloutType: filter.RolloutType,
		Owner:       filter.Owner,
		ClusterType: filter.ClusterType,
	}
	return f
}
