package user_device

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	ID     uint64 `key:"id_user" description:"User ID"`
	Device string `key:"device" description:"ios,android"`
	Token  string `key:"token" description:"Device token for push notification"`
}

type V1Res struct {
	Token string `json:"token" description:"Device token for push notification"`
}

type V1ErrorTypes struct {
	FAILED_ADDING_TOKEN error `text:"Failed adding device token"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Device/V1")
	cache.DisableTransportCache(ctx)

	device := &models.Device{
		UserFK: opts.ID,
		Device: opts.Device,
		Token:  opts.Token,
	}
	if _, err := handler.deviceService.Set(ctx, device); err != nil {
		logger.Error(ctx, "Failed adding device (%s) token (%s) to user (%d): %v", opts.Device, opts.Token, opts.ID, err)
		return nil, v1Errors.FAILED_ADDING_TOKEN
	}
	return &V1Res{
		Token: opts.Token,
	}, nil
}
