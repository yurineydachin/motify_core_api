package user_approve_send

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"

	"motify_core_api/godep_libs/service/logger"
	"motify_core_api/models"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	IntegrationFK *uint64 `key:"fk_integration" description:"Integration ID"`
	Login         string  `key:"login" description:"Login"`
}

type V1Res struct {
	Result string `json:"result" description:"Result status"`
	User   *User  `json:"user" description:"User if success"`
}

type User struct {
	ID            uint64  `json:"id_user"`
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          string  `json:"name"`
	Short         string  `json:"p_description"`
	Description   string  `json:"description"`
	Avatar        string  `json:"avatar"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

type V1ErrorTypes struct {
	USER_NOT_FOUND error `text:"user not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Create/V1")
	cache.DisableTransportCache(ctx)

	login := opts.Login + models.LoginSufix(opts.IntegrationFK)
	userID, err := handler.userService.GetUserIDByLogin(ctx, login)
	if err != nil {
		logger.Error(ctx, "User not found: %v", err)
		return nil, v1Errors.USER_NOT_FOUND
	}
	if userID == 0 {
		logger.Error(ctx, "User not found: userID = 0")
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

	status := "Email not sended"
	if opts.IntegrationFK != nil && *opts.IntegrationFK > 0 {
		magicCode := wrapToken.NewApproveUser(userID, *opts.IntegrationFK).String()
		if user.Email != "" && handler.emailFrom != "" {
			err = handler.emailService.UserApprove(ctx, user.Email, handler.emailFrom, magicCode)
			if err != nil {
				logger.Error(ctx, "Error sending email: %v", err)
				status = "Error sending email"
			} else {
				status = "OK"
			}
		}
	}

	return &V1Res{
		Result: status,
		User: &User{
			ID:            user.ID,
			IntegrationFK: user.IntegrationFK,
			Name:          user.Name,
			Short:         user.Short,
			Description:   user.Description,
			Avatar:        user.Avatar,
			Phone:         user.Phone,
			Email:         user.Email,
			UpdatedAt:     user.UpdatedAt,
			CreatedAt:     user.CreatedAt,
		},
	}, nil
}
