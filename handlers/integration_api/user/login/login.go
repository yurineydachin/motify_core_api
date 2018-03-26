package user_login

import (
	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type Handler struct {
	coreApi *coreApiAdapter.MotifyCoreAPI
}

func New(coreApi *coreApiAdapter.MotifyCoreAPI) *Handler {
	return &Handler{
		coreApi: coreApi,
	}
}

func (*Handler) Caption() string {
	return "Login user"
}

func (*Handler) Description() string {
	return "Try login user and return user with token if success"
}
