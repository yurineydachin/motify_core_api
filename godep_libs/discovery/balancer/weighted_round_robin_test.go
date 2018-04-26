package balancer

import (
	"testing"
	"time"

	"godep.lzd.co/discovery/locator"
)

func TestStats_Create(t *testing.T) {
	stats := rttStats{}

	if stats.avg != 0 {
		t.FailNow()
	}
}

func TestStats_Add(t *testing.T) {
	stats := rttStats{}

	stats.Add(maxResponseTime)

	if stats.avg != maxResponseTime {
		t.Fatalf("expected %s, actual %s", maxResponseTime, stats.avg)
	}
}

func TestStats_AverageIs1Second_WhenFilledWith1Second(t *testing.T) {
	stats := rttStats{}

	for i := 0; i < rttHistorySize; i++ {
		stats.Add(time.Second)
	}

	if stats.avg != time.Second {
		t.Fatalf("expected %s, actual %s", time.Second, stats.avg)
	}
}

func TestStats_Add10Seconds(t *testing.T) {
	stats := rttStats{}

	for i := 0; i < rttHistorySize; i++ {
		stats.Add(time.Second)
	}
	stats.Add(rttHistorySize * time.Second)

	// Warning: Hardcoded value!
	// Depends on the EXP value, chosen for weightedAverage.
	// Facepalm:
	//   .-'---`-.
	// ,'          `.
	// |             \
	// |              \
	// \           _  \
	// ,\  _    ,'-,/-)\
	// ( * \ \,' ,' ,'-)
	//  `._,)     -',-')
	//    \/         ''/
	//     )        / /
	//    /       ,'-'
	expected, _ := time.ParseDuration("1.049804687s")
	if stats.avg != expected {
		t.Fatalf("expected %s, actual %s", expected, stats.avg)
	}
}

func TestStats_AddOneSecond(t *testing.T) {
	stats := rttStats{}

	for i := 0; i < rttHistorySize; i++ {
		stats.Add(rttHistorySize * time.Second)
	}
	stats.Add(time.Second)

	// Warning: Hardcoded value!
	expected, _ := time.ParseDuration("3.950195312s")
	if stats.avg != expected {
		t.Fatalf("expected %s, actual %s", expected, stats.avg)
	}
}

func TestWeightedRoundRobin_Next(t *testing.T) {
	balancer := createTestWRRBalancer()
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

// TODO: parallel
func TestWeightedRoundRobin_NextError(t *testing.T) {
	l := NewInMemLocator([]string{})
	balancer := NewWeightedRoundRobinEtcd2(l, nil, locator.LocationInfo{
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

func TestWeightedRoundRobin_GetServiceName(t *testing.T) {
	balancer := createTestWRRBalancer()
	defer balancer.Stop()

	actual := balancer.ServiceName()
	expected := "test"
	if actual != expected {
		t.Fatalf("expecting service name is %q, but got %q", actual, expected)
	}
}

func createTestWRRBalancer() ILoadBalancer {
	l := NewInMemLocator([]string{
		"test/test/test/test/1=1",
		"test/test/test/test/2=2",
		"test/test/test/test/3=3",
	})
	return NewWeightedRoundRobinEtcd2(l, nil, locator.LocationInfo{
		Namespace:   "test",
		Venture:     "test",
		Environment: "test",
		ServiceName: "test",
	})
}
