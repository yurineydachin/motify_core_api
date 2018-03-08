package balancer

import (
	"fmt"
	"godep.lzd.co/discovery"
)

// FallbackBalancer is a load balancer, which should be used for migration period
// between different discovery protocols and interfaces.
// Every method returns the data of the first available balancer. In case of error
// the next balancer is used and so on. The error is returned only in case when
// all balancers failed.
type FallbackBalancer struct {
	balancers   []ILoadBalancer
	logger      discovery.ILogger
	logName     string
	serviceName string
}

var _ ILoadBalancer = &FallbackBalancer{}

// NewFallbackBalancer returns FallbackBalancer initialized with given set of balancers.
// Balancers priority is high-to-low - the first one has the highest.
func NewFallbackBalancer(logger discovery.ILogger, balancers ...ILoadBalancer) *FallbackBalancer {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}
	if len(balancers) == 0 {
		panic("No balancers provided")
	}

	b := &FallbackBalancer{
		balancers: balancers,
		logger:    logger,
	}

	// cache balancer name and logName, don't want to calculate every time
	for _, balancer := range b.balancers {
		if name := balancer.ServiceName(); name != "" {
			b.serviceName = name
		}
		if stringer, ok := balancer.(fmt.Stringer); ok {
			b.logName = stringer.String()
		}
	}
	if b.logName == "" {
		// fallback
		b.logName = b.ServiceName()
	}

	return b
}

// Next returns the first valid target address from available balancers.
func (b *FallbackBalancer) Next() (string, error) {
	for i, balancer := range b.balancers {
		res, err := balancer.Next()
		if err == nil {
			return res, nil
		}
		b.logger.Debugf("balancer[%d].Next() failed: %s", i, err)
	}
	return "", newErrNoServiceAvailable(b.logName)
}

// ServiceName returns service name for which the balancer is set up.
func (b *FallbackBalancer) ServiceName() string {
	return b.serviceName
}

// Stats returns the first non-empty node stats from available balancers.
func (b *FallbackBalancer) Stats() []NodeStat {
	for i, balancer := range b.balancers {
		res := balancer.Stats()
		if len(res) != 0 {
			return res
		}
		b.logger.Debugf("balancer[%d].Stats() is empty", i)
	}
	return nil
}

// Stop stops all balancers
func (b *FallbackBalancer) Stop() {
	for _, balancer := range b.balancers {
		balancer.Stop()
	}
}
