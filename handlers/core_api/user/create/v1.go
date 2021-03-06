package user_create

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"

	"motify_core_api/godep_libs/service/logger"
	"motify_core_api/models"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	IntegrationFK *uint64 `key:"fk_integration" description:"Integration ID"`
	Name          *string `key:"name" description:"Name"`
	Short         *string `key:"p_description" description:"Short description"`
	Description   *string `key:"description" description:"Long Description"`
	Avatar        *string `key:"avatar" description:"Avatar url"`
	Phone         *string `key:"phone" description:"Phone number"`
	Email         *string `key:"email" description:"Email"`
	Password      string  `key:"password" description:"Password"`
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
	MISSED_REQUIRED_FIELDS error `text:"Missed required fields. You should set 'phone' or 'email'"`
	USER_EXISTS            error `text:"user exists"`
	CREATE_FAILED          error `text:"creating user is failed"`
	USER_NOT_CREATED       error `text:"created user not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Create/V1")
	cache.DisableTransportCache(ctx)

	newUser := &models.User{
		IntegrationFK: opts.IntegrationFK,
	}
	if opts.Email != nil && *opts.Email != "" {
		isBusy, err := handler.userService.IsLoginBusy(ctx, *opts.Email)
		if err != nil || isBusy {
			logger.Error(ctx, "User exists: %v, err: %v", isBusy, err)
			return nil, v1Errors.USER_EXISTS
		}
		newUser.Email = *opts.Email
	}
	if opts.Phone != nil && *opts.Phone != "" {
		isBusy, err := handler.userService.IsLoginBusy(ctx, *opts.Phone)
		if err != nil || isBusy {
			logger.Error(ctx, "User exists: %v, err: %v", isBusy, err)
			return nil, v1Errors.USER_EXISTS
		}
		newUser.Phone = *opts.Phone
	}
	if newUser.Email == "" && newUser.Phone == "" {
		return nil, v1Errors.MISSED_REQUIRED_FIELDS
	}

	if opts.Name != nil && *opts.Name != "" {
		newUser.Name = *opts.Name
	}
	if opts.Short != nil && *opts.Short != "" {
		newUser.Short = *opts.Short
	}
	if opts.Description != nil && *opts.Description != "" {
		newUser.Description = *opts.Description
	}
	if opts.Avatar != nil && *opts.Avatar != "" {
		newUser.Avatar = *opts.Avatar
	}
	userID, err := handler.userService.SetUser(ctx, newUser)
	if err != nil {
		logger.Error(ctx, "Failed creating user: %v", err)
		return nil, v1Errors.CREATE_FAILED
	}
	if userID == 0 {
		logger.Error(ctx, "Failed creating user: userID is 0")
		return nil, v1Errors.CREATE_FAILED
	}

	user, err := handler.userService.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.USER_NOT_CREATED
	}
	if user == nil {
		logger.Error(ctx, "Failed login: user is nil")
		return nil, v1Errors.USER_NOT_CREATED
	}

	if user.Email != "" {
		if err := handler.addUserAccess(ctx, user.Email, opts.Password, models.UserAccessEmail, user.ID, opts.IntegrationFK); err != nil {
			logger.Error(ctx, "Failed creating user access: %v", err)
			if err := handler.userService.DeleteUser(ctx, user.ID); err != nil {
				logger.Error(ctx, "Failed creating user: failed deleting user")
			}
			return nil, v1Errors.CREATE_FAILED
		}
	}
	if user.Phone != "" {
		if err := handler.addUserAccess(ctx, user.Phone, opts.Password, models.UserAccessPhone, user.ID, opts.IntegrationFK); err != nil {
			logger.Error(ctx, "Failed creating user access: %v", err)
			if err := handler.userService.DeleteUser(ctx, user.ID); err != nil {
				logger.Error(ctx, "Failed creating user: failed deleting user")
			}
			return nil, v1Errors.CREATE_FAILED
		}
	}

	status := "Email not sended"
	if opts.IntegrationFK != nil && *opts.IntegrationFK > 0 { // only for integration_api user-agents
		magicCode := wrapToken.NewApproveUser(user.ID, *opts.IntegrationFK).String()
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

func (handler *Handler) addUserAccess(ctx context.Context, login, password, field string, userID uint64, integrationFK *uint64) error {
	newUserAccess := &models.UserAccess{
		IntegrationFK: integrationFK,
		UserFK:        userID,
		Type:          field,
	}
	newUserAccess.SetLogin(login)
	newUserAccess.SetPassword(password)
	_, err := handler.userService.SetUserAccess(ctx, newUserAccess)
	return err
}
