package employee_update

import (
	"motify_core_api/srv/agent"
	"motify_core_api/srv/user"
)

type Handler struct {
	agentService *agent_service.AgentService
	userService  *user_service.UserService
}

func New(agentService *agent_service.AgentService, userService *user_service.UserService) *Handler {
	return &Handler{
		agentService: agentService,
		userService:  userService,
	}
}

func (*Handler) Caption() string {
	return "Update employee for agent"
}

func (*Handler) Description() string {
	return "Update employee for agent"
}
