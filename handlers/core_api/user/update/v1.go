package user_update

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	ID            uint64  `key:"id_user" description:"User ID"`
	IntegrationFK *uint64 `key:"fk_integration" description:"Integration ID"`
	Name          *string `key:"name" description:"Name"`
	Short         *string `key:"p_description" description:"Short description"`
	Description   *string `key:"description" description:"Long Description"`
	Avatar        *string `key:"avatar" description:"Avatar url"`
	Phone         *string `key:"phone" description:"Phone number"`
	Email         *string `key:"email" description:"Email"`
	PhoneApproved *bool   `key:"phone_approved" description:"Is phone approved"`
	EmailApproved *bool   `key:"email_approved" description:"Is email approved"`
	Password      *string `key:"password" description:"Password"`
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
	Avatar        string  `json:"avatar"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

type V1ErrorTypes struct {
	USER_NOT_FOUND             error `text:"user not found"`
	UPDATE_FAILED              error `text:"updating user is failed"`
	NEW_EMAIL_OR_PHONE_IS_BUSY error `text:"new email or phone is busy"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Create/V1")
	cache.DisableTransportCache(ctx)

	needUpdate := false
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

		accessList, err := handler.userService.GetUserAssessListByUserID(ctx, user.ID)
		if err != nil {
			logger.Error(ctx, "Fail loading user_access: %v", err)
			return nil, v1Errors.USER_NOT_FOUND
		}

		if err := handler.saveUserAccess(ctx, user.ID, opts.IntegrationFK, opts.Email, opts.Password, user.Email, models.UserAccessEmail, accessList[models.UserAccessEmail], accessList[models.UserAccessPhone]); err != nil {
			return nil, err
		}
		if err := handler.saveUserAccess(ctx, user.ID, opts.IntegrationFK, opts.Phone, opts.Password, user.Phone, models.UserAccessPhone, accessList[models.UserAccessPhone], accessList[models.UserAccessEmail]); err != nil {
			return nil, err
		}
	}

	needUpdate = false
	if opts.IntegrationFK != nil && (user.IntegrationFK == nil || *opts.IntegrationFK != *user.IntegrationFK) {
		needUpdate = true
		user.IntegrationFK = opts.IntegrationFK
	}
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
	if opts.Avatar != nil && *opts.Avatar != user.Avatar {
		needUpdate = true
		user.Avatar = *opts.Avatar
	}
	if opts.Phone != nil && *opts.Phone != user.Phone {
		needUpdate = true
		user.Phone = *opts.Phone
	}
	if opts.Email != nil && *opts.Email != user.Email {
		needUpdate = true
		user.Email = *opts.Email
	}
	if opts.Phone == nil && opts.PhoneApproved != nil && *opts.PhoneApproved != user.PhoneApproved {
		needUpdate = true
		user.PhoneApproved = *opts.PhoneApproved
	}
	if opts.Email == nil && opts.EmailApproved != nil && *opts.EmailApproved != user.EmailApproved {
		needUpdate = true
		user.EmailApproved = *opts.EmailApproved
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

func (handler *Handler) saveUserAccess(ctx context.Context, userID uint64, integrationFK *uint64, newLogin, newPass *string, oldLogin, field string, access *models.UserAccess, accessPair *models.UserAccess) error {
	needUpdate := false
	if access != nil {
		if newLogin != nil && *newLogin != "" && *newLogin != oldLogin {
			if isBusy, err := handler.userService.IsLoginBusy(ctx, *newLogin); err != nil || isBusy {
				logger.Error(ctx, "User exists: %v, err: %v", isBusy, err)
				return v1Errors.NEW_EMAIL_OR_PHONE_IS_BUSY
			}
			needUpdate = true
			access.SetLogin(*newLogin)
		}
	} else if (newLogin != nil && *newLogin != "") || oldLogin != "" {
		needUpdate = true
		access = &models.UserAccess{
			IntegrationFK: integrationFK,
			UserFK:        userID,
			Type:          field,
		}
		if newLogin != nil && *newLogin != "" {
			access.SetLogin(*newLogin)
		} else {
			access.SetLogin(oldLogin)
		}
		if accessPair != nil {
			access.Password = accessPair.Password
		}
	}
	if access != nil {
		if newPass != nil && *newPass != "" {
			needUpdate = true
			access.SetPassword(*newPass)
		}
		if access.Password == "" {
			logger.Error(ctx, "User add access: no pass")
			return v1Errors.UPDATE_FAILED
		}
		if needUpdate {
			logger.Debug(ctx, "update user_access: %#v", access)
			if _, err := handler.userService.SetUserAccess(ctx, access); err != nil {
				logger.Error(ctx, "Failed updating user access: %v", err)
				return v1Errors.UPDATE_FAILED
			}
		}
	}
	return nil
}
