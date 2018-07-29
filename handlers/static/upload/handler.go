package upload 

import (
    "io"
    "os"
    "fmt"
    //"context"
    "net/http"

    "motify_core_api/godep_libs/service/logger"
    "motify_core_api/resources/file_storage"
    "motify_core_api/godep_libs/mobapi_lib/handlersmanager"
    wrapToken "motify_core_api/utils/token"
    coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type Handler struct {
    tokenModel uint64
    dirPath string
    fileStorage *file_storage_service.FileStorageService
    coreApi     *coreApiAdapter.MotifyCoreAPI
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

var model = map[uint64]string{
    wrapToken.ModelMobileUser: "users",
}
var defaultModelName = "unknown"

type V1ErrorTypes struct {
	FILE_NOT_LOADED    error `text:"File not loaded"`
	USER_UPDATE_FAILED error `text:"User creating failed"`
}

var v1Errors = V1ErrorTypes{}

func New(tokenModel uint64, dirPath string, coreApi *coreApiAdapter.MotifyCoreAPI, fileStorage *file_storage_service.FileStorageService) *Handler {
    return &Handler{
        tokenModel: tokenModel,
        dirPath: dirPath,
        coreApi:     coreApi,
        fileStorage: fileStorage,
    }
}

func getModel(tokenModel uint64) string {
    if m, ok := model[tokenModel]; ok && m != "" {
        return m
    }
    return defaultModelName
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    res, err := h.midleware(r)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error":"` + err.Error() + `"}`))
    }

    response := struct{
        Data interface{} `json:"data"`
    }
    response.Data = res
    b, err := json.Marshal(response)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error":"` + err.Error() + `"}`))
    }
    w.WriteHeader(http.StatusOK)
    w.Write(b))
}

func (h *Handler) midleware(r *http.Request) (interface{}, error)) {
    if r.Method == "POST" {
        return nil, fmt.Errorf("Need POST-request")
    }
    apiToken, _, err := handlersmanager.prepareToken(r.Header.Get(HeaderAPIToken), handlersmanager.TokenTypeAuthorized, h.tokenModel)
    if err != nil {
        return nil, fmt.Errorf("Token error: %s", err)
    }

    // if r.URL.Path == ""
    res, err := h.v1(r, apiToken)
    if err != nil {
        return nil, err
    }
    return res, nil
}

func v1(r *http.Prequest, apiToken token.IToken) (*V1Res, error) {
    r.ParseMultipartForm(32 << 20)
    file, fileHeader, err := r.FormFile("uploadFile")
    if err != nil {
        logger.Error(r.Context(), "Error upload file: %s", err)
        return nil, v1Errors.FILE_NOT_LOADED
    }
    defer file.Close()

    fileName := file_storage_service.GenerateFileName(getModel(h.tokenModel), fileHeader.FileName, uint64(apiToken.GetID()))

    err = h.fileStorage.Upload(fileName, file)
    if err != nil {
        logger.Error(r.Context(), "File not saved: %s", err)
        return nil, v1Errors.FILE_NOT_LOADED
    }

	coreOpts := coreApiAdapter.UserUpdateV1Args{
		ID:     uint64(apiToken.GetID()),
		Avatar: &avatarFile,
	}

	data, err := handler.coreApi.UserUpdateV1(r.Context(), coreOpts)
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
	if data.User == nil {
		return nil, v1Errors.USER_UPDATE_FAILED
	}

	user := data.User
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
    return nil
}
