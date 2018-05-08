package user_remind_send

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type V1Args struct {
	Login string `key:"login" description:"Email or phone"`
}

type V1Res struct {
	Result string `json:"result" description:"Result status"`
}

type V1ErrorTypes struct {
	EMAIL_NOT_SENDED       error `text:"Email not sended"`
	USER_NOT_FOUND         error `text:"User not found"`
	USER_ALREADY_LOGGED_IN error `text:"Request with already authorized apiToken"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.INullToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Remind/Send/V1")
	cache.DisableTransportCache(ctx)
	if apiToken != nil && !apiToken.IsGuest() {
		return nil, v1Errors.USER_ALREADY_LOGGED_IN
	}

	coreOpts := coreApiAdapter.UserRemindV1Args{
		Login: opts.Login,
	}

	remindData, err := handler.coreApi.UserRemindV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: EMAIL_NOT_SENDED" {
			return nil, v1Errors.EMAIL_NOT_SENDED
		} else if err.Error() == "MotifyCoreAPI: USER_NOT_FOUND" {
			return nil, v1Errors.USER_NOT_FOUND
		}
		return nil, err
	}
	if remindData.User == nil {
		return nil, v1Errors.USER_NOT_FOUND
	}

	return &V1Res{
		Result: remindData.Result,
	}, nil
}
