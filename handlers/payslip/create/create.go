package payslip_create

import (
	"motify_core_api/srv/agent"
	"motify_core_api/srv/payslip"
)

type Handler struct {
	agentService   *agent_service.AgentService
	paySlipService *payslip_service.PaySlipService
}

func New(agentService *agent_service.AgentService, paySlipService *payslip_service.PaySlipService) *Handler {
	return &Handler{
		agentService:   agentService,
		paySlipService: paySlipService,
	}
}

func (*Handler) Caption() string {
	return "Create payslip"
}

func (*Handler) Description() string {
	return "Create payslip"
}
