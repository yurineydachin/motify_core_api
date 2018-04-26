package context

import (
	"context"
	"errors"
	"net/http"

	"motify_core_api/godep_libs/mobapi_lib/logger"
	"motify_core_api/godep_libs/mobapi_lib/sessionlogger"
	"motify_core_api/godep_libs/mobapi_lib/sessionmocker"
)

type key int

var keyContext key = 0

var (
	ErrorNoLoggerInContext    = errors.New("There's no session logger in context. Maybe this feature is disabled for now.")
	ErrorNoMockerInContext    = errors.New("There's no session mocker in context. Maybe this feature is disabled for now.")
	ErrorNoValue              = errors.New("Value not found in context. It looks like context is nil or has not been initialized yet.")
	ErrorUnexpectedDataFormat = errors.New("Context data type has not expected format")
)

type Context struct {
	ReqClientId      string
	ReqTokenHeader   string
	ReqUserAgent     string
	ReqRemoteAddr    string
	ReqURI           string
	ReqLang          string
	ReqMobAppVersion string
	RespCookies      []*http.Cookie
	SessionLogger    *sessionlogger.Session
	SessionMocker    *sessionmocker.Mocker
	WebPSupported    bool
	ScreenWidth      int
	ScreenHeight     int
	ScreenDensity    float64

	dictForErrorTranslation dict
}

type dict interface {
	Translate(in string, argsH map[string]interface{}, argsA []interface{}) (out string, ok bool)
}

func (ctx *Context) SetErrorTranslatorDictionary(d dict) {
	ctx.dictForErrorTranslation = d
}

func (ctx *Context) GetErrorTranslatorDictionary() dict {
	return ctx.dictForErrorTranslation
}

func FromContext(ctx context.Context) (*Context, error) {
	var val interface{}
	if ctx != nil {
		val = ctx.Value(keyContext)
	}
	if val == nil {
		return &Context{}, ErrorNoValue
	}
	ctxData, ok := val.(*Context)
	if !ok {
		return ctxData, ErrorUnexpectedDataFormat
	}
	return ctxData, nil
}

func NewContext(ctx context.Context, data *Context) context.Context {
	if ctx == nil {
		logger.Warning(nil, "Context is nil")
		ctx = context.Background()
	}

	oldContext, err := FromContext(ctx)
	if err == nil {
		if data.SessionLogger == nil && oldContext.SessionLogger != nil {
			data.SessionLogger = oldContext.SessionLogger
		}
		if data.SessionMocker == nil && oldContext.SessionMocker != nil {
			data.SessionMocker = oldContext.SessionMocker
		}
		if len(oldContext.RespCookies) > 0 {
			data.RespCookies = append(data.RespCookies, oldContext.RespCookies...)
		}
	}

	return context.WithValue(ctx, keyContext, data)
}

// GetLoggerSession returns logger session stored in context
func GetLoggerSession(ctx context.Context) (*sessionlogger.Session, error) {
	ctxData, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if ctxData.SessionLogger == nil {
		return sessionlogger.NewBlackholeSession(), ErrorNoLoggerInContext
	}
	return ctxData.SessionLogger, nil
}

func NewLoggerSession(ctx context.Context, caption string, request interface{}) (*sessionlogger.Session, error) {
	// prepare logger session from context
	session, err := GetLoggerSession(ctx)
	if err != nil {
		return nil, err
	}
	return session.NewSession(caption, request), nil
}

func GetSessionMocker(ctx context.Context) (*sessionmocker.Mocker, error) {
	ctxData, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if ctxData.SessionMocker == nil {
		return nil, ErrorNoMockerInContext
	}
	return ctxData.SessionMocker, nil
}
