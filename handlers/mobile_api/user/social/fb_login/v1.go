package user_fb_login

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/utils/oauth2"
	//coreApiAdapter "motify_core_api/resources/motify_core_api"
	//wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Code    *string `key:"code" description:"code"`
	FBToken    *string `key:"token" description:"FB token"`
	State    *string `key:"state" description:"state"`
	//Login    string `key:"login" description:"Email or phone"`
	//Password string `key:"password" description:"Password"`
}

type V1Res struct {
	FBUrl       string `json:"url" description:"url"`
	Code    *string `json:"code" description:"code"`
	State    *string `json:"state" description:"state"`
    FBProfile *FBProfile `json:"profile" description:"r"`
    Error string `json:"error" description:"error"`
/*
	Token       string `json:"token" description:"Authorized token"`
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
*/
}
type FBProfile struct {
	AccessToken    string `json:"access_token" description:"token"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type V1ErrorTypes struct {
	LOGIN_FAILED           error `text:"Login is failed"`
	USER_NOT_FOUND         error `text:"User not found"`
	USER_ALREADY_LOGGED_IN error `text:"Request with already authorized apiToken"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.INullToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Social/FBLogin/V1")
	cache.DisableTransportCache(ctx)
	if apiToken != nil && !apiToken.IsGuest() {
		return nil, v1Errors.USER_ALREADY_LOGGED_IN
	}

    res := V1Res{
        Code: opts.Code,
        State: opts.State,
    }
    res.FBUrl, _ = oauth2.GetCodeUrl()

    if opts.Code != nil || opts.FBToken != nil {
        profile, err := oauth2.GetFBUser(opts.Code, opts.FBToken)
        if err != nil {
            res.Error = err.Error()
        } else if profile != nil {
            res.FBProfile = &FBProfile{
                ID: profile.ID,
                Name: profile.Name,
                AccessToken: profile.AccessToken,
            }
        } else {
            res.Error = "No profile from oauth2 module"
        }
    } else {
        res.Error = "No code or access token"
    }

/*
	loginData, err := handler.coreApi.UserLoginV1(ctx, coreApiAdapter.UserLoginV1Args{
		Login:    opts.Login,
		Password: opts.Password,
	})
	if err != nil {
		if err.Error() == "MotifyCoreAPI: LOGIN_FAILED" {
			return nil, v1Errors.LOGIN_FAILED
		}
		logger.Error(ctx, "Failed login: %v", err)
		return nil, err
	}
	if loginData == nil || loginData.User == nil {
		logger.Error(ctx, "Failed login: user is nil")
		return nil, v1Errors.USER_NOT_FOUND
	}
	user := loginData.User
*/

    return &res, nil
        /*
	return &V1Res{
	FBUrl: Url.String(),
    Code: *opts.Code,
    State: *opts.State,
		Token:       wrapToken.NewMobileUser(user.ID).String(),
		Hash:        wrapToken.NewMobileUser(user.ID).Fixed().String(),
		Name:        user.Name,
		Short:       user.Short,
		Description: user.Description,
		Avatar:      user.Avatar,
		Phone:       user.Phone,
		Email:       user.Email,
	}, nil
        */
}
