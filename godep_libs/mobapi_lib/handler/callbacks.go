package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sergei-svistunov/gorpc"
	"github.com/sergei-svistunov/gorpc/transport/cache"
	"github.com/sergei-svistunov/gorpc/transport/http_json"
	"motify_core_api/godep_libs/go-trace"
	"motify_core_api/godep_libs/metrics"
	"motify_core_api/godep_libs/metrics/httpmon"
	ctxmanager "motify_core_api/godep_libs/mobapi_lib/context"
	"motify_core_api/godep_libs/mobapi_lib/logger"
	"motify_core_api/godep_libs/mobapi_lib/response"
	"motify_core_api/godep_libs/mobapi_lib/sessionlogger"
	"motify_core_api/godep_libs/mobapi_lib/sessionmocker"
	"motify_core_api/godep_libs/mobapi_lib/utils"
)

const (
	HeaderAPIToken      = "X-API-TOKEN"
	HeaderAppVersion    = "X-APP-VERSION"
	HeaderDebug         = "X-MTF-DEBUG"
	HeaderMock          = "X-MTF-MOCK"
	HeaderCID           = "X-CID"
	HeaderLang          = "X-MTF-LANG"
	HeaderWebPSupported = "X-WEBP-SUPPORTED"
	HeaderScreenWidth   = "X-SCREEN-WIDTH"
	HeaderScreenHeight  = "X-SCREEN-HEIGHT"
	HeaderScreenDensity = "X-SCREEN-DENSITY"
)

func NewHTTPHandlerCallbacks(serviceId string, appVersion string, hostname string, sessionLogger *sessionlogger.Logger) http_json.APIHandlerCallbacks {
	return http_json.APIHandlerCallbacks{
		OnEndServing: func(ctx context.Context, req *http.Request, startTime time.Time) {
			span := opentracing.SpanFromContext(ctx)
			if span == nil {
				return
			}
			code := "200"
			if r, ok := response.FromContext(ctx); ok {
				code = strconv.Itoa(r.StatusCode)
			}
			appName := ""
			if sc, ok := span.Context().(gotrace.Span); ok {
				appName = sc.ParentAppInfo.AppName
			}
			var (
				labels = map[string]string{
					"code":        code,
					"handler":     metrics.MakePathTag(req.URL.Path),
					"client_name": appName,
				}
				duration = metrics.SinceMs(startTime)
			)
			httpmon.ResponseTime.With(labels).Observe(duration)
			httpmon.ResponseTimeSummary.With(labels).Observe(duration)

			span.SetTag(string(ext.HTTPMethod), req.Method)
			span.SetTag(string(ext.HTTPStatusCode), code)

			ctxData, _ := ctxmanager.FromContext(ctx)
			if ctxData.ReqClientId != "" {
				span.SetTag("client.id", ctxData.ReqClientId)
			}
			if ctxData.ReqTokenHeader != "" {
				span.SetTag("client.token", ctxData.ReqTokenHeader)
			}
			if ctxData.ReqMobAppVersion != "" {
				span.SetTag("client.version", ctxData.ReqMobAppVersion)
			}
			span.Finish()
		},
		OnInitCtx: func(ctx context.Context, req *http.Request) context.Context {
			var (
				span    opentracing.Span
				traceId string
				appName string
				err     error
			)
			ctxData := ctxmanager.Context{
				ReqLang:          req.Header.Get(HeaderLang),
				ReqClientId:      req.Header.Get(HeaderCID),
				ReqMobAppVersion: GetVersionFromRequest(req),
				ReqTokenHeader:   req.Header.Get(HeaderAPIToken),
				ReqUserAgent:     req.Header.Get("User-Agent"),
				ReqURI:           req.RequestURI,
				ReqRemoteAddr:    req.RemoteAddr,
			}
			if req.Header.Get(gotrace.AppNameHeader) == "" {
				req.Header.Set(gotrace.AppNameHeader, GetPlatformFromUserAgent(ctxData.ReqUserAgent))
			}
			if req.Header.Get(gotrace.NodeHeader) == "" {
				req.Header.Set(gotrace.NodeHeader, req.RemoteAddr)
			}
			span, ctx = gotrace.StartSpanFromRequest(ctx, req.URL.Path, opentracing.HTTPHeadersCarrier(req.Header))
			if sc, ok := span.Context().(gotrace.Span); ok {
				traceId = sc.TraceID
				appName = sc.ParentAppInfo.AppName
			} else {
				logger.Error(ctx, opentracing.ErrInvalidSpanContext.Error())
			}

			if strings.ToLower(req.Header.Get(HeaderWebPSupported)) == "true" {
				ctxData.WebPSupported = true
			}

			if width, err := strconv.Atoi(req.Header.Get(HeaderScreenWidth)); err == nil {
				ctxData.ScreenWidth = width
			}

			if height, err := strconv.Atoi(req.Header.Get(HeaderScreenHeight)); err == nil {
				ctxData.ScreenHeight = height
			}

			if density, err := strconv.ParseFloat(req.Header.Get(HeaderScreenDensity), 10); err == nil {
				ctxData.ScreenDensity = density
			}

			httpmon.RequestCount.WithLabelValues(
				metrics.MakePathTag(req.URL.Path), appName).Inc()

			if strings.ToLower(req.Header.Get(HeaderDebug)) == "true" {
				cache.EnableDebug(ctx)
				ctxData.SessionLogger, err = sessionLogger.NewSession(
					traceId, "HTTP_JSON", CurlFromRequest(req))
			}
			if err != nil {
				logger.Error(ctx, "Can't create session logger: %s", err)
			}

			mockerEnabled := req.Header.Get(HeaderMock) == "true"
			ctxData.SessionMocker = sessionmocker.NewMocker(mockerEnabled, traceId)
			ctxData.SessionMocker.Request(req)
			if mockerEnabled {
				cache.EnableDebug(ctx)
			}
			ctx = ctxmanager.NewContext(ctx, &ctxData)
			ctx = response.NewContext(ctx, &response.Response{StatusCode: 200})
			return ctx
		},
		On404: func(ctx context.Context, req *http.Request) {
			if r, ok := response.FromContext(ctx); ok {
				r.StatusCode = 404
			}
			path := utils.GetPath(req)
			logger.Info(ctx, "path %q not found", path)
		},
		OnError: func(ctx context.Context, w http.ResponseWriter, req *http.Request, resp interface{}, err *gorpc.CallHandlerError) {
			var statusCode = 500
			var path = utils.GetPath(req)

			logger.Error(ctx, "Error on processing request '%q': %v", path, err)

			switch err.Type {
			case gorpc.ErrorReturnedFromCall:
				// business errors must be served as successful results
				statusCode = 200

				// try to translate business error message if dictionary is set in context
				if ctxData, _ := ctxmanager.FromContext(ctx); ctxData != nil {
					if dict := ctxData.GetErrorTranslatorDictionary(); dict != nil {
						if response, ok := resp.(*http_json.HttpSessionResponse); ok {
							response.Data, ok = dict.Translate(err.UserMessage(), nil, nil)
						} else {
							logger.Error(ctx, "Error message is not translated")
						}
					}
				}
			case gorpc.ErrorInParameters:
			case gorpc.ErrorInvalidMethod:
				logger.Error(ctx, "Invalid method has used")
			case gorpc.ErrorWriteResponse:
				if isBrokenPipe(err.Err) {
					logger.Error(ctx, "could not write response of %q: %v; client disconnected", path, err)
				} else {
					logger.Critical(ctx, "could not write response of %q: %v; response: %s", path, err, spew.Sprintf("%#v", resp))
				}
			}

			if r, ok := response.FromContext(ctx); ok {
				r.StatusCode = statusCode
			}

			if loggerSession, err := ctxmanager.GetLoggerSession(ctx); err == nil {
				loggerSession.Error(err)
			}
		},
		OnPanic: func(ctx context.Context, w http.ResponseWriter, r interface{}, trace []byte, req *http.Request) {
			if r, ok := response.FromContext(ctx); ok {
				r.StatusCode = 500
			}
			logger.Critical(ctx, "%v\n%s", r, trace)
			if loggerSession, err := ctxmanager.GetLoggerSession(ctx); err == nil {
				loggerSession.Error(string(trace))
			}
		},
		OnSuccess: func(ctx context.Context, req *http.Request, handlerResponse interface{}, startTime time.Time) {
			if loggerSession, err := ctxmanager.GetLoggerSession(ctx); err == nil {
				loggerSession.Finish(handlerResponse)
			}
		},
		OnBeforeWriteResponse: func(ctx context.Context, w http.ResponseWriter) {
			if err := gotrace.InjectSpanToResponseFromContext(ctx, opentracing.HTTPHeadersCarrier(w.Header())); err != nil {
				logger.Error(ctx, "Could not inject headers to response: %s", err)
			}
			cookies, ok := ctxmanager.GetResponseCookies(ctx)
			if ok {
				for _, cookie := range cookies {
					http.SetCookie(w, cookie)
				}
			}
		},
		GetCacheKey: func(ctx context.Context, req *http.Request, params interface{}) []byte {
			ctxData, err := ctxmanager.FromContext(ctx)
			if err != nil {
				logger.Warning(ctx, "%s", err)
				return nil
			}

			buf := bytes.NewBufferString(req.URL.Path)
			if ctxData.ReqTokenHeader != "" {
				buf.WriteString(ctxData.ReqTokenHeader)
			}

			if ctxData.WebPSupported {
				buf.WriteString("_webp_true_")
			}

			encoder := json.NewEncoder(buf)
			if err := encoder.Encode(params); err != nil {
				logger.Warning(ctx, "%s", err)
				return nil
			}
			return buf.Bytes()
		},
	}
}

func isBrokenPipe(err error) bool {
	if oe, oeok := err.(*net.OpError); oeok {
		if se, seok := oe.Err.(*os.SyscallError); seok {
			if se.Err != syscall.EPIPE {
				logger.Warning(nil, "DUMP: %#v", se.Err)
			}
			return se.Err == syscall.EPIPE
		}
	}
	return false
}
