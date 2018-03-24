package employer_details

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type V1Args struct {
	ID      *uint64 `key:"id_employee" description:"Employee ID"`
	AgentFK *uint64 `key:"id_agent" description:"Agent ID"`
}

type V1Res struct {
	Agent    *Agent    `json:"agent" description:"Agent"`
	Employee *Employee `json:"employee" description:"Employee"`
	Payslips []Payslip `json:"payslips" description:"Payslips"`
}

type Agent struct {
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
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
}

type Payslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
}

type V1ErrorTypes struct {
	MISSED_REQUIRED_FIELDS error `text:"Missed required fields. You should set id_employee or fk_agent and fk_user to find employee"`
	AGENT_NOT_FOUND        error `text:"agent not found"`
	EMPLOYEE_NOT_FOUND     error `text:"employee not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Employeer/List/V1")
	cache.DisableTransportCache(ctx)

	userID := uint64(apiToken.GetCustomerID())
	coreOpts := coreApiAdapter.EmployeeDetailsV1Args{
		ID:      opts.ID,
		AgentFK: opts.AgentFK,
		UserFK:  &userID,
	}

	data, err := handler.coreApi.EmployeeDetailsV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: MISSED_REQUIRED_FIELDS" {
			return nil, v1Errors.MISSED_REQUIRED_FIELDS
		} else if err.Error() == "MotifyCoreAPI: AGENT_NOT_FOUND" {
			return nil, v1Errors.AGENT_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: EMPLOYEE_NOT_FOUND" {
			return nil, v1Errors.EMPLOYEE_NOT_FOUND
		}
		return nil, err
	}
	if data.Agent == nil {
		return nil, v1Errors.AGENT_NOT_FOUND
	} else if data.Employee == nil {
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}

	payslipsRes := make([]Payslip, 0, len(data.Payslips))
	for i := range data.Payslips {
		p := data.Payslips[i]
		payslipsRes = append(payslipsRes, Payslip{
			ID:         p.ID,
			EmployeeFK: p.EmployeeFK,
			Title:      p.Title,
			Currency:   p.Currency,
			Amount:     p.Amount,
		})
	}

	agent := data.Agent
	employee := data.Employee
	return &V1Res{
		Agent: &Agent{
			ID:          agent.ID,
			Name:        agent.Name,
			CompanyID:   agent.CompanyID,
			Description: agent.Description,
			Logo:        agent.Logo,
			Phone:       agent.Phone,
			Email:       agent.Email,
			Address:     agent.Address,
			Site:        agent.Site,
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
		},
		Payslips: payslipsRes,
	}, nil
}
