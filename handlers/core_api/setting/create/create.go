package setting_create

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
	return "Create setting for agent"
}

func (*Handler) Description() string {
	return "Create setting for agent"
}
