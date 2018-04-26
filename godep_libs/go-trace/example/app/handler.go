package app

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"time"

	"flag"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"godep.lzd.co/go-trace"
	"godep.lzd.co/metrics"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	AppdashPort     = flag.Int("appdash-port", 8282, "Appdash tracer port")
	Port            = flag.Int("port", 8080, "Listening port")
	GrpcPort        = flag.Int("grpc-port", 8181, "Listening GRPC port")
	registry        = metrics.DefaultRegistry()
	request_counter = registry.NewCounterVec("request_counter", "Counts requests", "status")
)

func MustNotErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		_, file = path.Split(file)
		log.Fatalf("Error occurred at %s:%d: %s", file, line, err)
	}
}

func HttpTracer(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var status = "200"
		span, ctx := gotrace.StartSpanFromRequest(
			r.Context(), metrics.MakePathTag(r.URL.Path), opentracing.HTTPHeadersCarrier(r.Header),
			gotrace.NewMetricSpanOption(request_counter, map[string]string{"status": status}),
			gotrace.RequestDataSpanOption(r.URL.RequestURI()))
		defer span.Finish()

		// inject response headers
		gotrace.InjectSpanToResponse(span.Context(), opentracing.HTTPHeadersCarrier(w.Header()))

		// call handler
		h(w, r.WithContext(ctx))

		// set HTTP tags
		span.SetTag(string(ext.HTTPStatusCode), status)
		span.SetTag(string(ext.HTTPMethod), r.Method)
		span.SetTag(string(ext.HTTPUrl), r.URL.Path)
	}
}

func GrpcTracer(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {

	md, _ := metadata.FromContext(ctx)
	span, ctx := gotrace.StartSpanFromRequest(ctx, info.FullMethod, gotrace.ProtoMetadataCarrier(md))
	defer span.Finish()

	resp, err = handler(ctx, req)

	// inject response metadata
	headers := metadata.MD{}
	gotrace.InjectSpanToResponse(span.Context(), gotrace.ProtoMetadataCarrier(headers))
	grpc.SetHeader(ctx, headers)
	return
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// set baggage
	sp := opentracing.SpanFromContext(r.Context())
	sp.SetBaggageItem("test", "value")

	// ext call
	serviceCall, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/service", *Port), nil)
	// inject the trace information into the HTTP Headers
	err := gotrace.InjectSpanToRequest(sp.Context(), opentracing.HTTPHeadersCarrier(serviceCall.Header))
	MustNotErr(err)

	if _, err := http.DefaultClient.Do(serviceCall); err != nil {
		log.Printf("%s: Synchronous call failed (%v)", r.URL.Path, err)
		return
	}

	// do some work
	time.Sleep(100 * time.Millisecond)

	writeSpan, _ := gotrace.StartSpanFromContext(r.Context(), "write-response")
	w.Write([]byte("done!"))
	writeSpan.Finish()
}

func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	sp := opentracing.SpanFromContext(r.Context())
	sp.SetTag("baggage.test", sp.BaggageItem("test"))
	sp.SetBaggageItem("test", "value-2")

	// create ext request
	dbCall, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/db", *Port), nil)

	// inject the trace information into the HTTP Headers in request
	err := gotrace.InjectSpanToRequest(sp.Context(), opentracing.HTTPHeadersCarrier(dbCall.Header))
	MustNotErr(err)

	// do request
	_, err = http.DefaultClient.Do(dbCall)
	MustNotErr(err)

	// do some additional work
	someWork(r.Context())
}

func someWork(ctx context.Context) {
	// start child Span for someWork
	sp, ctx := gotrace.StartSpanFromContext(ctx, "someWork")
	defer sp.Finish()

	// prepare grpc conn
	conn, err := grpc.Dial(":"+strconv.Itoa(*GrpcPort), grpc.WithInsecure())
	MustNotErr(err)
	defer conn.Close()
	c := NewExampleClient(conn)

	// inject the trace information into the protobuf Metadata
	md := metadata.New(nil)
	err = gotrace.InjectSpanToRequest(sp.Context(), gotrace.ProtoMetadataCarrier(md))
	MustNotErr(err)

	// response metadata
	respMD := metadata.New(nil)

	// do grpc call
	_, err = c.Hello(metadata.NewContext(ctx, md), &HelloRequest{Name: "Ivan"}, grpc.Header(&respMD))
	MustNotErr(err)

	// do some work
	time.Sleep(50 * time.Millisecond)

	// log some field
	sp.LogFields(otlog.String("log3", "hello3"))

	// do some work
	time.Sleep(50 * time.Millisecond)

	// log some field
	sp.LogFields(otlog.Int("log4", 4))
}

func DbHandler(w http.ResponseWriter, r *http.Request) {
	sp := opentracing.SpanFromContext(r.Context())
	// log sql query
	sp.LogKV("SQL.query", "SELECT 1")

	// do some work
	time.Sleep(100 * time.Millisecond)
}

type GrpcServerImpl struct{}

func (s *GrpcServerImpl) Hello(ctx context.Context, in *HelloRequest) (*HelloResponse, error) {
	sp := opentracing.SpanFromContext(ctx)

	// set tag from baggage
	sp.SetTag("baggage.test", sp.BaggageItem("test"))

	// log some fields
	sp.LogFields(
		otlog.String("log1", "hello1"),
		otlog.Int("log2", 2))

	// do some work
	time.Sleep(100 * time.Millisecond)

	return &HelloResponse{Hello: "Hello " + in.Name}, nil
}
