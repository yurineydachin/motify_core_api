package user_fb_login

import (
	"context"
	"strings"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	"motify_core_api/utils/oauth2"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Code    *string `key:"code" description:"code"`
	FBToken *string `key:"token" description:"FB token"`
}

type V1Res struct {
	Token string `json:"token" description:"Authorized token"`
	User  *User  `json:"user" description:"User"`
}
type User struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type V1ErrorTypes struct {
	MISSED_REQUIRED_FIELDS         error `text:"Need code or token"`
	SOCIAL_USER_HAS_ALREADY_PINNED error `text:"User has already pinned to anouther account"`
	LOGIN_FAILED                   error `text:"Login is failed"`
	USER_NOT_FOUND                 error `text:"User not found"`
	CODE_HAS_EXPIRED               error `text:"This authorization code has expired"`
	CODE_HAS_BEEN_USED             error `text:"This authorization code has been used"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.INullToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Social/FBLogin/V1")
	cache.DisableTransportCache(ctx)
	userID := uint64(0)
	if apiToken != nil && !apiToken.IsGuest() {
		userID = uint64(apiToken.GetID())
	}

	FBUrl, _ := oauth2.GetCodeUrl(oauth2.FBConf)
	logger.Error(ctx, "FB url for code: %s", FBUrl)

	if opts.Code == nil && opts.FBToken == nil {
		return nil, v1Errors.MISSED_REQUIRED_FIELDS
	}

	oauthErrors := []error{
		v1Errors.CODE_HAS_EXPIRED,
		v1Errors.CODE_HAS_BEEN_USED,
	}

	profile, err := oauth2.GetFBUser(opts.Code, opts.FBToken)
	if err != nil {
		for _, e := range oauthErrors {
			if strings.Index(err.Error(), e.Error()) > 0 {
				return nil, e
			}
		}
		logger.Error(ctx, "Some error from oauth2: %v", err)
		return nil, err
	} else if profile == nil {
		logger.Error(ctx, "FB profile is nil")
		return nil, v1Errors.USER_NOT_FOUND
	}

	loginData, err := handler.coreApi.UserSocialV1(ctx, coreApiAdapter.UserSocialV1Args{
		UserID: &userID,
		Social: oauth2.FBConf,
		Login:  profile.ID,
		Name:   profile.Name,
	})
	if err != nil {
		if err.Error() == "MotifyCoreAPI: LOGIN_FAILED" {
			return nil, v1Errors.LOGIN_FAILED
		} else if err.Error() == "MotifyCoreAPI: SOCIAL_USER_HAS_ALREADY_PINNED" {
			return nil, v1Errors.SOCIAL_USER_HAS_ALREADY_PINNED
		}
		logger.Error(ctx, "Failed login: %v", err)
		return nil, err
	}
	if loginData == nil || loginData.User == nil {
		logger.Error(ctx, "Failed login: user is nil")
		return nil, v1Errors.USER_NOT_FOUND
	}
	user := loginData.User

	return &V1Res{
		Token: wrapToken.NewMobileUser(user.ID).String(),
		User: &User{
			Hash:        wrapToken.NewMobileUser(user.ID).Fixed().String(),
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Avatar:      user.Avatar,
			Phone:       user.Phone,
			Email:       user.Email,
		},
	}, nil
}
