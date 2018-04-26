package user_login

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	IntegrationFK *uint64 `key:"fk_integration" description:"Integration ID"`
	Login         string  `key:"login" description:"Email or phone"`
	Password      string  `key:"password" description:"Password"`
}

type V1Res struct {
	User *User `json:"user" description:"User if success"`
}

type User struct {
	ID            uint64  `json:"id_user"`
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          string  `json:"name"`
	Short         string  `json:"p_description"`
	Description   string  `json:"description"`
	Awatar        string  `json:"awatar"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
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

	login := opts.Login + models.LoginSufix(opts.IntegrationFK)
	userID, err := handler.userService.Authentificate(ctx, login, opts.Password)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.LOGIN_FAILED
	}
	if userID == 0 {
		logger.Error(ctx, "Failed login: userID = 0")
		return nil, v1Errors.USER_NOT_FOUND
	}
	user, err := handler.userService.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.USER_NOT_FOUND
	}
	if user == nil {
		logger.Error(ctx, "Failed login: user is nil")
		return nil, v1Errors.USER_NOT_FOUND
	}
	return &V1Res{
		User: &User{
			ID:            user.ID,
			IntegrationFK: user.IntegrationFK,
			Name:          user.Name,
			Short:         user.Short,
			Description:   user.Description,
			Awatar:        user.Awatar,
			Phone:         user.Phone,
			Email:         user.Email,
			UpdatedAt:     user.UpdatedAt,
			CreatedAt:     user.CreatedAt,
		},
	}, nil
}
