package payslip_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type V1Args struct {
	Limit  *uint64 `key:"limit" description:"Limit"`
	Offset *uint64 `key:"offset" description:"Offset"`
}

type V1Res struct {
	List []ListItem `json:"list" description:"List of agents and employees"`
}

type ListItem struct {
	Agent    Agent    `json:"agent" description:"Agent"`
	Employee Employee `json:"employee" description:"Employee"`
	Payslip  Payslip  `json:"payslip" description:"Payslip"`
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
	UpdateAt   string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

type V1ErrorTypes struct {
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/List/V1")
	cache.DisableTransportCache(ctx)

	coreOpts := coreApiAdapter.PayslipListV1Args{
		UserID: uint64(apiToken.GetCustomerID()),
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}
	data, err := handler.coreApi.PayslipListV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}

	res := V1Res{
		List: make([]ListItem, 0, len(data.List)),
	}
	for i := range data.List {
		agent := data.List[i].Agent
		employee := data.List[i].Employee
		p := data.List[i].Payslip
		res.List = append(res.List, ListItem{
			Agent: Agent{
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
			Employee: Employee{
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
			Payslip: Payslip{
				ID:         p.ID,
				EmployeeFK: p.EmployeeFK,
				Title:      p.Title,
				Currency:   p.Currency,
				Amount:     p.Amount,
				UpdateAt:   p.UpdateAt,
				CreatedAt:  p.CreatedAt,
			},
		})
	}

	return &res, nil
}
