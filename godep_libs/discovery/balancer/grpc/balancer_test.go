package grpc

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	etcdcl3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/integration"
	"motify_core_api/godep_libs/discovery/balancer"
	pb "motify_core_api/godep_libs/discovery/balancer/grpc/proto"
	"motify_core_api/godep_libs/discovery/locator"
	"motify_core_api/godep_libs/discovery/provider/etcdV3"
	"motify_core_api/godep_libs/discovery/registrator"
	"motify_core_api/godep_libs/go-logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var serverLogger = logger.NewBuiltinLogger(log.New(os.Stderr, "SERVER: ", log.Ldate|log.Ltime), logger.LevelWarning)
var clientLogger = logger.NewBuiltinLogger(log.New(os.Stderr, "CLIENT: ", log.Ldate|log.Ltime), logger.LevelWarning)

func createTestWRRBalancer() balancer.ILoadBalancer {
	l := balancer.NewInMemLocator([]string{
		"test/test/test/test/1=1",
		"test/test/test/test/2=2",
		"test/test/test/test/3=3",
	})
	return balancer.NewWeightedRoundRobinEtcd2(l, nil, locator.LocationInfo{
		Namespace:   "test",
		Venture:     "test",
		Environment: "test",
		ServiceName: "test",
	})
}

func TestGrpcBalancer_Get(t *testing.T) {
	b := NewBalancer(createTestWRRBalancer())
	if err := b.Start("", grpc.BalancerConfig{}); err != nil {
		t.Fatal(err)
		return
	}
	defer b.Close()

	<-b.Notify()
	url, _, err := b.Get(context.Background(), grpc.BalancerGetOptions{})
	if err != nil {
		t.Fatal(err)
		return
	}

	if url.Addr != "1" {
		t.Errorf("expected url is: '1', got: '%s'\n", url.Addr)
	}
}

func TestGrpcBalancer_Notify(t *testing.T) {
	b := NewBalancer(createTestWRRBalancer())
	if err := b.Start("", grpc.BalancerConfig{}); err != nil {
		t.Fatal(err)
		return
	}
	defer b.Close()

	for i, url := range <-b.Notify() {
		if url.Addr != strconv.Itoa(i+1) {
			t.Errorf("expected url is: '%s', got: '%s'\n", strconv.Itoa(i+1), url.Addr)
		}
	}
}

func registerInEtcd(t *testing.T, reg registrator.IRegistrator) {
	if err := reg.Register(); err != nil {
		t.Fatalf("Cannot register service to ETCD discovery: %s", err)
	}

	if err := reg.EnableDiscovery(); err != nil {
		t.Fatalf("Cannot enable dicovery ability: %s", err)
	}
}

func getEtcdRegistrator(t *testing.T, extServiceName string, endpoints []string, addr string) registrator.IRegistrator {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	grpcPort, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	etcdClientV3, err := etcdcl3.New(
		etcdcl3.Config{
			Endpoints:   endpoints,
			DialTimeout: 1 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("Failed to create etcd client v3: %v", err)
	}
	info, err := registrator.NewAppRegistrationInfo(registrator.AppRegistrationParams{
		ServiceName:    extServiceName,
		RolloutType:    "stable",
		Host:           host,
		HTTPPort:       grpcPort - 6,
		AdminPort:      grpcPort - 4,
		GRPCPort:       grpcPort,
		MonitoringPort: grpcPort - 2,
		Version: registrator.VersionInfo{
			AppVersion: "1.0",
		},
		Venture:     "id",
		Environment: "dev",
	})
	if err != nil {
		t.Fatalf("Failed to create etcd client v3: %v", err)
	}
	etcdRegistrator := registrator.New(etcdV3.NewProvider(etcdClientV3, serverLogger), info, serverLogger)

	return etcdRegistrator
}

func newEtcd(t *testing.T) (cluster *integration.ClusterV3, endpoints []string) {
	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 3})
	eps := make([]string, 3)
	for i := range eps {
		eps[i] = clus.Members[i].GRPCAddr()
	}
	return clus, eps
}

type testServer struct {
	req int
}

func (s *testServer) Call(ctx context.Context, req *pb.TestRequest) (*pb.TestResponse, error) {
	s.req++
	return &pb.TestResponse{
		Message: "Message: " + req.Message,
	}, nil
}

func newGrpcServer(t *testing.T, impl pb.TestServer) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterTestServer(server, impl)
	go server.Serve(lis)

	return server, lis
}

type balancerCreator func(serviceName string, endpoints []string) balancer.ILoadBalancer

func newWRRBalancer(serviceName string, endpoints []string) balancer.ILoadBalancer {
	etcdClientV3, err := etcdcl3.New(
		etcdcl3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
		},
	)
	if err != nil {
		return nil
	}

	etcdProvider := etcdV3.NewProvider(etcdClientV3, clientLogger)
	loc := locator.New(etcdProvider, clientLogger)
	return balancer.NewWeightedRoundRobin(
		loc, clientLogger, balancer.LoadBalancerOptions{ServiceName: serviceName, EndpointType: locator.TypeAppAdditional})

}

func newRRBalancer(serviceName string, endpoints []string) balancer.ILoadBalancer {
	etcdClientV3, err := etcdcl3.New(
		etcdcl3.Config{
			Endpoints:   endpoints,
			DialTimeout: 5 * time.Second,
		},
	)
	if err != nil {
		return nil
	}

	etcdProvider := etcdV3.NewProvider(etcdClientV3, clientLogger)
	loc := locator.New(etcdProvider, clientLogger)
	return balancer.NewRoundRobin(
		loc, clientLogger, balancer.LoadBalancerOptions{ServiceName: serviceName, EndpointType: locator.TypeAppAdditional})
}

func newGrpcClient(b interface{}, withBlock bool) (conn *grpc.ClientConn, client pb.TestClient, err error) {
	var gb *Balancer
	if v, ok := b.(*Balancer); ok {
		gb = v
	} else if v, ok := b.(balancer.ILoadBalancer); ok {
		gb = NewBalancer(v)
	}

	opts := []grpc.DialOption{
		grpc.WithBalancer(gb),
		grpc.WithInsecure(),
		grpc.WithTimeout(6 * time.Second),
		grpc.WithBackoffConfig(grpc.BackoffConfig{
			MaxDelay: 6 * time.Second,
		}),
	}

	if withBlock {
		opts = append(opts, grpc.WithBlock())
	}

	if conn, err = grpc.Dial("", opts...); err != nil {
		return nil, nil, err
	}

	return conn, pb.NewTestClient(conn), nil
}

func testBalancerNoNodes(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatalf("grpc.Dial() failed: %s", err)
	}
	defer conn.Close()

	_, err = client.Call(context.Background(), &pb.TestRequest{Message: "test"})

	if s, ok := status.FromError(err); !ok || !strings.HasSuffix(s.Message(), "no service available") {
		t.Errorf("expected err '...no service available', got '%s'\n", err)
	}
}

func TestRRBalancerNoNodes(t *testing.T) {
	testBalancerNoNodes(t, newRRBalancer)
}

func TestWRRBalancerNoNodes(t *testing.T) {
	testBalancerNoNodes(t, newWRRBalancer)
}

func testBalancerSimple(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	time.Sleep(2 * time.Second)
	resp, err := client.Call(context.Background(), &pb.TestRequest{
		Message: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message != "Message: test" {
		t.Errorf("expected resp: '%s', got: '%s'", "Message: test", resp.Message)
	}
}

func TestWRRBalancerSimple(t *testing.T) {
	testBalancerSimple(t, newWRRBalancer)
}

func TestRRBalancerSimple(t *testing.T) {
	testBalancerSimple(t, newRRBalancer)
}

func testBalancerTwoNodes(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	srvImpl2 := &testServer{}
	server2, lis2 := newGrpcServer(t, srvImpl2)
	defer server2.Stop()
	defer lis2.Close()

	reg2 := getEtcdRegistrator(t, "TestServer", endpoints, lis2.Addr().String())
	registerInEtcd(t, reg2)
	defer reg2.Unregister()

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// TODO: remove Sleep
	time.Sleep(2 * time.Second)

	for i := 0; i < 100; i++ {
		_, _ = client.Call(context.Background(), &pb.TestRequest{})
	}

	if srvImpl1.req == 0 {
		t.Error("expected requests count for first node > 0")
	}
	if srvImpl2.req == 0 {
		t.Error("expected requests count for second node > 0")
	}
}

func TestRRBalancerTwoNodes(t *testing.T) {
	testBalancerTwoNodes(t, newRRBalancer)
}

func TestWRRBalancerTwoNodes(t *testing.T) {
	testBalancerTwoNodes(t, newWRRBalancer)
}

func testBalancerLastNodeWillUnregisterFromEtcd(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)

	// If register event takes too much time, unregister event (see below) may occur before it (?), and test will fail
	time.Sleep(2 * time.Second) // So we sleep here to make sure that node is registered before going on

	defer reg1.Unregister()

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	reg1.Unregister()

	// TODO: remove Sleep
	time.Sleep(2 * time.Second)

	_, err = client.Call(context.Background(), &pb.TestRequest{})

	expectedSubstring := "no service available"
	if !strings.Contains(grpc.ErrorDesc(err), expectedSubstring) {
		t.Errorf("expected ErrNoServiceAvailable, got '%#v'\n", err)
	}
}

func TestRRBalancerLastNodeWillUnregisterFromEtcd(t *testing.T) {
	testBalancerLastNodeWillUnregisterFromEtcd(t, newRRBalancer)
}

func TestWRRBalancerLastNodeWillUnregisterFromEtcd(t *testing.T) {
	testBalancerLastNodeWillUnregisterFromEtcd(t, newWRRBalancer)
}

func testBalancerOneNodeWillUnregisterFromEtcdTraficShouldGoToSecondNode(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	srvImpl2 := &testServer{}
	server2, lis2 := newGrpcServer(t, srvImpl2)
	defer server2.Stop()
	defer lis2.Close()

	reg2 := getEtcdRegistrator(t, "TestServer", endpoints, lis2.Addr().String())
	registerInEtcd(t, reg2)
	defer reg2.Unregister()

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	time.Sleep(2 * time.Second)

	reg1.Unregister()

	// TODO: remove Sleep
	time.Sleep(2 * time.Second)

	_, _ = client.Call(context.Background(), &pb.TestRequest{})

	if srvImpl1.req != 0 {
		t.Errorf("expected requests count for first node 0, got: %d", srvImpl1.req)
	}

	if srvImpl2.req == 0 {
		t.Errorf("expected requests count for second node 1, got: %d", srvImpl2.req)
	}
}

func TestRRBalancerOneNodeWillUnregisterFromEtcdTraficShouldGoToSecondNode(t *testing.T) {
	testBalancerOneNodeWillUnregisterFromEtcdTraficShouldGoToSecondNode(t, newRRBalancer)
}

func TestWRRBalancerOneNodeWillUnregisterFromEtcdTraficShouldGoToSecondNode(t *testing.T) {
	testBalancerOneNodeWillUnregisterFromEtcdTraficShouldGoToSecondNode(t, newWRRBalancer)
}

func testBalancerAddNodeAfterClientStartTraficShouldGoToNewNodeToo(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	srvImpl2 := &testServer{}
	server2, lis2 := newGrpcServer(t, srvImpl2)
	defer server2.Stop()
	defer lis2.Close()

	reg2 := getEtcdRegistrator(t, "TestServer", endpoints, lis2.Addr().String())
	registerInEtcd(t, reg2)
	defer reg2.Unregister()

	for i := 0; i < 1000; i++ {
		_, _ = client.Call(context.Background(), &pb.TestRequest{})
	}

	if srvImpl1.req == 0 {
		t.Error("expected requests count for first node > 0")
	}
	if srvImpl2.req == 0 {
		t.Error("expected requests count for second node > 0")
	}
}

func TestRRBalancerAddNodesAfterClientStartTraficShouldGoToNewNodes(t *testing.T) {
	testBalancerAddNodeAfterClientStartTraficShouldGoToNewNodeToo(t, newRRBalancer)
}

func TestWRRBalancerAddNodesAfterClientStartTraficShouldGoToNewNodes(t *testing.T) {
	testBalancerAddNodeAfterClientStartTraficShouldGoToNewNodeToo(t, newWRRBalancer)
}

func testBalancerWithFailFastTrueOneNodeWillStopTraficShouldGoToSecondNode(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	srvImpl2 := &testServer{}
	server2, lis2 := newGrpcServer(t, srvImpl2)
	defer server2.Stop()
	defer lis2.Close()

	reg2 := getEtcdRegistrator(t, "TestServer", endpoints, lis2.Addr().String())
	registerInEtcd(t, reg2)
	defer reg2.Unregister()

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	server1.Stop()

	// TODO: remove Sleep
	time.Sleep(2 * time.Second)

	_, err = client.Call(context.Background(), &pb.TestRequest{}, grpc.FailFast(true))
	if err != nil {
		t.Fatal(err)
	}

	if srvImpl1.req != 0 {
		t.Errorf("expected requests count for first node == 0, got: %d", srvImpl1.req)
	}
	if srvImpl2.req == 0 {
		t.Error("expected requests count for second node > 0")
	}
}

func TestRRBalancerWithFailFastTrueOneNodeWillStopTraficShouldGoToSecondNode(t *testing.T) {
	testBalancerWithFailFastTrueOneNodeWillStopTraficShouldGoToSecondNode(t, newRRBalancer)
}

func TestWRRBalancerWithFailFastTrueOneNodeWillStopTraficShouldGoToSecondNode(t *testing.T) {
	testBalancerWithFailFastTrueOneNodeWillStopTraficShouldGoToSecondNode(t, newWRRBalancer)
}

func testBalancerWithFailFastFalseOneNodeWillStopTraficShouldGoToSecondNode(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	srvImpl2 := &testServer{}
	server2, lis2 := newGrpcServer(t, srvImpl2)
	defer server2.Stop()
	defer lis2.Close()

	reg2 := getEtcdRegistrator(t, "TestServer", endpoints, lis2.Addr().String())
	registerInEtcd(t, reg2)
	defer reg2.Unregister()

	conn, client, err := newGrpcClient(b("TestServer", endpoints), false)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	server1.Stop()

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(3*time.Second))
	_, err = client.Call(ctx, &pb.TestRequest{}, grpc.FailFast(false))
	if err != nil {
		t.Fatal(err)
	}

	if srvImpl1.req != 0 {
		t.Errorf("expected requests count for first node == 0, got: %d", srvImpl1.req)
	}
	if srvImpl2.req == 0 {
		t.Error("expected requests count for second node > 0")
	}
}

func TestRRBalancerWithFailFastFalseOneNodeWillStopTraficShouldGoToSecondNode(t *testing.T) {
	testBalancerWithFailFastFalseOneNodeWillStopTraficShouldGoToSecondNode(t, newRRBalancer)
}

func TestWRRBalancerWithFailFastFalseOneNodeWillStopTraficShouldGoToSecondNode(t *testing.T) {
	testBalancerWithFailFastFalseOneNodeWillStopTraficShouldGoToSecondNode(t, newWRRBalancer)
}

// Test case for issue: GOLIBS-1354
func TestReuseGRPCBalancerWillNotCausePanic(t *testing.T) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	b := NewBalancer(newWRRBalancer("TestServer", endpoints))
	for i := 0; i < 10; i++ {
		_, _, err := newGrpcClient(b, false)
		if err != nil {
			t.Errorf("expected err <nil>, got '%s'\n", err)
		}
	}
}

func testBalancerAddNodeAfterClientStart(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	gb := NewBalancer(b("TestServer", endpoints))
	conn, client, err := newGrpcClient(gb, false)
	if err != nil {
		t.Fatalf("grpc.Dial() failed: %s", err)
	}
	defer conn.Close()

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	// TODO: remove sleep
	time.Sleep(2 * time.Second)

	_, _ = client.Call(context.Background(), &pb.TestRequest{})

	if srvImpl1.req == 0 {
		t.Error("expected requests count for first node > 0")
	}
}

func TestRRBalancerAddNodeAfterClientStart(t *testing.T) {
	testBalancerAddNodeAfterClientStart(t, newRRBalancer)
}

func TestWRRBalancerAddNodeAfterClientStart(t *testing.T) {
	testBalancerAddNodeAfterClientStart(t, newWRRBalancer)
}

func testReuseBalancer(t *testing.T, b balancerCreator) {
	clus, endpoints := newEtcd(t)
	defer clus.Terminate(t)

	gb := NewBalancer(b("TestServer", endpoints))
	conn1, _, err := newGrpcClient(gb, false)
	if err != nil {
		t.Fatalf("grpc.Dial() failed: %s", err)
	}
	defer conn1.Close()

	srvImpl1 := &testServer{}
	server1, lis1 := newGrpcServer(t, srvImpl1)
	defer server1.Stop()
	defer lis1.Close()

	reg1 := getEtcdRegistrator(t, "TestServer", endpoints, lis1.Addr().String())
	registerInEtcd(t, reg1)
	defer reg1.Unregister()

	// TODO: remove sleep
	time.Sleep(2 * time.Second)

	//uncomment this and test will be passed
	//gb = NewBalancer(b("TestServer", endpoints))
	conn2, client2, err := newGrpcClient(gb, false)
	if err != nil {
		t.Fatalf("grpc.Dial() failed: %s", err)
	}
	defer conn2.Close()

	_, _ = client2.Call(context.Background(), &pb.TestRequest{})

	if srvImpl1.req == 0 {
		t.Error("expected requests count for first node > 0")
	}
}

// TODO this test should be fixed and enabled, see https://jira.lzd.co/browse/GOLIBS-1432
func _TestReuseRRBalancer(t *testing.T) {
	testReuseBalancer(t, newRRBalancer)
}

// TODO this test should be fixed and enabled, see https://jira.lzd.co/browse/GOLIBS-1432
func _TestReuseWRRBalancer(t *testing.T) {
	testReuseBalancer(t, newWRRBalancer)
}
