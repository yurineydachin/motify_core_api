package integration_check

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
	return "Check integration"
}

func (*Handler) Description() string {
	return "Check integration"
}
