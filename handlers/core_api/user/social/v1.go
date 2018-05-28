package user_social

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	IntegrationFK *uint64 `key:"fk_integration" description:"Integration ID"`
	UserID        *uint64 `key:"id_user" description:"User ID"`
	Social        string  `key:"social" description:"Social: FB or Google"`
	Login         string  `key:"login" description:"Email or phone"`
	Name          string  `key:"name" description:"name"`
	Avatar        *string `key:"avatar" description:"avatar"`
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
	LOGIN_FAILED                   error `text:"Login is failed"`
	USER_NOT_FOUND                 error `text:"User not found"`
	SOCIAL_USER_HAS_ALREADY_PINNED error `text:"User has already pinned to anouther account"`
	USER_EXISTS                    error `text:"user exists"`
	CREATE_FAILED                  error `text:"creating user is failed"`
	USER_NOT_CREATED               error `text:"created user not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "User/Social/V1")
	cache.DisableTransportCache(ctx)

	login := fmt.Sprintf("social:%s/%s%s", opts.Social, opts.Login, models.LoginSufix(opts.IntegrationFK))
	password := "pass_" + login

	userID, err := handler.getUserIDByLoginAndPass(ctx, login, password, opts.IntegrationFK)
	if err != nil {
		return nil, err
	}
	if opts.UserID != nil && *opts.UserID > 0 {
		if userID == 0 {
			if err := handler.addUserAccess(ctx, login, password, social, *opts.UserID, opts.IntegrationFK); err != nil {
				return 0, v1Errors.CREATE_FAILED
			}
		} else if userID != *opts.UserID {
			return nil, v1Errors.SOCIAL_USER_HAS_ALREADY_PINNED
		}
		userID = *opts.UserID
	} else if userID == 0 {
		userID, err = handler.createUser(ctx, login, password, social, opts.Name, opts.Avatar, opts.IntegrationFK)
	}

	if userID == 0 {
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
			Avatar:        user.Avatar,
			Phone:         user.Phone,
			Email:         user.Email,
			UpdatedAt:     user.UpdatedAt,
			CreatedAt:     user.CreatedAt,
		},
	}, nil
}

func (handler *Handler) getUserIDByLoginAndPass(ctx context.Context, login, password string) (uint64, error) {
	userID, err := handler.userService.Authentificate(ctx, login, password)
	if err != nil {
		logger.Error(ctx, "Failed social login: %v", err)
		return 0, v1Errors.LOGIN_FAILED
	}
	if userID == 0 {
		return 0, nil
	}
	return userID, nil
}

func (handler *Hanlder) createUser(ctx context.Context, login, password, social, name string, avatar *string, intergrationFK *uint64) (uint64, error) {
	newUser := &models.User{
		IntegrationFK: integrationFK,
		Name:          name,
	}
	if avatar != nil && *avatar != "" {
		newUser.Avatar = *avatar
	}

	userID, err := handler.userService.SetUser(ctx, newUser)
	if err != nil {
		logger.Error(ctx, "Failed creating user: %v", err)
		return 0, v1Errors.CREATE_FAILED
	}

	if err := handler.addUserAccess(ctx, login, password, social, userID, intergrationFK); err != nil {
		logger.Error(ctx, "Failed creating user access: %v", err)
		if err := handler.userService.DeleteUser(ctx, userID); err != nil {
			logger.Error(ctx, "Failed creating user: failed deleting user")
		}
		return 0, v1Errors.CREATE_FAILED
	}

	return userID, nil
}

func (handler *Handler) addUserAccess(ctx context.Context, login, password, social, userID uint64, intergrationFK *uint64) error {
	newUserAccess := &models.UserAccess{
		IntegrationFK: integrationFK,
		UserFK:        userID,
		Type:          social,
		Password:      password,
	}
	newUserAccess.SetLogin(login)
	_, err := handler.userService.SetUserAccess(ctx, newUserAccess)
	return err
}
