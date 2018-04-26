package payslip_details

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"
)

type v1Args struct {
	ID uint64 `key:"payslip_id" description:"Payslip id"`
}

type V1Res struct {
	Agent    *Agent    `json:"agent"`
	Employee *Employee `json:"employee"`
	Payslip  Payslip   `json:"payslip"`
}

type Agent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
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
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
	Data       string  `json:"data"`
}

type V1ErrorTypes struct {
	AGENT_NOT_FOUND    error `text:"agent not found"`
	EMPLOYEE_NOT_FOUND error `text:"employee not found"`
	PAYSLIP_NOT_FOUND  error `text:"payslip not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *v1Args) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/Details/V1")
	cache.EnableTransportCache(ctx)

	payslip, err := handler.payslipService.GetPayslipByID(ctx, opts.ID)
	if err != nil {
		logger.Error(ctx, "Failed loading from DB: %v", err)
		return nil, v1Errors.PAYSLIP_NOT_FOUND
	}
	if payslip == nil {
		logger.Error(ctx, "Failed loading payslip is nil")
		return nil, v1Errors.PAYSLIP_NOT_FOUND
	}

	employee, err := handler.agentService.GetEmployeeByID(ctx, payslip.EmployeeFK)
	if err != nil {
		logger.Error(ctx, "Failed loading from DB: %v", err)
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}
	if employee == nil {
		logger.Error(ctx, "Failed loading employee is nil")
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}

	agent, err := handler.agentService.GetAgentByID(ctx, employee.AgentFK)
	if err != nil {
		logger.Error(ctx, "Failed loading agent %d: %v", employee.AgentFK, err)
		return nil, v1Errors.AGENT_NOT_FOUND
	}
	if agent == nil {
		logger.Error(ctx, "Failed loading agent (%d) is nil", employee.AgentFK)
		return nil, v1Errors.AGENT_NOT_FOUND
	}

	return &V1Res{
		Agent: &Agent{
			ID:            agent.ID,
			IntegrationFK: agent.IntegrationFK,
			Name:          agent.Name,
			CompanyID:     agent.CompanyID,
			Description:   agent.Description,
			Logo:          agent.Logo,
			Phone:         agent.Phone,
			Email:         agent.Email,
			Address:       agent.Address,
			Site:          agent.Site,
			UpdatedAt:     agent.UpdatedAt,
			CreatedAt:     agent.CreatedAt,
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
		Payslip: Payslip{
			ID:         payslip.ID,
			EmployeeFK: payslip.EmployeeFK,
			Title:      payslip.Title,
			Currency:   payslip.Currency,
			Amount:     payslip.Amount,
			UpdatedAt:  payslip.UpdatedAt,
			CreatedAt:  payslip.CreatedAt,
			Data:       string(payslip.Data),
		},
	}, nil
}
