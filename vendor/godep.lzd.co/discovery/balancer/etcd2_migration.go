package balancer

// This file contains etcd2 -> etcd3 migration helpers.
// TODO: purge after dropping etcd2 support

import (
	etcdcl "github.com/coreos/etcd/client"
	"godep.lzd.co/discovery"
	"godep.lzd.co/discovery/locator"
	"godep.lzd.co/discovery/provider"
)

const (
	defaultNamespace = "lazada_api"
)

// FallbackBalancerEtcd2Options contains config for fallback balancer constructor
type FallbackBalancerEtcd2Options struct {
	// ServiceName is application name string to watch the discovery updates.
	ServiceName string

	// BalancerType describes balancer type to be used.
	// If not specified, TypeRoundRobin is used by default
	BalancerType balancerType

	// EndpointType describes which endpoint updates to listen to.
	// If not specified, TypeAppMain is used by default.
	EndpointType locator.EndpointType

	// Etcd2-specific fields
	Venture     string
	Environment string
	// Namespace is etcd2 key namespace
	// If not specified, "lazada_api" is used by default
	Namespace string
}

// defaultize sets default values or panics if some mandatory params not set
func (opts *FallbackBalancerEtcd2Options) defaultize() {
	if opts.ServiceName == "" {
		panic("empty ServiceName")
	}
	if opts.Venture == "" {
		panic("empty Venture")
	}
	if opts.Environment == "" {
		panic("empty Environment")
	}

	if opts.EndpointType == locator.TypeUnknown {
		opts.EndpointType = locator.TypeAppMain
	}
	if opts.Namespace == "" {
		opts.Namespace = defaultNamespace
	}
}

func newV3Balancer(l locator.ILocator, logger discovery.ILogger, opts FallbackBalancerEtcd2Options) ILoadBalancer {
	lbOpts := LoadBalancerOptions{ServiceName: opts.ServiceName, EndpointType: opts.EndpointType}
	switch opts.BalancerType {
	case TypeRoundRobin:
		return NewRoundRobin(l, logger, lbOpts)
	case TypeWeightedRoundRobin:
		return NewWeightedRoundRobin(l, logger, lbOpts)
	default:
		panic("unknown balancer type")
	}
}

func newV2Balancer(client etcdcl.Client, logger discovery.ILogger, opts FallbackBalancerEtcd2Options) ILoadBalancer {
	l := locator.NewLocatorEtcd2(client, logger)
	info := locator.LocationInfo{
		Namespace:   opts.Namespace,
		Venture:     opts.Venture,
		Environment: opts.Environment,
		ServiceName: opts.ServiceName,
	}
	switch opts.BalancerType {
	case TypeRoundRobin:
		return NewRoundRobinEtcd2(l, logger, info)
	case TypeWeightedRoundRobin:
		return NewWeightedRoundRobinEtcd2(l, logger, info)
	default:
		panic("unknown balancer type")
	}
}

// NewFallbackBalancerEtcd2 returns new ILoadBalancer supporting fallback from etcd3 to etcd2 service discovery
func NewFallbackBalancerEtcd2(p provider.IProvider, client etcdcl.Client, logger discovery.ILogger, opts FallbackBalancerEtcd2Options) ILoadBalancer {
	opts.defaultize()

	bV3 := newV3Balancer(locator.New(p, logger), logger, opts)
	bV2 := newV2Balancer(client, logger, opts)

	return NewFallbackBalancer(logger, bV3, bV2)
}
