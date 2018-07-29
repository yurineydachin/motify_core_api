package handlersmanager

import (
	"reflect"

	"context"
	"fmt"
	"github.com/sergei-svistunov/gorpc"
	"motify_core_api/godep_libs/metrics"
	ctxmanager "motify_core_api/godep_libs/mobapi_lib/context"
	"motify_core_api/godep_libs/mobapi_lib/logger"
	"motify_core_api/godep_libs/mobapi_lib/token"
)

type HandlersManager struct {
	*gorpc.HandlersManager
}

func New(handlersPath string, tokenModel uint64) *HandlersManager {
	return &HandlersManager{
		HandlersManager: gorpc.NewHandlersManager(handlersPath, getHmCallbacks(tokenModel)),
	}
}

func (hm *HandlersManager) GetGoRPCHandlersManager() *gorpc.HandlersManager {
	return hm.HandlersManager
}

const (
	metricNameAccessWithoutToken    = "not_provided_in_request"
	metricNameTokenParsingFailed    = "parsing_problem"
	metricNameTokenParsedButInvalid = "parsed_but_invalid"
	metricNameGuestTokenIsNotEnough = "guest_token_is_not_enough"
)

var (
	registry      = metrics.DefaultRegistry()
	TokenCounters = registry.NewCounterVec("token_counters", "Some token counters.", "type")
)

func getHmCallbacks(tokenModel uint64) gorpc.HandlersManagerCallbacks {
	return gorpc.HandlersManagerCallbacks{
		// OnHandlerRegistration will be called only one time for each handler version while handler registration is in progress
		OnHandlerRegistration: func(path string, method reflect.Method) interface{} {
			return getTokenType(method)
		},

		// AppendInParams will be called for each handler call just right after input parameters will be prepared in Handler Manager
		AppendInParams: func(ctx context.Context, preparedParams []reflect.Value, extraData interface{}) (context.Context, []reflect.Value, error) {
			// fetch logger session from context
			ctxData, err := ctxmanager.FromContext(ctx)
			if err != nil {
				return ctx, nil, err
			}
			if ctxData.SessionLogger != nil {
				// create child logger session and replace it in context
				ctxData.SessionLogger = ctxData.SessionLogger.NewSession("handlers_manager", preparedParams[2].Interface())
				ctx = ctxmanager.NewContext(ctx, ctxData)
			}

			preparedParams[1] = reflect.ValueOf(ctx)

			// prepare token by input data and extraData (expected token type)
			apiToken, shouldAppendToInputs, err := PrepareToken(ctxData.ReqTokenHeader, extraData, tokenModel)
			if err != nil {
				if tokenErr, ok := err.(*gorpc.HandlerError); ok {
					switch tokenErr.UserMessage {
					case errMsgAccessWithoutRequiredToken:
						TokenCounters.WithLabelValues(metricNameAccessWithoutToken).Inc()
						logger.Error(ctx, "Authorized token is requeried but not provided for request '%s' by client '%s'. X-API-TOKEN header: %s", ctxData.ReqURI, ctxData.ReqUserAgent, ctxData.ReqTokenHeader)
					case errMsgTokenParsingFailed:
						TokenCounters.WithLabelValues(metricNameTokenParsingFailed).Inc()
						logger.Error(ctx, "Token parsing error: '%s' for request '%s' by client '%s'. X-API-TOKEN header: %s", err.Error(), ctxData.ReqURI, ctxData.ReqUserAgent, ctxData.ReqTokenHeader)
					case errMsgTokenDataParsedButInvalid:
						TokenCounters.WithLabelValues(metricNameTokenParsedButInvalid).Inc()
						logger.Error(ctx, "Token parsed but it's invalid for request '%s' by client '%s'. X-API-TOKEN header: %s", ctxData.ReqURI, ctxData.ReqUserAgent, ctxData.ReqTokenHeader)
					case errMsgGuestTokenProvidedButShouldBeAuthorizedToken:
						TokenCounters.WithLabelValues(metricNameGuestTokenIsNotEnough).Inc()
						logger.Error(ctx, "Authorized token is requeried but GUEST token has been provided in request '%s' by client '%s'. X-API-TOKEN header: %s", ctxData.ReqURI, ctxData.ReqUserAgent, ctxData.ReqTokenHeader)
					default:
						TokenCounters.WithLabelValues(metricNameTokenParsingFailed).Inc()
					}
				}
				return ctx, nil, err
			}

			if apiToken != nil {
				fields := make(map[string]string, 3)

				customerType := "guest"
				if !apiToken.IsGuest() {
					customerType = "customer"
					fields["id"] = fmt.Sprintf("%d", apiToken.GetID())
					fields["model"] = fmt.Sprintf("%d", apiToken.GetModel())
				}
				fields["customer-type"] = customerType

				ctxData.SessionMocker.SetFields(fields)
			}

			if apiToken == nil {
				// for INullToken type
				if shouldAppendToInputs {
					nullTokenValue := reflect.ValueOf(new(token.INullToken)).Elem()
					preparedParams = append(preparedParams, nullTokenValue)
				}
				return ctx, preparedParams, nil
			}

			val := reflect.ValueOf(apiToken)

			// append extra inputs
			preparedParams = append(preparedParams, val)

			return ctx, preparedParams, nil
		},

		// OnError will be called if any error occurs while CallHandler() method is in processing
		OnError: func(ctx context.Context, err error) {
			if err == nil {
				logger.Error(ctx, "OnError callback has been called but nil error has provided! Context: %v", ctx)
				return
			}
			loggerSession, sessionError := ctxmanager.GetLoggerSession(ctx)
			if sessionError == nil {
				loggerSession.Error(err)
			}
		},

		// OnSuccess will be called if CallHandler() method is successfully finished
		OnSuccess: func(ctx context.Context, result interface{}) {
			loggerSession, err := ctxmanager.GetLoggerSession(ctx)
			if err == nil {
				loggerSession.Finish(result)
			}
		},
	}
}
