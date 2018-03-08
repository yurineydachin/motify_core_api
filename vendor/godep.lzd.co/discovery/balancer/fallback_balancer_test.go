package balancer

import (
	"fmt"
	"testing"
)

func TestFallbackBalancer_NextFirst(t *testing.T) {
	balancers := []ILoadBalancer{
		newInMemBalancer([]string{"1", "2", "3"}),
		newInMemBalancer([]string{"4", "5", "6"}),
	}
	b := NewFallbackBalancer(nil, balancers...)

	t.Logf("First available balancer should generate targets")
	if target, _ := b.Next(); target != "1" {
		t.Fail()
	}
	if target, _ := b.Next(); target != "2" {
		t.Fail()
	}
	if target, _ := b.Next(); target != "3" {
		t.Fail()
	}
	if target, _ := b.Next(); target != "1" {
		t.Fail()
	}

}

func TestFallbackBalancer_NextFirstError(t *testing.T) {
	balancers := []ILoadBalancer{
		&LoadBalancerMock{
			NextCallback: func() (string, error) {
				return "", fmt.Errorf("test")
			},
		},
		newInMemBalancer([]string{"4", "5", "6"}),
	}
	b := NewFallbackBalancer(nil, balancers...)

	t.Logf("Second balancer targets should be generated")
	if target, _ := b.Next(); target != "4" {
		t.Fail()
	}
}

func TestFallbackBalancer_NextError(t *testing.T) {
	balancers := []ILoadBalancer{
		&LoadBalancerMock{
			NextCallback: func() (string, error) {
				return "", fmt.Errorf("test")
			},
		},
		&LoadBalancerMock{
			NextCallback: func() (string, error) {
				return "", fmt.Errorf("test")
			},
		},
	}
	b := NewFallbackBalancer(nil, balancers...)

	t.Logf("Balancer has no targets, error should be returned")
	_, err := b.Next()
	if err == nil || !IsErrNoServiceAvailable(err) {
		t.Fail()
	}
}

func TestFallbackBalancer_Stats(t *testing.T) {
	balancers := []ILoadBalancer{
		&LoadBalancerMock{
			StatsCallback: func() []NodeStat {
				return nil
			},
		},
		newInMemBalancer([]string{"4", "5", "6"}),
	}
	b := NewFallbackBalancer(nil, balancers...)

	t.Logf("First available stas should be returned")
	if stats := b.Stats(); len(stats) != 3 {
		t.Fail()
	}
}
