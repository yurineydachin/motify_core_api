package employee_invite

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
	return "Send invitation to employee"
}

func (*Handler) Description() string {
	return "Send email invitation to employee"
}
