package user_remind_reset

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
	return "Reset user password using code from email"
}

func (*Handler) Description() string {
	return "Reset user password using code from email"
}
