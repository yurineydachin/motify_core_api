package employee_details

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/models"
)

const (
	PayslipLimit = 10
)

type V1Args struct {
	ID            *uint64 `key:"id_employee" description:"Employee ID"`
	AgentFK       *uint64 `key:"fk_agent" description:"Agent ID"`
	UserFK        *uint64 `key:"fk_user" description:"Mobile user ID"`
	IntegrationFK *uint64 `key:"fk_integraion" description:"Integration ID"`
	CompanyID     *string `key:"company_id" description:"Company id"`
	Code          *string `key:"employee_code" description:"employee code"`
}

type V1Res struct {
	Agent    *Agent    `json:"agent" description:"Agent"`
	Employee *Employee `json:"employee" description:"Employee"`
	Payslips []Payslip `json:"payslips" description:"Payslips"`
}

type Agent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
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
}

type V1ErrorTypes struct {
	MISSED_REQUIRED_FIELDS error `text:"Missed required fields. You should set id_employee or fk_agent, fk_user or fk_agent, emploee_code or fk_integration, company_id, emploee_code to find employee"`
	AGENT_NOT_FOUND        error `text:"agent not found"`
	EMPLOYEE_NOT_FOUND     error `text:"employee not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Employee/Update/V1")
	cache.DisableTransportCache(ctx)

	var employee *models.Employee
	var err error
	if opts.ID != nil && *opts.ID > 0 {
		employee, err = handler.agentService.GetEmployeeByID(ctx, *opts.ID)
	} else if opts.AgentFK != nil && *opts.AgentFK > 0 && opts.UserFK != nil && *opts.UserFK > 0 {
		employee, err = handler.agentService.GetEmployeeByAgentAndMobileUser(ctx, *opts.AgentFK, *opts.UserFK)
	} else if opts.AgentFK != nil && *opts.AgentFK > 0 && opts.Code != nil && *opts.Code != "" {
		employee, err = handler.agentService.GetEmployeeByAgentAndEmploeeCode(ctx, *opts.AgentFK, *opts.Code)
	} else if opts.IntegrationFK != nil && *opts.IntegrationFK > 0 && opts.CompanyID != nil && *opts.CompanyID != "" && opts.Code != nil && *opts.Code != "" {
		employee, err = handler.agentService.GetEmployeeByCompanyIDAndEmploeeCode(ctx, *opts.IntegrationFK, *opts.CompanyID, *opts.Code)
	} else {
		return nil, v1Errors.MISSED_REQUIRED_FIELDS
	}

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

	payslips, err := handler.payslipService.GetListByEmployeeID(ctx, employee.ID, PayslipLimit, 0)
	if err != nil {
		logger.Error(ctx, "Failed loading payslips %d: %v", employee.ID, err)
	}
	payslipsRes := make([]Payslip, 0, len(payslips))
	for i := range payslips {
		p := payslips[i]
		payslipsRes = append(payslipsRes, Payslip{
			ID:         p.ID,
			EmployeeFK: p.EmployeeFK,
			Title:      p.Title,
			Currency:   p.Currency,
			Amount:     p.Amount,
			UpdatedAt:  p.UpdatedAt,
			CreatedAt:  p.CreatedAt,
		})
	}

	return &V1Res{
		Agent: &Agent{
			ID:            agent.ID,
			IntegrationFK: agent.IntegrationFK,
			Name:          agent.Name,
			CompanyID:     agent.CompanyID,
			Description:   agent.Description,
			Logo:          agent.Logo,
			Background:    agent.Background,
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
		Payslips: payslipsRes,
	}, nil
}
