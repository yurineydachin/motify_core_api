package gotrace

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	. "gopkg.in/check.v1"
)

type TestOpenTracingSuite struct {
	TestSuite
}

var _ = Suite(&TestOpenTracingSuite{})

var tracer opentracing.Tracer

type mockGenerator struct {
	val string
}

func (mock *mockGenerator) Generate() string {
	return mock.val
}

func init() {
	tracer = NewTracer(WithAppEnv("service", "1.0.0", "localhost"))
	opentracing.SetGlobalTracer(tracer)
}

func (s *TestOpenTracingSuite) TestNewEmptySpan(c *C) {
	SetGenerator(&mockGenerator{val: "10"})

	sp, _ := StartSpanFromContext(context.Background(), "test")
	span := sp.Context().(Span)

	c.Assert(span.ParentSpanID, Equals, "")
	c.Assert(span.TraceID, Equals, "10")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.CurrentAppInfo.ForwardedApps, Equals, "service")
	c.Assert(span.CurrentAppInfo.AppName, Equals, "service")
	c.Assert(span.CurrentAppInfo.AppVersion, Equals, "1.0.0")
	c.Assert(span.CurrentAppInfo.Node, Equals, "localhost")
	c.Assert(span.ParentAppInfo.ForwardedApps, Equals, "")
	c.Assert(span.ParentAppInfo.AppName, Equals, "")
	c.Assert(span.ParentAppInfo.AppVersion, Equals, "")
	c.Assert(span.ParentAppInfo.Node, Equals, "")
}

func (s *TestOpenTracingSuite) TestNewSpanWithSpanIDButWithoutTraceID(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 1)
	headers.Add(SpanIDHeader, "1")

	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "10")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.ParentSpanID, Equals, "")
}

func (s *TestOpenTracingSuite) TestNewSpanWithTraceID(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 1)
	headers.Add(TraceIDHeader, "1")
	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "1")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.ParentSpanID, Equals, "")
}

func (s *TestOpenTracingSuite) TestNewSpanWithParentSpanID(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 1)
	headers.Add(ParentSpanIDHeader, "1")
	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "10")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.ParentSpanID, Equals, "1")
}

func (s *TestOpenTracingSuite) TestNewSpanWithTraceIDWithSpanIDWithParentSpanID(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 3)
	headers.Add(TraceIDHeader, "1")
	headers.Add(SpanIDHeader, "2")
	headers.Add(ParentSpanIDHeader, "3")

	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "1")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.ParentSpanID, Equals, "3")
}

func (s *TestOpenTracingSuite) TestNewSpanWithTraceIDWithSpanID(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 3)
	headers.Add(TraceIDHeader, "1")
	headers.Add(SpanIDHeader, "2")

	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "1")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.ParentSpanID, Equals, "")
}

func (s *TestOpenTracingSuite) TestNewSpanWithoutSpanID(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 2)
	headers.Add(TraceIDHeader, "1")
	headers.Add(ParentSpanIDHeader, "3")

	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "1")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.ParentSpanID, Equals, "3")
}

func (s *TestOpenTracingSuite) TestAppInfo(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 7)
	headers.Set(TraceIDHeader, "1")
	headers.Set(ForwardedAppsHeader, "test_forwarded_apps")
	headers.Set(AppNameHeader, "source_app_name")
	headers.Set(AppVersionHeader, "version.2.2.2")
	headers.Set(NodeHeader, "localhost:port")
	headers.Set(SegregationIDHeader, "999")
	headers.Set(ForceDebugLogHeader, "1")

	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.CurrentAppInfo.ForwardedApps, Equals, "test_forwarded_apps,service")
	c.Assert(span.CurrentAppInfo.AppName, Equals, "service")
	c.Assert(span.CurrentAppInfo.AppVersion, Equals, "1.0.0")
	c.Assert(span.CurrentAppInfo.Node, Equals, "localhost")
	c.Assert(span.ParentAppInfo.ForwardedApps, Equals, "test_forwarded_apps")
	c.Assert(span.ParentAppInfo.AppName, Equals, "source_app_name")
	c.Assert(span.ParentAppInfo.AppVersion, Equals, "version.2.2.2")
	c.Assert(span.ParentAppInfo.Node, Equals, "localhost:port")
	c.Assert(span.SegregationID, Equals, "999")
	c.Assert(span.ForceDebugLog, Equals, "1")
	c.Assert(span.TraceID, Equals, "1")
}

func (s *TestOpenTracingSuite) TestSetHeaderRequest(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 8)
	headers.Set(TraceIDHeader, "1")
	headers.Set(ParentSpanIDHeader, "3")
	headers.Set(ForwardedAppsHeader, "test_forwarded_apps")
	headers.Set(AppNameHeader, "source_app_name")
	headers.Set(AppVersionHeader, "version.2.2.2")
	headers.Set(NodeHeader, "localhost:port")
	headers.Set(SegregationIDHeader, "999")
	headers.Set(ForceDebugLogHeader, "1")

	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))

	headers2 := make(http.Header, 9)
	InjectSpanToRequest(sp.Context(), opentracing.HTTPHeadersCarrier(headers2))

	c.Assert(headers2.Get(TraceIDHeader), Equals, "1")
	c.Assert(headers2.Get(SpanIDHeader), Equals, "")
	c.Assert(headers2.Get(ParentSpanIDHeader), Equals, "10")
	c.Assert(headers2.Get(ForwardedAppsHeader), Equals, "test_forwarded_apps,service")
	c.Assert(headers2.Get(AppNameHeader), Equals, "service")
	c.Assert(headers2.Get(AppVersionHeader), Equals, "1.0.0")
	c.Assert(headers2.Get(NodeHeader), Equals, "localhost")
	c.Assert(headers2.Get(SegregationIDHeader), Equals, "999")
	c.Assert(headers2.Get(ForceDebugLogHeader), Equals, "1")
}

func (s *TestOpenTracingSuite) TestSetHeaderResponse(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 7)
	headers.Set(TraceIDHeader, "1")
	headers.Set(ParentSpanIDHeader, "3")
	headers.Set(AppNameHeader, "source_app_name")
	headers.Set(AppVersionHeader, "version.2.2.2")
	headers.Set(NodeHeader, "localhost:port")
	headers.Set(SegregationIDHeader, "999")
	headers.Set(ForceDebugLogHeader, "1")

	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))

	headers2 := make(http.Header, 8)
	InjectSpanToResponse(sp.Context(), opentracing.HTTPHeadersCarrier(headers2))

	c.Assert(headers2.Get(TraceIDHeader), Equals, "1")
	c.Assert(headers2.Get(SpanIDHeader), Equals, "10")
	c.Assert(headers2.Get(ParentSpanIDHeader), Equals, "3")
	c.Assert(headers2.Get(AppNameHeader), Equals, "service")
	c.Assert(headers2.Get(AppVersionHeader), Equals, "1.0.0")
	c.Assert(headers2.Get(NodeHeader), Equals, "localhost")
	c.Assert(headers2.Get(SegregationIDHeader), Equals, "999")
	c.Assert(headers2.Get(ForceDebugLogHeader), Equals, "1")
}

func (s *TestOpenTracingSuite) TestSpanFromAnotherSpan(c *C) {
	SetGenerator(&mockGenerator{val: "10"})

	_, ctx := StartSpanFromContext(context.Background(), "parent")

	SetGenerator(&mockGenerator{val: "11"})

	sp, _ := StartSpanFromContext(ctx, "child")
	span := sp.Context().(Span)

	c.Assert(span.ParentSpanID, Equals, "10")
	c.Assert(span.TraceID, Equals, "10")
	c.Assert(span.SpanID, Equals, "11")
}

func (s *TestOpenTracingSuite) TestGetGotraceSpan(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 8)
	headers.Set(TraceIDHeader, "1")
	headers.Set(ParentSpanIDHeader, "3")
	headers.Set(ForwardedAppsHeader, "test_forwarded_apps")
	headers.Set(AppNameHeader, "source_app_name")
	headers.Set(AppVersionHeader, "version.2.2.2")
	headers.Set(NodeHeader, "localhost:port")
	headers.Set(SegregationIDHeader, "999")
	headers.Set(ForceDebugLogHeader, "1")
	expected := Span{
		TraceID:      "1",
		SpanID:       "10",
		ParentSpanID: "3",
		CurrentAppInfo: AppInfo{
			ForwardedApps: "test_forwarded_apps,service",
			AppName:       "service",
			AppVersion:    "1.0.0",
			Node:          "localhost",
		},
		ParentAppInfo: AppInfo{
			ForwardedApps: "test_forwarded_apps",
			AppName:       "source_app_name",
			AppVersion:    "version.2.2.2",
			Node:          "localhost:port",
		},
		SegregationID: "999",
		ForceDebugLog: "1",
		Baggage:       map[string]string(nil),
	}
	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))

	gotraceSpan, err := SpanContext(sp)

	c.Assert(err, IsNil)
	c.Assert(gotraceSpan, DeepEquals, expected)
}

func (s *TestOpenTracingSuite) TestNewSpanWithoutHeaders(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(nil))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "10")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.CurrentAppInfo.ForwardedApps, Equals, "service")
	c.Assert(span.CurrentAppInfo.AppName, Equals, "service")
	c.Assert(span.CurrentAppInfo.AppVersion, Equals, "1.0.0")
	c.Assert(span.CurrentAppInfo.Node, Equals, "localhost")
	c.Assert(span.ParentAppInfo.ForwardedApps, Equals, "")
	c.Assert(span.ParentAppInfo.AppName, Equals, "")
	c.Assert(span.ParentAppInfo.AppVersion, Equals, "")
	c.Assert(span.ParentAppInfo.Node, Equals, "")
}

func (s *TestOpenTracingSuite) TestNewSpanWithSegregationIDHeader(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 1)
	headers.Add(SegregationIDHeader, "999")
	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "10")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.SegregationID, Equals, "999")
}

func (s *TestOpenTracingSuite) TestNewSpanWithForceDebugLogHeader(c *C) {
	SetGenerator(&mockGenerator{val: "10"})
	headers := make(http.Header, 1)
	headers.Add(ForceDebugLogHeader, "true")
	sp, _ := StartSpanFromRequest(context.Background(), "test", opentracing.HTTPHeadersCarrier(headers))
	span := sp.Context().(Span)

	c.Assert(span.TraceID, Equals, "10")
	c.Assert(span.SpanID, Equals, "10")
	c.Assert(span.ForceDebugLog, Equals, "true")
}
