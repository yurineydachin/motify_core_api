package user_login

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Hash     string `key:"integration_hash" description:"Hash for checking integration"`
	Login    string `key:"login" description:"Email or phone"`
	Password string `key:"password" description:"Password"`
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
	Awatar      string `json:"awatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type V1ErrorTypes struct {
	INTEGRATION_NOT_FOUND  error `text:"Integraion not found by hash"`
	LOGIN_FAILED           error `text:"Login is failed"`
	USER_NOT_FOUND         error `text:"User not found"`
	USER_ALREADY_LOGGED_IN error `text:"Request with already authorized apiToken"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.INullToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Login/V1")
	cache.DisableTransportCache(ctx)
	if apiToken != nil && !apiToken.IsGuest() {
		return nil, v1Errors.USER_ALREADY_LOGGED_IN
	}

	intData, err := handler.coreApi.IntegrationCheckV1(ctx, coreApiAdapter.IntegrationCheckV1Args{
		Hash: opts.Hash,
	})
	if err != nil {
		if err.Error() == "MotifyCoreAPI: INTEGRATION_NOT_FOUND" {
			return nil, v1Errors.INTEGRATION_NOT_FOUND
		}
		return nil, err
	}
	if intData == nil || intData.Integration == nil {
		return nil, v1Errors.INTEGRATION_NOT_FOUND
	}

	loginData, err := handler.coreApi.UserLoginV1(ctx, coreApiAdapter.UserLoginV1Args{
		IntegrationFK: &intData.Integration.ID,
		Login:         opts.Login,
		Password:      opts.Password,
	})
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.USER_NOT_FOUND
	}
	if loginData == nil || loginData.User == nil {
		logger.Error(ctx, "Failed login: user is nil")
		return nil, v1Errors.USER_NOT_FOUND
	}
	user := loginData.User
	if user.IntegrationFK == nil || intData.Integration.ID != *user.IntegrationFK {
		return nil, v1Errors.USER_NOT_FOUND
	}

	return &V1Res{
		Token: wrapToken.NewAgentUser(user.ID, *user.IntegrationFK).String(),
		User: &User{
			Hash:        wrapToken.NewAgentUser(user.ID, *user.IntegrationFK).Fixed().String(),
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Awatar:      user.Awatar,
			Phone:       user.Phone,
			Email:       user.Email,
		},
	}, nil
}
