package user_create

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	Name        *string `key:"name" description:"Name"`
	Short       *string `key:"p_description" description:"Short description"`
	Description *string `key:"description" description:"Long Description"`
	Awatar      *string `key:"awatar" description:"Awatar url"`
	Phone       *string `key:"phone" description:"Phone number"`
	Email       *string `key:"email" description:"Email"`
	Password    string  `key:"password" description:"Password"`
}

type V1Res struct {
	User *User `json:"user" description:"User if success"`
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

	newUser := &models.User{}
	if opts.Email != nil && *opts.Email != "" {
		isBusy, err := handler.userService.IsEmailOrPhoneBusy(ctx, *opts.Email)
		if err != nil || isBusy {
			logger.Error(ctx, "User exists: %v, err: %v", isBusy, err)
			return nil, v1Errors.USER_EXISTS
		}
		newUser.Email = *opts.Email
	}
	if opts.Phone != nil && *opts.Phone != "" {
		isBusy, err := handler.userService.IsEmailOrPhoneBusy(ctx, *opts.Phone)
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
	if opts.Awatar != nil && *opts.Awatar != "" {
		newUser.Awatar = *opts.Awatar
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

	newUserAccess := &models.UserAccess{
		UserFK:   user.ID,
		Type:     models.UserAccessEmail,
		Email:    &user.Email,
		Phone:    &user.Phone,
		Password: opts.Password,
	}
	userAccessID, err := handler.userService.SetUserAccess(ctx, newUserAccess)
	if err != nil {
		logger.Error(ctx, "Failed creating user access: %v", err)
		if err := handler.userService.DeleteUser(ctx, user.ID); err != nil {
			logger.Error(ctx, "Failed creating user: failed deleting user")
		}
		return nil, v1Errors.CREATE_FAILED
	}
	if userAccessID == 0 {
		logger.Error(ctx, "Failed creating user access: userAccessID is 0")
		if err := handler.userService.DeleteUser(ctx, user.ID); err != nil {
			logger.Error(ctx, "Failed creating user: failed deleting user")
		}
		return nil, v1Errors.CREATE_FAILED
	}

	return &V1Res{
		User: &User{
			ID:          user.ID,
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Awatar:      user.Awatar,
			Phone:       user.Phone,
			Email:       user.Email,
			UpdatedAt:   user.UpdatedAt,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}
