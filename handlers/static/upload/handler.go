package upload

import (
	"encoding/json"
	"errors"
	"net/http"

	"motify_core_api/godep_libs/mobapi_lib/handler"
	"motify_core_api/godep_libs/mobapi_lib/handlersmanager"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"
	"motify_core_api/resources/file_storage"
	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type Handler struct {
	tokenModel  uint64
	dirPath     string
	fileStorage *file_storage.FileStorage
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

var (
	FILE_NOT_LOADED    = errors.New("File not loaded")
	USER_UPDATE_FAILED = errors.New("User updating failed")
)

func New(tokenModel uint64, dirPath string, coreApi *coreApiAdapter.MotifyCoreAPI, fileStorage *file_storage.FileStorage) *Handler {
	return &Handler{
		tokenModel:  tokenModel,
		dirPath:     dirPath,
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
		b, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		w.Write(b)
		return
	}

	b, err := json.Marshal(struct {
		Data interface{} `json:"data"`
	}{
		Data: res,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		w.Write(b)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *Handler) midleware(r *http.Request) (interface{}, error) {
	if r.Method != "POST" {
		return nil, errors.New("Need POST-request")
	}
	apiToken, _, err := handlersmanager.PrepareToken(r.Header.Get(handler.HeaderAPIToken), handlersmanager.TokenTypeAuthorized, h.tokenModel)
	if err != nil {
		return nil, errors.New("Token error: " + err.Error())
	}

	res, err := h.v1(r, apiToken)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) v1(r *http.Request, apiToken token.IToken) (*V1Res, error) {
	r.ParseMultipartForm(32 << 20)
	file, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		logger.Error(r.Context(), "Error upload file: %s", err)
		return nil, FILE_NOT_LOADED
	}
	defer file.Close()

	fileName := file_storage.GenerateFileName(getModel(h.tokenModel), fileHeader.Filename, uint64(apiToken.GetID()))

	savepath, err := h.fileStorage.Upload(fileName, file)
	if err != nil {
		logger.Error(r.Context(), "File not saved: %s", err)
		return nil, FILE_NOT_LOADED
	}
	logger.Error(r.Context(), "File uploaded to: %s", savepath)

	coreOpts := coreApiAdapter.UserUpdateV1Args{
		ID:     uint64(apiToken.GetID()),
		Avatar: &fileName,
	}

	data, err := h.coreApi.UserUpdateV1(r.Context(), coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: NEW_EMAIL_IS_BUSY" {
			return nil, USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: NEW_PHONE_IS_BUSY" {
			return nil, USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: USER_NOT_FOUND" {
			return nil, USER_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: UPDATE_FAILED" {
			return nil, USER_UPDATE_FAILED
		}
		return nil, err
	}
	if data.User == nil {
		return nil, USER_UPDATE_FAILED
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
}
