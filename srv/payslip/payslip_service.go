package payslip_service

import (
	"motify_core_api/resources/database/interfaces"
)

type PaySlipService struct {
	db interfaces.IQueryer
}

func NewPayslipService(db interfaces.IQueryer) *PaySlipService {
	return &PaySlipService{
		db: db,
	}
}
