package user_device_ios

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type V1Args struct {
	Token string `key:"token" description:"device token for push notification"`
}

type V1Res struct {
	Status string `json:"status" description:"status"`
}
type V1ErrorTypes struct {
	FAILED_ADDING_TOKEN error `text:"Failed adding device token"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Register/Device/IOS/V1")
	cache.DisableTransportCache(ctx)

	data, err := handler.coreApi.UserDeviceV1(ctx, coreApiAdapter.UserDeviceV1Args{
		ID:     uint64(apiToken.GetID()),
		Device: "ios",
		Token:  opts.Token,
	})
	if err != nil {
		if err.Error() == "MotifyCoreAPI: FAILED_ADDING_TOKEN" {
			return nil, v1Errors.FAILED_ADDING_TOKEN
		}
		logger.Error(ctx, "Failed add device token: %v", err)
		return nil, err
	}
	if data == nil || data.Token != opts.Token {
		logger.Error(ctx, "Failed add device token: %v", err)
		return nil, v1Errors.FAILED_ADDING_TOKEN
	}

	return &V1Res{
		Status: "OK",
	}, nil
}
