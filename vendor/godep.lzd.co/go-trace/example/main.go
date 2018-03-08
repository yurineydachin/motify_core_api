package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	golog "godep.lzd.co/go-log"
	"godep.lzd.co/go-trace"
	"godep.lzd.co/go-trace/example/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"sourcegraph.com/sourcegraph/appdash"
	adTracing "sourcegraph.com/sourcegraph/appdash/opentracing"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

func init() {
	flag.Parse()

	var (
		logger       = golog.NewLogger("service", "", "", golog.DEBUG) // Setup golog
		logCollector = golog.NewSpanCollector(logger)                      // Setup golog Span Collector

		gtOpts     = gotrace.WithLogCollector(logCollector) // options for gtRecorder
		gtRecorder = gotrace.NewRecorder(gtOpts)            // Setup gotrace gtRecorder.

		addr        = startAppdashServer(*app.AppdashPort)       // Setup appdash server,
		adCollector = appdash.NewRemoteCollector(addr)           // collector,
		adOpts      = adTracing.DefaultOptions()                 // options and
		adRecorder  = adTracing.NewRecorder(adCollector, adOpts) // gtRecorder,
		btRecorder  = app.NewBasictracerRecorder(adRecorder)     // then pass it to basictracer-compatible gtRecorder.

		zipCollector, _ = zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans") // zipkin collector
		zipRecorder     = app.NewZipkinRecorder(zipkin.NewRecorder(zipCollector, true, "localhost", "service"))
	)

	fmt.Println("zipkin: http://localhost:9411/zipkin/")

	tracer := gotrace.NewTracer(
		gotrace.WithAppEnv("service", "1.0.0", "localhost"),
		gotrace.WithSpanRecorder(gtRecorder),
		gotrace.WithSpanRecorder(btRecorder),
		gotrace.WithSpanRecorder(zipRecorder),
	)

	// Start processing traces.
	opentracing.InitGlobalTracer(tracer)
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(3)
	go grpcServer(&wg)
	go httpServer(&wg)
	go prometheusServer(&wg)
	wg.Wait()
}

func httpServer(wg *sync.WaitGroup) {
	defer wg.Done()
	addr := "localhost:" + strconv.Itoa(*app.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.HttpTracer(app.IndexHandler))
	mux.HandleFunc("/service", app.HttpTracer(app.ServiceHandler))
	mux.HandleFunc("/db", app.HttpTracer(app.DbHandler))

	fmt.Printf("Go to http://%s to start a request!\n", addr)
	http.ListenAndServe(addr, mux)
}

func grpcServer(wg *sync.WaitGroup) {
	defer wg.Done()

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(*app.GrpcPort))
	app.MustNotErr(err)

	s := grpc.NewServer(grpc.UnaryInterceptor(app.GrpcTracer))
	app.RegisterExampleServer(s, &app.GrpcServerImpl{})

	reflection.Register(s)
	app.MustNotErr(s.Serve(lis))
}

func prometheusServer(wg *sync.WaitGroup) {
	defer wg.Done()

	func() {
		if err := http.ListenAndServe(":8585", promhttp.Handler()); err != nil {
			log.Printf("Prometheus handler failed: %v\n", err)
		}
	}()
}

// startAppdashServer returns the remote collector address.
func startAppdashServer(appdashPort int) string {
	store := appdash.NewMemoryStore()

	// Listen on any available TCP port locally.
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	app.MustNotErr(err)

	collectorPort := l.Addr().(*net.TCPAddr).Port

	// Start an Appdash collection server that will listen for spans and
	// annotations and add them to the local collector (stored in-memory).
	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
	go cs.Start()

	// Print the URL at which the web UI will be running.
	appdashURLStr := fmt.Sprintf("http://localhost:%d", appdashPort)
	appdashURL, err := url.Parse(appdashURLStr)
	app.MustNotErr(err)

	fmt.Printf("appdash: %s/traces\n", appdashURL)

	// Start the web UI in a separate goroutine.
	tapp, err := traceapp.New(nil, appdashURL)
	app.MustNotErr(err)

	tapp.Store = store
	tapp.Queryer = store
	go func() {
		app.MustNotErr(http.ListenAndServe(fmt.Sprintf(":%d", appdashPort), tapp))
	}()
	return fmt.Sprintf(":%d", collectorPort)
}
