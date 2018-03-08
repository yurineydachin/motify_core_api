package balancer

import (
	"godep.lzd.co/discovery/locator"
	"godep.lzd.co/discovery/provider"
)

type balancerType uint8

// Load balancer types
const (
	TypeRoundRobin balancerType = iota
	TypeWeightedRoundRobin
)

// LoadBalancerOptions contains config for load balancer constructors
type LoadBalancerOptions struct {
	// ServiceName is application name string to watch the discovery updates.
	ServiceName string

	// EndpointType describes which endpoint updates to listen to.
	// Unfortunately, there is no way for guessing it up - you may want to listen either for "endpoint_main"
	// (where HTTP is set in most cases) or "endpoint_additional" (where gRPC is set if service supports it).
	// If not specified - TypeAppMain is used.
	EndpointType locator.EndpointType

	// Filter is an optional discovery key filtering params to watch updates of specific service instances.
	// Pass "nil" to set proper default values, covering all common balancing cases.
	Filter *locator.KeyFilter
}

// defaultize sets default values or panics if some mandatory params not set
func (opts *LoadBalancerOptions) defaultize() {
	if opts.ServiceName == "" {
		panic("empty ServiceName")
	}
	if opts.EndpointType == locator.TypeUnknown {
		opts.EndpointType = locator.TypeAppMain
	}
	if opts.Filter == nil {
		// if no filters provided - use default ones, covering all common use cases
		// for applications
		opts.Filter = &locator.KeyFilter{
			RolloutType: provider.RolloutTypeStable,
			Owner:       provider.DefaultOwner,
			ClusterType: provider.DefaultClusterType,
		}
	}
}

// RolloutBalancerOptions contains config for rollout balancer constructor
type RolloutBalancerOptions struct {
	LoadBalancerOptions

	// BalancerType is the balancer type to be used.
	BalancerType balancerType

	// FallbackBalancer is an optional balancer to be used as a fallback balancer
	// for "stable" RolloutType.
	// Should be used if you want to have a fallback to etcd2 balancers along with new Rollout balancing feature.
	// TODO: delete in next major version along with removing etcd2 support.
	FallbackBalancer ILoadBalancer
}

// defaultize sets default values or panics if some mandatory params not set
func (opts *RolloutBalancerOptions) defaultize() {
	opts.LoadBalancerOptions.defaultize()
}
