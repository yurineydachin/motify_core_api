package balancer

import (
	"sync"

	"godep.lzd.co/discovery"
	"godep.lzd.co/discovery/provider"

	etcdcl "github.com/coreos/etcd/client"
)

// IBalancerFabric implements GetBalancerForService method
type IBalancerFabric interface {
	GetBalancerForService(name string) ILoadBalancer
}

type balancerFabric struct {
	p      provider.IProvider
	client etcdcl.Client
	logger discovery.ILogger

	// etcd2 namespace is removed, because it's one for all services
	venture     string
	environment string
}

var _ IBalancerFabric = (*balancerFabric)(nil)

// NewRoundRobinBalancerFabric creates balancer fabric of round robin balancers
//
// TODO: please, check this stuff is still usefull. If not - mark as deprecated.
// IMHO, NewFallbackBalancerEtcd2 helper should be used directly.
func NewRoundRobinBalancerFabric(p provider.IProvider, client etcdcl.Client, logger discovery.ILogger,
	venture, environment string) IBalancerFabric {

	return &balancerFabric{
		p:           p,
		client:      client,
		logger:      logger,
		venture:     venture,
		environment: environment,
	}
}

// GetBalancerForService returns ILoadBalancer for service by name
func (this *balancerFabric) GetBalancerForService(name string) ILoadBalancer {
	opts := FallbackBalancerEtcd2Options{
		ServiceName: name,
		Venture:     this.venture,
		Environment: this.environment,
	}
	return NewFallbackBalancerEtcd2(this.p, this.client, this.logger, opts)
}

// IBalancerRegestry implements GetServiceAddrByName
type IBalancerRegestry interface {
	GetServiceAddrByName(name string) (string, error)
}

type balancerRegestry struct {
	venture     string
	environment string
	logger      discovery.ILogger
	client      etcdcl.Client
	p           provider.IProvider
	balancers   map[string]ILoadBalancer
	m           sync.RWMutex
}

var _ IBalancerRegestry = (*balancerRegestry)(nil)

// NewRoundRobinBalancerRegestry creates round robin balancer
//
// TODO: please, check this stuff is still usefull. If not - mark as deprecated.
// IMHO, NewFallbackBalancerEtcd2 helper should be used directly.
func NewRoundRobinBalancerRegestry(p provider.IProvider, client etcdcl.Client, logger discovery.ILogger,
	venture, environment string) IBalancerRegestry {

	return &balancerRegestry{
		venture:     venture,
		environment: environment,
		logger:      logger,
		client:      client,
		p:           p,
		balancers:   make(map[string]ILoadBalancer, 1),
	}
}

// GetServiceAddrByName returns service addr by name
func (this *balancerRegestry) GetServiceAddrByName(name string) (string, error) {
	this.m.RLock()
	balancer, exist := this.balancers[name]
	this.m.RUnlock()
	if !exist {
		balancer = this.create(name)
		this.m.Lock()
		this.balancers[name] = balancer
		this.m.Unlock()
	}
	return balancer.Next()
}

func (this *balancerRegestry) create(name string) ILoadBalancer {
	opts := FallbackBalancerEtcd2Options{
		ServiceName: name,
		Venture:     this.venture,
		Environment: this.environment,
	}
	return NewFallbackBalancerEtcd2(this.p, this.client, this.logger, opts)
}
