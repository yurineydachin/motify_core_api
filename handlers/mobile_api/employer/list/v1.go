package employer_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
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
}

type Agent struct {
	Hash        string `json:"hash"`
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
	Hash               string  `json:"hash"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
}

type V1ErrorTypes struct {
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Employeer/List/V1")
	cache.DisableTransportCache(ctx)

	coreOpts := coreApiAdapter.AgentListV1Args{
		UserID: uint64(apiToken.GetID()),
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}

	data, err := handler.coreApi.AgentListV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}

	res := V1Res{
		List: make([]ListItem, len(data.List)),
	}
	for i := range data.List {
		agent := data.List[i].Agent
		res.List[i].Agent = Agent{
			Hash:        wrapToken.NewAgent(agent.ID, agent.IntegrationFK).Fixed().String(),
			Name:        agent.Name,
			CompanyID:   agent.CompanyID,
			Description: agent.Description,
			Logo:        agent.Logo,
			Phone:       agent.Phone,
			Email:       agent.Email,
			Address:     agent.Address,
			Site:        agent.Site,
		}
		employee := data.List[i].Employee
		res.List[i].Employee = Employee{
			Hash:               wrapToken.NewEmployee(employee.ID, agent.IntegrationFK).Fixed().String(),
			Code:               employee.Code,
			Name:               employee.Name,
			Role:               employee.Role,
			Email:              employee.Email,
			HireDate:           employee.HireDate,
			NumberOfDepandants: employee.NumberOfDepandants,
			GrossBaseSalary:    employee.GrossBaseSalary,
		}
	}
	return &res, nil
}
