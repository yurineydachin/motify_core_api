package employer_details

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	EmployeeHash *string `key:"employee_hash" description:"Employee hash"`
	AgentHash    *string `key:"agent_hash" description:"Agent hash"`
}

type V1Res struct {
	Agent    *Agent    `json:"agent" description:"Agent"`
	Employee *Employee `json:"employee" description:"Employee"`
	Payslips []Payslip `json:"payslips" description:"Payslips"`
}

type Agent struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Background  string `json:"bg_image"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
}

type Employee struct {
	Hash               string  `json:"hash"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
}

type Payslip struct {
	Hash     string  `json:"hash"`
	Title    string  `json:"title"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type V1ErrorTypes struct {
	MISSED_REQUIRED_FIELDS error `text:"Missed required fields. You should set id_employee or fk_agent and fk_user to find employee"`
	AGENT_NOT_FOUND        error `text:"agent not found"`
	EMPLOYEE_NOT_FOUND     error `text:"employee not found"`
	ERROR_PARSING_HASH     error `text:"Error parsing hash"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Employer/List/V1")
	cache.DisableTransportCache(ctx)

	var eID uint64
	if opts.EmployeeHash != nil && *opts.EmployeeHash != "" {
		t, err := wrapToken.ParseEmployee(*opts.EmployeeHash)
		if err != nil {
			logger.Error(ctx, "Error parse employee hash: ", err)
			return nil, v1Errors.ERROR_PARSING_HASH
		}
		eID = t.GetID()
	}
	var aID uint64
	if opts.AgentHash != nil && *opts.AgentHash != "" {
		t, err := wrapToken.ParseAgent(*opts.AgentHash)
		if err != nil {
			logger.Error(ctx, "Error parse agent hash: ", err)
			return nil, v1Errors.ERROR_PARSING_HASH
		}
		aID = t.GetID()
	}

	userID := uint64(apiToken.GetID())
	coreOpts := coreApiAdapter.EmployeeDetailsV1Args{
		ID:      &eID,
		AgentFK: &aID,
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

	agent := data.Agent
	employee := data.Employee
	payslipsRes := make([]Payslip, 0, len(data.Payslips))
	for i := range data.Payslips {
		p := data.Payslips[i]
		payslipsRes = append(payslipsRes, Payslip{
			Hash:     wrapToken.NewPayslip(p.ID, agent.IntegrationFK).Fixed().String(),
			Title:    p.Title,
			Currency: p.Currency,
			Amount:   p.Amount,
		})
	}

	return &V1Res{
		Agent: &Agent{
			Hash:        wrapToken.NewAgent(agent.ID, agent.IntegrationFK).Fixed().String(),
			Name:        agent.Name,
			CompanyID:   agent.CompanyID,
			Description: agent.Description,
			Logo:        agent.Logo,
			Background:  agent.Background,
			Phone:       agent.Phone,
			Email:       agent.Email,
			Address:     agent.Address,
			Site:        agent.Site,
		},
		Employee: &Employee{
			Hash:               wrapToken.NewEmployee(employee.ID, agent.IntegrationFK).Fixed().String(),
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
