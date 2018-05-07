package user_remind_reset

import (
	"context"
	"time"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

const (
	validMagicCode = 3600 // 1 hour
)

type V1Args struct {
	Code     string `key:"code" description:"Code from email"`
	Password string `key:"password" description:"New password"`
}

type V1Res struct {
	Token string `json:"token" description:"Authorized token"`
	User  *User  `json:"user" description:"User if success"`
}

type User struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type V1ErrorTypes struct {
	USER_UPDATE_FAILED     error `text:"User creating failed"`
	USER_NOT_FOUND         error `text:"User not found"`
	ERROR_PARSE_MAGIC_CODE error `text:"Error parse magic code"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.INullToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Update/V1")
	cache.DisableTransportCache(ctx)

	userToken, err := wrapToken.ParseRemindUser(opts.Code)
	if err != nil {
		logger.Error(ctx, "Error parse magic code: ", err)
		return nil, v1Errors.ERROR_PARSE_MAGIC_CODE
	} else if userToken.GetID() == 0 {
		logger.Error(ctx, "Error parse magic code: invalid user.id: %d", userToken.GetID())
		return nil, v1Errors.ERROR_PARSE_MAGIC_CODE
	} else if userToken.GetDate().Add(validMagicCode * time.Second).Before(time.Now()) {
		logger.Error(ctx, "Error parse magic code: invalid date: %s + %d sec < %s now", userToken.GetDate(), validMagicCode, time.Now())
		return nil, v1Errors.ERROR_PARSE_MAGIC_CODE
	}

	coreOpts := coreApiAdapter.UserUpdateV1Args{
		ID:       uint64(userToken.GetID()),
		Password: &opts.Password,
	}

	createData, err := handler.coreApi.UserUpdateV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: NEW_EMAIL_IS_BUSY" {
			return nil, v1Errors.USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: NEW_PHONE_IS_BUSY" {
			return nil, v1Errors.USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: USER_NOT_FOUND" {
			return nil, v1Errors.USER_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: UPDATE_FAILED" {
			return nil, v1Errors.USER_UPDATE_FAILED
		}
		return nil, err
	}
	if createData.User == nil {
		return nil, v1Errors.USER_NOT_FOUND
	}

	user := createData.User
	return &V1Res{
		Token: wrapToken.NewMobileUser(user.ID).String(),
		User: &User{
			Hash:        wrapToken.NewMobileUser(user.ID).Fixed().String(),
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Avatar:      user.Avatar,
			Phone:       user.Phone,
			Email:       user.Email,
		},
	}, nil
}
