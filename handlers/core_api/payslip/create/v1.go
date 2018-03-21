package payslip_create

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	EmployeeFK uint64      `key:"fk_employee" description:"Employee ID"`
	Payslip    PayslipData `key:"payslip" description:"Payslip"`
}

type PayslipData struct {
	Currency string  `key:"currency" description:"Currency"`
	Amount   float64 `key:"amount" description:"Amount"`
	Data     string  `key:"data" description:"Data"`
}

type V1Res struct {
	Employee *Employee `json:"agent" description:"Agent"`
	Payslip  *Payslip  `json:"agent" description:"Agent"`
}

type Employee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type Payslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	Data       []byte  `json:"data"`
	UpdateAt   string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

type V1ErrorTypes struct {
	EMPLOYEE_NOT_FOUND  error `text:"employee not found"`
	CREATE_FAILED       error `text:"creating payslip is failed"`
	PAYSLIP_NOT_CREATED error `text:"created payslip not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/Create/V1")
	cache.DisableTransportCache(ctx)

	employee, err := handler.agentService.GetEmployeeByID(ctx, opts.EmployeeFK)
	if err != nil {
		logger.Error(ctx, "Failed loading: %v", err)
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}
	if employee == nil {
		logger.Error(ctx, "Failed loading: employee is nil")
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}

	newPayslip := &models.Payslip{
		EmployeeFK: opts.EmployeeFK,
		Currency:   opts.Payslip.Currency,
		Amount:     opts.Payslip.Amount,
		Data:       []byte(opts.Payslip.Data),
	}

	payslipID, err := handler.paySlipService.SetPayslip(ctx, newPayslip)
	if err != nil {
		logger.Error(ctx, "Failed creating payslip: %v", err)
		return nil, v1Errors.CREATE_FAILED
	}
	if payslipID == 0 {
		logger.Error(ctx, "Failed creating payslip: payslipID is 0")
		return nil, v1Errors.CREATE_FAILED
	}

	payslip, err := handler.paySlipService.GetPayslipByID(ctx, payslipID)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.PAYSLIP_NOT_CREATED
	}
	if payslip == nil {
		logger.Error(ctx, "Failed login: payslip is nil")
		return nil, v1Errors.PAYSLIP_NOT_CREATED
	}

	return &V1Res{
		Payslip: &Payslip{
			ID:         payslip.ID,
			EmployeeFK: payslip.EmployeeFK,
			Currency:   payslip.Currency,
			Amount:     payslip.Amount,
			//Data:       payslip.Data,
			UpdateAt:  payslip.UpdateAt,
			CreatedAt: payslip.CreatedAt,
		},
		Employee: &Employee{
			ID:                 employee.ID,
			AgentFK:            employee.AgentFK,
			UserFK:             employee.UserFK,
			Code:               employee.Code,
			Name:               employee.Name,
			Role:               employee.Role,
			Email:              employee.Email,
			HireDate:           employee.HireDate,
			NumberOfDepandants: employee.NumberOfDepandants,
			GrossBaseSalary:    employee.GrossBaseSalary,
			UpdatedAt:          employee.UpdatedAt,
			CreatedAt:          employee.CreatedAt,
		},
	}, nil
}
