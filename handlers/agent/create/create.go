package agent_create

import (
	"motify_core_api/srv/agent"
)

type Handler struct {
	agentService *agent_service.AgentService
}

func New(agentService *agent_service.AgentService) *Handler {
	return &Handler{
		agentService: agentService,
	}
}

func (*Handler) Caption() string {
	return "Create agent"
}

func (*Handler) Description() string {
	return "Create agent"
}
