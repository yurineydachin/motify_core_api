package employee_invite

import (
	"motify_core_api/srv/agent"
	"motify_core_api/srv/email"
)

type Handler struct {
	agentService *agent_service.AgentService
	emailService *email_service.Service
	emailFrom    string
}

func New(agentService *agent_service.AgentService, emailService *email_service.Service, emailFrom string) *Handler {
	return &Handler{
		agentService: agentService,
		emailService: emailService,
		emailFrom:    emailFrom,
	}
}

func (*Handler) Caption() string {
	return "Send invitation to employee"
}

func (*Handler) Description() string {
	return "Send email invitation to employee"
}
