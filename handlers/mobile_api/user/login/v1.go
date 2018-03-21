package user_login

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type V1Args struct {
	Login    string `key:"login" description:"Email or phone"`
	Password string `key:"password" description:"Password"`
}

type V1Res struct {
	Token string `json:"token" description:"Authorized token"`
	User  *User  `json:"user" description:"User if success"`
}

type User struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Awatar      string `json:"awatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type V1ErrorTypes struct {
	LOGIN_FAILED   error `text:"Login is failed"`
	USER_NOT_FOUND error `text:"User not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Login/V1")
	cache.DisableTransportCache(ctx)

    loginData, err := handler.coreApi.UserLoginV1(ctx, coreApiAdapter.UserLoginV1Args{
        Login: opts.Login,
        Password: opts.Password,
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

	return &V1Res{
		Token: token.NewTokenV1(user.ID).String(),
		User: &User{
			ID:          user.ID,
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Awatar:      user.Awatar,
			Phone:       user.Phone,
			Email:       user.Email,
		},
	}, nil
}
