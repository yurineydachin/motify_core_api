package payslip_list

import (
	"motify_core_api/srv/payslip"
)

type Handler struct {
	payslipService *payslip_service.PayslipService
}

func New(payslipService *payslip_service.PayslipService) *Handler {
	return &Handler{
		payslipService: payslipService,
	}
}

func (*Handler) Caption() string {
	return "Payslips list"
}

func (*Handler) Description() string {
	return "Payslips list"
}
