package agent_details

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
	return "Agent details"
}

func (*Handler) Description() string {
	return "Agent details"
}
