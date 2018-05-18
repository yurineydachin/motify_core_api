package user_update

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/resources/file_storage"
	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	File FileData `key:"file" description:"File"`
}

type FileData struct {
	Name   string `key:"name" description:"file name"`
	Base64 string `key:"base64" description:"base64"`
}

type V1Res struct {
	User *User `json:"user" description:"User"`
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
	FILE_NOT_LOADED    error `text:"File not loaded"`
	USER_UPDATE_FAILED error `text:"User creating failed"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Avatar/V1")
	cache.DisableTransportCache(ctx)

	avatarFile := file_storage_service.GenerateFileName("users", opts.File.Name, uint64(apiToken.GetID()))
	fileSource, err := base64.StdEncoding.DecodeString(opts.File.Base64)
	if err != nil {
		logger.Error(ctx, "Base64: %s", err)
		return nil, v1Errors.FILE_NOT_LOADED
	}
	err = handler.fileStorage.Upload(avatarFile, bytes.NewReader(fileSource))
	if err != nil {
		logger.Error(ctx, "AWS: %s", err)
		return nil, v1Errors.FILE_NOT_LOADED
	}

	coreOpts := coreApiAdapter.UserUpdateV1Args{
		ID:     uint64(apiToken.GetID()),
		Avatar: &avatarFile,
	}

	createData, err := handler.coreApi.UserUpdateV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: NEW_EMAIL_IS_BUSY" {
			return nil, v1Errors.USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: NEW_PHONE_IS_BUSY" {
			return nil, v1Errors.USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: USER_NOT_FOUND" {
			return nil, v1Errors.USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: UPDATE_FAILED" {
			return nil, v1Errors.USER_UPDATE_FAILED
		}
		return nil, err
	}
	if createData.User == nil {
		return nil, v1Errors.USER_UPDATE_FAILED
	}

	user := createData.User
	return &V1Res{
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
