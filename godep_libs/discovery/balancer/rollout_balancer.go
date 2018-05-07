package balancer

import (
	"fmt"

	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/locator"
	"motify_core_api/godep_libs/discovery/provider"
)

const (
	// unstableRolloutCount is the maximum count of unstable rollouts
	unstableRolloutCount = 20
)

// rolloutBalancer is a balancer, supporting Progressive Services rollout balancing scheme
type rolloutBalancer struct {
	serviceName string
	watcher     IRolloutWatcher
	logger      discovery.ILogger

	stableBalancer    ILoadBalancer
	unstableBalancers map[string]ILoadBalancer
	rolloutKeysSorted []string
}

// NewRolloutBalancer returns new IRolloutBalancer instance with predefined rollout watcher instance.
// Watcher instance should be shared between rollout balancers of different services for less performance footprint.
func NewRolloutBalancer(p provider.IProvider, w IRolloutWatcher, logger discovery.ILogger, opts RolloutBalancerOptions) IRolloutBalancer {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}
	opts.defaultize()

	r := &rolloutBalancer{
		serviceName:       opts.ServiceName,
		watcher:           w,
		logger:            logger,
		unstableBalancers: make(map[string]ILoadBalancer),
	}

	l := locator.New(p, logger)
	if opts.FallbackBalancer == nil {
		// ensure watching "stable" rollout type
		opts.Filter.RolloutType = provider.RolloutTypeStable
		r.stableBalancer = newBalancer(l, logger, opts)
	} else {
		r.stableBalancer = opts.FallbackBalancer
	}

	// set unstable balancers
	r.rolloutKeysSorted = make([]string, 0, unstableRolloutCount)
	for i := 1; i <= unstableRolloutCount; i++ {
		rolloutType := fmt.Sprintf("unstable%d", i)
		r.rolloutKeysSorted = append(r.rolloutKeysSorted, rolloutType)

		unstableOpts := opts
		f := *opts.Filter
		f.RolloutType = rolloutType
		unstableOpts.Filter = &f

		r.unstableBalancers[rolloutType] = newBalancer(l, logger, unstableOpts)
	}

	return r
}

func newBalancer(l locator.ILocator, logger discovery.ILogger, opts RolloutBalancerOptions) ILoadBalancer {
	switch opts.BalancerType {
	case TypeRoundRobin:
		return NewRoundRobin(l, logger, opts.LoadBalancerOptions)
	case TypeWeightedRoundRobin:
		return NewWeightedRoundRobin(l, logger, opts.LoadBalancerOptions)
	default:
		panic("unknown balancer type")
	}
}

// Next returns next balanced service address for given NextOptions
func (r *rolloutBalancer) Next(opts NextOptions) (string, error) {
	rolloutType, err := r.watcher.GetRolloutType(opts.SegregationID)
	if err != nil {
		// There can be no SegregationID header at all or wrong format.
		// When Infra team makes this header mandatory (INFRASYS-3138) it makes sense to increase log level to Warning
		r.logger.Infof("%q GetRolloutType() error: %q, using stable balancer", r.serviceName, err)
		return r.stableBalancer.Next()
	}
	if rolloutType == provider.RolloutTypeStable {
		return r.stableBalancer.Next()
	}

	balancer := r.unstableBalancers[rolloutType]
	if balancer == nil {
		// almost unreal, because the balancer are initialized in constructor, but better to log here.
		r.logger.Errorf("%q no balancer for rollout type %q, using stable", r.serviceName, rolloutType)
		return r.stableBalancer.Next()
	}

	unstableAdd, err := balancer.Next()
	if err != nil {
		r.logger.Infof("%q error from %q balancer, using stable: %s", r.serviceName, rolloutType, err)
		return r.stableBalancer.Next()
	}

	return unstableAdd, nil
}

// Stats returns nodes statistics for given StatsOptions
func (r *rolloutBalancer) Stats(opts StatsOptions) []NodeStat {
	if opts.RolloutType != "" {
		return r.statsByRolloutType(opts.RolloutType)
	}

	stats := []NodeStat{}
	stats = append(stats, r.stableBalancer.Stats()...)
	for _, t := range r.rolloutKeysSorted {
		b := r.unstableBalancers[t]
		stats = append(stats, b.Stats()...)
	}
	return stats
}

// Stop stops watching for node address updates
func (r *rolloutBalancer) Stop() {
	r.stableBalancer.Stop()
	for _, b := range r.unstableBalancers {
		b.Stop()
	}
}

// statsByRolloutType returns stats of corresponding rollout balancer
func (r *rolloutBalancer) statsByRolloutType(rolloutType string) []NodeStat {
	var b ILoadBalancer
	if rolloutType == provider.RolloutTypeStable {
		b = r.stableBalancer
	} else {
		b = r.unstableBalancers[rolloutType]
	}
	if b == nil {
		return nil
	}

	stats := b.Stats()
	for i := range stats {
		stats[i].RolloutType = rolloutType
	}
	return stats
}
