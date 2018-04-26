package tracking

import (
	"context"
	"net/http"

	"motify_core_api/godep_libs/service/flake"
)

const trackingKey = "tracking"

var f *flake.Flake

func init() {
	var err error
	f, err = flake.WithRandomID()
	if err != nil {
		panic(err)
	}
}

type Tracking struct {
	AppName       string
	AppVersion    string
	Node          string
	ForwardedApps string
	TraceID       string
	SpanID        string
	ParentSpanID  string
}

func NewTracking(name, version, node, forwardedApps string, traceID, parentSpanID string) *Tracking {
	if traceID == "" {
		traceID = nextID()
	}

	if forwardedApps == "" {
		forwardedApps = name
	} else {
		forwardedApps += "," + name
	}

	return &Tracking{
		AppName:       name,
		AppVersion:    version,
		Node:          node,
		ForwardedApps: forwardedApps,
		TraceID:       traceID,
		SpanID:        nextID(),
		ParentSpanID:  parentSpanID,
	}
}

func nextID() string {
	id := f.NextID()
	return id.String()
}

func NewContext(ctx context.Context, t *Tracking) context.Context {
	return context.WithValue(ctx, trackingKey, t)
}

func FromContext(ctx context.Context) (t *Tracking, ok bool) {
	if ctx == nil {
		return
	}
	ctxVal := ctx.Value(trackingKey)
	if ctxVal != nil {
		t, ok = ctxVal.(*Tracking)
	}
	return
}

func FromRequest(req *http.Request, name, version, node string) *Tracking {
	return NewTracking(name, version, node,
		req.Header.Get("X-LZD-Forwarded-Apps"),
		req.Header.Get("X-LZD-TraceId"),
		req.Header.Get("X-LZD-SpanId"))
}

func SetRequestHeaders(ctx context.Context, header http.Header) {
	if t, ok := FromContext(ctx); ok {
		header.Set("X-LZD-App-Name", t.AppName)
		header.Set("X-LZD-App-Version", t.AppVersion)
		header.Set("X-LZD-Node", t.Node)
		header.Set("X-LZD-Forwarded-Apps", t.ForwardedApps)
		header.Set("X-LZD-TraceId", t.TraceID)
		header.Set("X-LZD-SpanId", t.SpanID)
		if t.ParentSpanID != "" {
			header.Set("X-LZD-ParentSpanId", t.ParentSpanID)
		}
	}
}

func SetResponseHeaders(ctx context.Context, header http.Header) {
	if t, ok := FromContext(ctx); ok {
		header.Set("X-LZD-App-Name", t.AppName)
		header.Set("X-LZD-App-Version", t.AppVersion)
		header.Set("X-LZD-Node", t.Node)
		header.Set("X-LZD-TraceId", t.TraceID)
		header.Set("X-LZD-SpanId", t.SpanID)
		if t.ParentSpanID != "" {
			header.Set("X-LZD-ParentSpanId", t.ParentSpanID)
		}
	}
}
