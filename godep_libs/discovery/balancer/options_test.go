package balancer

import (
	"testing"

	"motify_core_api/godep_libs/discovery/locator"
	"motify_core_api/godep_libs/discovery/provider"
)

func checkDefaultOptions(opts LoadBalancerOptions, t *testing.T) {
	if opts.EndpointType != locator.TypeAppMain {
		t.Fatalf("invalid EndpointType: %s, expected: %s", opts.EndpointType, locator.TypeAppMain)
	}

	expected := locator.KeyFilter{
		RolloutType: "stable",
		Owner:       provider.DefaultOwner,
		ClusterType: provider.DefaultClusterType,
	}
	if opts.Filter == nil {
		t.Fatalf("Filter is nil")
	}
	if *opts.Filter != expected {
		t.Fatalf("invalid filter: %#v; expected: %#v", *opts.Filter, expected)
	}
}

func TestLoadBalancerOptions_Defaultize(t *testing.T) {
	opts := LoadBalancerOptions{
		ServiceName: "test",
	}
	opts.defaultize()
	checkDefaultOptions(opts, t)
}

func TestLoadBalancerOptions_Defaultize_EmptyServiceName(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("panic is expected if ServiceName is not provided")
		}
	}()
	opts := LoadBalancerOptions{}
	opts.defaultize()
}

func TestRolloutBalancerOptions_Defaultize(t *testing.T) {
	opts := RolloutBalancerOptions{
		LoadBalancerOptions: LoadBalancerOptions{
			ServiceName: "test",
		},
	}
	opts.defaultize()

	t.Logf("embedded LoadBalancerOptions should be defaultized")
	checkDefaultOptions(opts.LoadBalancerOptions, t)
	if opts.BalancerType != TypeRoundRobin {
		t.Fatalf("BalancerType is not default")
	}
	if opts.FallbackBalancer != nil {
		t.Fatalf("FallbackBalancer is not default")
	}
}
