package adapter

import (
	"context"
	"fmt"
	"net/http"
	"time"

//	"godep.lzd.co/go-trace"
	"godep.lzd.co/service"
	"godep.lzd.co/service/tracking"
	"godep.lzd.co/mobapi_lib/cache/inmem"
	ctxmanager "godep.lzd.co/mobapi_lib/context"
	"godep.lzd.co/mobapi_lib/handler"
	"godep.lzd.co/service/logger"
	"godep.lzd.co/mobapi_lib/utils"

	"motify_core_api/resources/service_error"
	"motify_core_api/utils/monitoring"
)

type MotifyCoreAPI struct {
	*MotifyCoreAPIGoRPC
}

func NewMotifyCoreAPIClient(srv *service.Service, apiTimeout time.Duration) *MotifyCoreAPI {
	serviceName := "motify_core_api"
	client := &http.Client{
		Transport: &http.Transport{
			//DisableCompression: true,
			MaxIdleConnsPerHost: 20,
		},
		Timeout: apiTimeout,
	}
	cb := Callbacks{
		OnPrepareRequest: func(ctx context.Context, req *http.Request, data interface{}) context.Context {
			tracking.SetRequestHeaders(ctx, req.Header)
			if s, e := ctxmanager.GetLoggerSession(ctx); e == nil {
				serviceSession := s.NewSession(serviceName, handler.CurlFromRequest(req))
				ctx = ctxmanager.NewContext(ctx, &ctxmanager.Context{
					SessionLogger: serviceSession,
				})
			}
			ctx = monitoring.SetMonitorData(ctx, serviceName, utils.GetPath(req))
			monitoring.FromContext(ctx).TimeStart = time.Now()

			if m, e := ctxmanager.GetSessionMocker(ctx); e == nil {
				sessionMocker := m.NewMocker(serviceName)
				sessionMocker.ExternalRequestParams(data)
				ctx = ctxmanager.NewContext(ctx, &ctxmanager.Context{
					SessionMocker: sessionMocker,
				})
			}
			return ctx
		},
		OnResponseUnmarshaling: func(ctx context.Context, req *http.Request, response *http.Response, result []byte) {
			if sessionMocker, err := ctxmanager.GetSessionMocker(ctx); err == nil {
				sessionMocker.ExternalRequest(req, response.StatusCode, result)
			} else {
				logger.Warning(ctx, err.Error())
			}
		},
		OnSuccess: func(ctx context.Context, req *http.Request, data interface{}) {
			if s, e := ctxmanager.GetLoggerSession(ctx); e == nil {
				s.Finish(data)
			}
			monitoring.MonitorTimeResponse(ctx, 200)
		},
		OnError: func(ctx context.Context, req *http.Request, err error) error {
			if s, e := ctxmanager.GetLoggerSession(ctx); e == nil {
				s.Error(err)
			}
			if _, ok := err.(*ServiceError); !ok {
				monitoring.MonitorTimeResponse(ctx, 500)
			}
			return service_error.AddServiceName(err, serviceName)
		},
		OnPanic: func(ctx context.Context, req *http.Request, r interface{}, trace []byte) error {
			if s, e := ctxmanager.GetLoggerSession(ctx); e == nil {
				s.Error(trace)
			}
			monitoring.MonitorTimeResponse(ctx, 500)
			return service_error.AddServiceName(fmt.Errorf("panic while calling service: %v", r), serviceName)
		},
	}
	coreApi := MotifyCoreAPI{
		NewMotifyCoreAPIGoRPC(client, srv.AddExternalService(serviceName), cb).
			SetCache(inmem.NewFromFlags(serviceName)),
	}
	return &coreApi
}
