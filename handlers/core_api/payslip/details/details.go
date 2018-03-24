package payslip_details

import (
	"motify_core_api/srv/agent"
	"motify_core_api/srv/payslip"
)

type Handler struct {
	agentService   *agent_service.AgentService
	payslipService *payslip_service.PayslipService
}

func New(agentService *agent_service.AgentService, payslipService *payslip_service.PayslipService) *Handler {
	return &Handler{
		agentService:   agentService,
		payslipService: payslipService,
	}
}

func (*Handler) Caption() string {
	return "Payslip details"
}

func (*Handler) Description() string {
	return "Payslip details"
}
