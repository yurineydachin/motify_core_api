package employee_invite

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
	return "Send instuction with QR code to employee"
}

func (*Handler) Description() string {
	return "Send instuction with QR code to employee"
}
