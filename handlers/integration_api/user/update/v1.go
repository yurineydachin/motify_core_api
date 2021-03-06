package user_update

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Name        *string `key:"name" description:"Name"`
	Short       *string `key:"p_description" description:"Short description"`
	Description *string `key:"description" description:"Long Description"`
	Avatar      *string `key:"avatar" description:"Avatar url"`
	Phone       *string `key:"phone" description:"Phone number"`
	Email       *string `key:"email" description:"Email"`
	Password    *string `key:"password" description:"Password"`
}

type V1Res struct {
	User *User `json:"user" description:"User if success"`
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
	USER_UPDATE_FAILED        error `text:"User creating failed"`
	USER_EMAIL_ALLREADY_EXIST error `text:"User with this email allready exist"`
	USER_PHONE_ALLREADY_EXIST error `text:"User with this phone allready exist"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Update/V1")
	cache.DisableTransportCache(ctx)

	integrationID := uint64(apiToken.GetExtraID())
	coreOpts := coreApiAdapter.UserUpdateV1Args{
		ID:            uint64(apiToken.GetID()),
		IntegrationFK: &integrationID,
		Name:          opts.Name,
		Short:         opts.Short,
		Description:   opts.Description,
		Avatar:        opts.Avatar,
		Phone:         opts.Phone,
		Email:         opts.Email,
		Password:      opts.Password,
	}

	createData, err := handler.coreApi.UserUpdateV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: NEW_EMAIL_IS_BUSY" {
			return nil, v1Errors.USER_EMAIL_ALLREADY_EXIST
		} else if err.Error() == "MotifyCoreAPI: NEW_PHONE_IS_BUSY" {
			return nil, v1Errors.USER_PHONE_ALLREADY_EXIST
		} else if err.Error() == "MotifyCoreAPI: USER_NOT_FOUND" {
			return nil, v1Errors.USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: UPDATE_FAILED" {
			return nil, v1Errors.USER_UPDATE_FAILED
		}
		return nil, err
	}

	user := createData.User
	if createData.User == nil || user.IntegrationFK == nil || *user.IntegrationFK != integrationID {
		return nil, v1Errors.USER_UPDATE_FAILED
	}

	return &V1Res{
		User: &User{
			Hash:        wrapToken.NewAgentUser(user.ID, *user.IntegrationFK).Fixed().String(),
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Avatar:      user.Avatar,
			Phone:       user.Phone,
			Email:       user.Email,
		},
	}, nil
}
