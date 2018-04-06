package user_signup

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Hash        string  `key:"integration_hash" description:"Hash for integration"`
	Name        *string `key:"name" description:"Name"`
	Short       *string `key:"p_description" description:"Short description"`
	Description *string `key:"description" description:"Long Description"`
	Awatar      *string `key:"awatar" description:"Awatar url"`
	Phone       *string `key:"phone" description:"Phone number"`
	Email       *string `key:"email" description:"Email"`
	Password    string  `key:"password" description:"Password"`
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
	INTEGRATION_NOT_FOUND     error `text:"Integraion not found by hash"`
	MISSED_REQUIRED_FIELDS    error `text:"Missed required fields. You should set 'phone' or 'email'"`
	USER_CREATE_FAILED        error `text:"User creating failed"`
	USER_EMAIL_ALLREADY_EXIST error `text:"User with this email allready exist"`
	USER_ALREADY_LOGGED_IN    error `text:"Request with already authorized apiToken"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.INullToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Signup/V1")
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

	coreOpts := coreApiAdapter.UserCreateV1Args{
		IntegrationFK: &intData.Integration.ID,
		Name:          opts.Name,
		Short:         opts.Short,
		Description:   opts.Description,
		Awatar:        opts.Awatar,
		Phone:         opts.Phone,
		Email:         opts.Email,
		Password:      opts.Password,
	}

	createData, err := handler.coreApi.UserCreateV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: MISSED_REQUIRED_FIELDS" {
			return nil, v1Errors.MISSED_REQUIRED_FIELDS
		} else if err.Error() == "MotifyCoreAPI: USER_EXISTS" {
			return nil, v1Errors.USER_EMAIL_ALLREADY_EXIST
		} else if err.Error() == "MotifyCoreAPI: CREATE_FAILED" {
			return nil, v1Errors.USER_CREATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: USER_NOT_CREATED" {
			return nil, v1Errors.USER_CREATE_FAILED
		}
		return nil, err
	}
	if createData.User == nil || createData.User.IntegrationFK == nil || *createData.User.IntegrationFK != intData.Integration.ID {
		return nil, v1Errors.USER_CREATE_FAILED
	}

	user := createData.User
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
