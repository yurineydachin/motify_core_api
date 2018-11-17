package payslip_create

import (
	"motify_core_api/srv/agent"
	"motify_core_api/srv/payslip"
	"motify_core_api/srv/push"
)

type Handler struct {
	agentService   *agent_service.AgentService
	payslipService *payslip_service.PayslipService
	pushService    *push_service.Service
}

func New(agentService *agent_service.AgentService, payslipService *payslip_service.PayslipService, pushService *push_service.Service) *Handler {
	return &Handler{
		agentService:   agentService,
		payslipService: payslipService,
		pushService:    pushService,
	}
}

func (*Handler) Caption() string {
	return "Create payslip"
}

func (*Handler) Description() string {
	return "Create payslip"
}
