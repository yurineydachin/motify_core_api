package balancer

// IRolloutBalancer is an interface for rollout-based load balancing
type IRolloutBalancer interface {
	Next(opts NextOptions) (string, error)
	Stats(opts StatsOptions) []NodeStat
	Stop()
}

// ILoadBalancer is an interface for load balancing
type ILoadBalancer interface {
	Next() (string, error)
	// ServiceName returns service name for which the balancer is set up. Used for some gRPC stuff.
	ServiceName() string
	Stats() []NodeStat
	Stop()
}

// NextOptions contains options for IRolloutBalancer.Next()
type NextOptions struct {
	// SegregationID forces balancer to map given string value to corresponding RolloutType.
	// If any server of given RolloutType is available - it's used in balancing.
	// If no server is available or in case of any error - stable server addresses are used.
	//
	// https://confluence.lzd.co/pages/viewpage.action?spaceKey=DEV&title=Progressive+services+rollout
	SegregationID string
}

// StatsOptions contains options for IRolloutBalancer.Stats()
type StatsOptions struct {
	// RolloutType - if specified, forces returning []NodeStat only for given RolloutType service nodes
	RolloutType string
}

type inMemBalancer struct {
	targets    []string
	currentIdx int
}

func newInMemBalancer(targets []string) *inMemBalancer {
	return &inMemBalancer{
		targets:    targets,
		currentIdx: -1,
	}
}

func (b *inMemBalancer) nextIndex() (int, error) {
	count := len(b.targets)
	if count == 0 {
		return 0, newErrNoServiceAvailable("inmem")
	}
	b.currentIdx = (b.currentIdx + 1) % count
	return b.currentIdx, nil
}

func (b *inMemBalancer) Next() (string, error) {
	idx, err := b.nextIndex()
	if err != nil {
		return "", err
	}
	return b.targets[idx], nil
}

func (b *inMemBalancer) Stats() []NodeStat {
	res := make([]NodeStat, 0, len(b.targets))
	for _, target := range b.targets {
		res = append(res, NodeStat{Value: target})
	}
	return res
}

func (b *inMemBalancer) Stop() {}

func (b *inMemBalancer) ServiceName() string { return "" }
