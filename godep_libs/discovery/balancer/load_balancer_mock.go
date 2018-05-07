package balancer

// LoadBalancerMock is a struct for mocking ILoadBalancer
type LoadBalancerMock struct {
	NextCallback        func() (string, error)
	StatsCallback       func() []NodeStat
	ServiceNameCallback func() string
	StopCallback        func()
}

var _ ILoadBalancer = &LoadBalancerMock{}

// Next calls NextCallback
func (b *LoadBalancerMock) Next() (string, error) {
	return b.NextCallback()
}

// ServiceName calls ServiceNameCallback
func (b *LoadBalancerMock) ServiceName() string {
	if b.ServiceNameCallback != nil {
		return b.ServiceNameCallback()
	}
	return ""
}

// Stats calls StatsCallback
func (b *LoadBalancerMock) Stats() []NodeStat {
	return b.StatsCallback()
}

// Stop calls StopCallback
func (b *LoadBalancerMock) Stop() {
	b.StatsCallback()
}
