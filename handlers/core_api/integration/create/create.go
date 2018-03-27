package integration_create

import (
	"motify_core_api/srv/integration"
)

type Handler struct {
	integrationService *integration_service.IntegrationService
}

func New(integrationService *integration_service.IntegrationService) *Handler {
	return &Handler{
		integrationService: integrationService,
	}
}

func (*Handler) Caption() string {
	return "Create integration"
}

func (*Handler) Description() string {
	return "Create integration"
}
