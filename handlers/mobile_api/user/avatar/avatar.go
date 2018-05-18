package user_avatar

import (
	"motify_core_api/resources/file_storage"
	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type Handler struct {
	coreApi     *coreApiAdapter.MotifyCoreAPI
	fileStorage *file_storage_service.FileStorageService
}

func New(coreApi *coreApiAdapter.MotifyCoreAPI, fileStorage *file_storage_service.FileStorageService) *Handler {
	return &Handler{
		coreApi:     coreApi,
		fileStorage: fileStorage,
	}
}

func (*Handler) Caption() string {
	return "Load user avatar"
}

func (*Handler) Description() string {
	return "Load user avatar in base64"
}
