package user_remind

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/models"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	IntegrationFK *uint64 `key:"fk_integration" description:"Integration ID"`
	Login         string  `key:"login" description:"Email or phone"`
}

type V1Res struct {
	Token string `json:"token" description:"Token to remind pass"`
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
	EMAIL_NOT_SENDED   error `text:"Email not sended"`
	USER_NOT_FOUND error `text:"User not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Remind/V1")
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
		logger.Error(ctx, "User not found: %v", err)
		return nil, v1Errors.USER_NOT_FOUND
	}
	if user == nil {
		logger.Error(ctx, "User not found: user is nil")
		return nil, v1Errors.USER_NOT_FOUND
	}

    integrationID := uint64(0)
    if opts.IntegrationFK != nil && *opts.IntegrationFK > 0 {
        integrationID = *opts.IntegrationFK
    }
	magicCode := wrapToken.NewRemindUser(userID, integrationID).String()
	if email != "" && handler.emailFrom != "" {
		err = handler.emailService.SendEmployeeInvite(ctx, email, handler.emailFrom, code)
		if err != nil {
			logger.Error(ctx, "Error sending email: %v", err)
			status = "Error sending email"
		} else {
			status = "OK"
		}
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
