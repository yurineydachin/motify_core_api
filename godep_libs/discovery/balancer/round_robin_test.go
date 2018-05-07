package balancer

import (
	"testing"
	"time"

	"motify_core_api/godep_libs/discovery/locator"
)

func TestRoundRobin_Next(t *testing.T) {
	balancer := createTestRRBalancer()
	defer balancer.Stop()

	if url, _ := balancer.Next(); url != "1" {
		t.Fail()
	}
	if url, _ := balancer.Next(); url != "2" {
		t.Fail()
	}
	if url, _ := balancer.Next(); url != "3" {
		t.Fail()
	}
	if url, _ := balancer.Next(); url != "1" {
		t.Fail()
	}
	if url, _ := balancer.Next(); url != "2" {
		t.Fail()
	}
	if url, _ := balancer.Next(); url != "3" {
		t.Fail()
	}
}

func TestRoundRobin_GetServiceName(t *testing.T) {
	balancer := createTestRRBalancer()
	defer balancer.Stop()

	actual := balancer.ServiceName()
	expected := "test"
	if actual != expected {
		t.Fatalf("expecting service name is %q, but got %q", actual, expected)
	}
}

// TODO: parallel
func TestRoundRobin_NextError(t *testing.T) {
	l := NewInMemLocator([]string{})
	balancer := NewRoundRobinEtcd2(l, nil, locator.LocationInfo{
		Namespace:   "test",
		Venture:     "test",
		Environment: "test",
		ServiceName: "test",
	})

	_, err := balancer.Next()
	if err == nil {
		t.Fatalf("error is expected")
	}

	t.Logf("Ready channel should be closed")
	select {
	case <-balancer.ready:
	default:
		t.Fatalf("balancer is not ready after first Next() call")
	}

	done := make(chan bool)
	go func() {
		balancer.Next()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("balancer blocks on Next() call, fast response is expected")
	}
}

func assertURLs(t *testing.T, urls, expectedURLs []string) {
	if len(urls) != len(expectedURLs) {
		t.Fatalf("actual urls of size %d, expected urls of size %d", len(urls), len(expectedURLs))
	}
	for i, url := range urls {
		expectedURL := expectedURLs[i]
		if url != expectedURL {
			t.Fatalf("actual url %q, expected url %q", url, expectedURL)
		}
	}
}

func createTestRRBalancer() ILoadBalancer {
	l := NewInMemLocator([]string{
		"test/test/test/test/1=1",
		"test/test/test/test/2=2",
		"test/test/test/test/3=3",
	})
	return NewRoundRobinEtcd2(l, nil, locator.LocationInfo{
		Namespace:   "test",
		Venture:     "test",
		Environment: "test",
		ServiceName: "test",
	})
}
