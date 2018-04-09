package payslip_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"
)

type V1Args struct {
	UserID uint64  `key:"user_id" description:"User ID"`
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
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/Create/V1")
	cache.DisableTransportCache(ctx)

	limit := uint64(0)
	if opts.Limit != nil && *opts.Limit > 0 {
		limit = *opts.Limit
	}
	offset := uint64(0)
	if opts.Offset != nil && *opts.Offset > 0 {
		offset = *opts.Offset
	}

	list, err := handler.payslipService.GetListByUserID(ctx, opts.UserID, limit, offset)
	if err != nil {
		logger.Error(ctx, "Failed loading payslips by user %d: %v", opts.UserID, err)
		return nil, err
	}
	res := V1Res{
		List: make([]ListItem, 0, len(list)),
	}
	for i := range list {
		agent := list[i].Agent
		employee := list[i].Employee
		p := list[i].Payslip
		res.List = append(res.List, ListItem{
			Agent: Agent{
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
				UpdatedAt:          employee.UpdatedAt,
				CreatedAt:          employee.CreatedAt,
			},
			Payslip: Payslip{
				ID:         p.ID,
				EmployeeFK: p.EmployeeFK,
				Title:      p.Title,
				Currency:   p.Currency,
				Amount:     p.Amount,
				UpdatedAt:  p.UpdatedAt,
				CreatedAt:  p.CreatedAt,
			},
		})
	}

	return &res, nil
}
