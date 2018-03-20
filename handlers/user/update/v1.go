package user_update

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"
	"godep.lzd.co/mobapi_lib/token"

	"motify_core_api/models"
)

type V1Args struct {
	ID          uint64  `key:"id_user" description:"User ID"`
	Name        *string `key:"name" description:"Name"`
	Short       *string `key:"p_description" description:"Short description"`
	Description *string `key:"description" description:"Long Description"`
	Awatar      *string `key:"awatar" description:"Awatar url"`
	Phone       *string `key:"phone" description:"Phone number"`
	Email       *string `key:"email" description:"Email"`
	Password    *string `key:"password" description:"Password"`
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
	USER_NOT_FOUND    error `text:"user not found"`
	UPDATE_FAILED     error `text:"updating user is failed"`
	NEW_EMAIL_IS_BUSY error `text:"new email is busy"`
	NEW_PHONE_IS_BUSY error `text:"new phone is busy"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Create/V1")
	cache.DisableTransportCache(ctx)

	user, err := handler.userService.GetUserByID(ctx, opts.ID)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.USER_NOT_FOUND
	}
	if user == nil {
		logger.Error(ctx, "Failed login: user is nil")
		return nil, v1Errors.USER_NOT_FOUND
	}

	if (opts.Email != nil && *opts.Email != user.Email) ||
		(opts.Phone != nil && *opts.Phone != user.Phone) ||
		(opts.Password != nil && *opts.Password != "") {

		access, err := handler.userService.GetUserAssessByUserIDAndType(ctx, user.ID, models.UserAccessEmail)
		if err != nil {
			logger.Error(ctx, "Fail loading user_access: %v", err)
			return nil, v1Errors.USER_NOT_FOUND
		}
		if access == nil {
			access = &models.UserAccess{
				UserFK:   user.ID,
				Type:     models.UserAccessEmail,
				Email:    &user.Email,
				Phone:    &user.Phone,
				Password: *opts.Password,
			}
		} else {
			if opts.Email != nil && *opts.Email != user.Email {
				isBusy, err := handler.userService.IsEmailOrPhoneBusy(ctx, *opts.Email)
				if err != nil || isBusy {
					logger.Error(ctx, "User exists: %v, err: %v", isBusy, err)
					return nil, v1Errors.NEW_EMAIL_IS_BUSY
				}
				access.SetEmail(*opts.Email)
			}
			if opts.Phone != nil && *opts.Phone != user.Phone {
				isBusy, err := handler.userService.IsEmailOrPhoneBusy(ctx, *opts.Phone)
				if err != nil || isBusy {
					logger.Error(ctx, "User exists: %v, err: %v", isBusy, err)
					return nil, v1Errors.NEW_EMAIL_IS_BUSY
				}
				access.SetPhone(*opts.Phone)
			}
			if opts.Password != nil && *opts.Password != "" {
				access.SetPassword(*opts.Password)
			}
		}

		logger.Debug(ctx, "update user_access: %#v", access)
		userAccessID, err := handler.userService.SetUserAccess(ctx, access)
		if err != nil {
			logger.Error(ctx, "Failed updating user access: %v", err)
			return nil, v1Errors.UPDATE_FAILED
		}
		if userAccessID == 0 {
			logger.Error(ctx, "Failed updating user access: userAccessID is 0")
			return nil, v1Errors.UPDATE_FAILED
		}
	}

	needUpdate := false
	if opts.Name != nil && *opts.Name != user.Name {
		needUpdate = true
		user.Name = *opts.Name
	}
	if opts.Short != nil && *opts.Short != user.Short {
		needUpdate = true
		user.Short = *opts.Short
	}
	if opts.Description != nil && *opts.Description != user.Description {
		needUpdate = true
		user.Description = *opts.Description
	}
	if opts.Awatar != nil && *opts.Awatar != user.Awatar {
		needUpdate = true
		user.Awatar = *opts.Awatar
	}
	if opts.Phone != nil && *opts.Phone != user.Phone {
		needUpdate = true
		user.Phone = *opts.Phone
	}
	if opts.Email != nil && *opts.Email != user.Email {
		needUpdate = true
		user.Email = *opts.Email
	}

	if needUpdate {
		logger.Debug(ctx, "update user: %#v", user)
		userID, err := handler.userService.SetUser(ctx, user)
		if err != nil {
			logger.Error(ctx, "Failed updating user: %v", err)
			return nil, v1Errors.UPDATE_FAILED
		}
		if userID == 0 {
			logger.Error(ctx, "Failed updating user: userID is 0")
			return nil, v1Errors.UPDATE_FAILED
		}
	}

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
			UpdatedAt:   user.UpdatedAt,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}
