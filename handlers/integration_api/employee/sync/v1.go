package employee_sync

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

var defaultLimit = uint64(1000000)

type V1Args struct {
	AgentHash string        `key:"agent_hash" description:"Agent hash"`
	List      []EmployeeArg `key:"employee_list" description:"Employee list"`
}

type EmployeeArg struct {
	Code               string   `key:"employee_code" description:"Code"`
	Name               string   `key:"name" description:"Name"`
	Role               *string  `key:"role" description:"Role"`
	Email              *string  `key:"email" description:"Email"`
	HireDate           *string  `key:"hire_date" description:"HireDate in format ISO8601"`
	NumberOfDepandants *uint    `key:"number_of_dependants" description:"Number of depandants"`
	GrossBaseSalary    *float64 `key:"gross_base_salary" description:"Gross base salary"`
}

type V1Res struct {
	List []*EmployeeStatus `json:"list" description:"Employee list with status"`
}

type EmployeeStatus struct {
	Hash   string `json:"hash"`
	Code   string `json:"employee_code"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type V1ErrorTypes struct {
	ERROR_PARSING_HASH   error `text:"Error parsing hash"`
	AGENT_NOT_FOUND      error `text:"Agent not found"`
	EMPLOYEE_NOT_UPDATED error `text:"Error updating employee"`
	EMPLOYEE_NOT_CREATED error `text:"Error creating employee"`
	EMPLOYEE_NOT_FOUND   error `text:"Employee not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Employee/Sync/V1")
	cache.DisableTransportCache(ctx)

	integrationID := apiToken.GetExtraID()

	t, err := wrapToken.ParseAgent(opts.AgentHash)
	agentID := t.GetID()
	if err != nil {
		logger.Error(ctx, "Error parse agent hash: ", err)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if t.GetExtraID() != integrationID {
		logger.Error(ctx, "Wrong agent hash (integration_id not equal): %d != %d", t.GetExtraID(), integrationID)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if agentID == 0 {
		logger.Error(ctx, "Wrong agent hash = 0, hash info: %#v", t)
		return nil, v1Errors.ERROR_PARSING_HASH
	}

	coreOpts := coreApiAdapter.EmployeeListV1Args{
		AgentID: agentID,
		Limit:   &defaultLimit,
	}

	data, err := handler.coreApi.EmployeeListV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}

	eMap := mapEmployees(data.List)

	res := &V1Res{
		List: make([]*EmployeeStatus, len(opts.List)),
	}
	for i := range opts.List {
		res.List[i] = &EmployeeStatus{
			Code:   opts.List[i].Code,
			Name:   opts.List[i].Name,
			Status: "-",
		}
		if e, ok := eMap[opts.List[i].Code]; ok {
			res.List[i].Hash = wrapToken.NewEmployee(e.ID, integrationID).Fixed().String()
			coreOpts := coreApiAdapter.EmployeeUpdateV1Args{
				ID:                 &e.ID,
				Code:               &opts.List[i].Code,
				Name:               &opts.List[i].Name,
				Role:               opts.List[i].Role,
				Email:              opts.List[i].Email,
				HireDate:           opts.List[i].HireDate,
				NumberOfDepandants: opts.List[i].NumberOfDepandants,
				GrossBaseSalary:    opts.List[i].GrossBaseSalary,
			}
			data, err := handler.coreApi.EmployeeUpdateV1(ctx, coreOpts)
			if err != nil {
				res.List[i].Status = "EMPLOYEE_NOT_UPDATED"
				continue
			}
			if data == nil || data.Employee == nil {
				res.List[i].Status = "EMPLOYEE_NOT_FOUND"
				continue
			}
			res.List[i].Status = "UPDATED"
		} else {
			coreOpts := coreApiAdapter.EmployeeCreateV1Args{
				AgentFK:            agentID,
				Code:               &opts.List[i].Code,
				Name:               opts.List[i].Name,
				Role:               opts.List[i].Role,
				Email:              opts.List[i].Email,
				HireDate:           opts.List[i].HireDate,
				NumberOfDepandants: opts.List[i].NumberOfDepandants,
				GrossBaseSalary:    opts.List[i].GrossBaseSalary,
			}
			data, err := handler.coreApi.EmployeeCreateV1(ctx, coreOpts)
			if err != nil {
				res.List[i].Status = "EMPLOYEE_NOT_CREATED"
				continue
			}
			if data == nil || data.Employee == nil {
				res.List[i].Status = "EMPLOYEE_NOT_FOUND"
				continue
			}
			res.List[i].Hash = wrapToken.NewEmployee(data.Employee.ID, integrationID).Fixed().String()
			res.List[i].Status = "CREATED"
		}
	}

	return res, nil
}

func mapEmployees(list []coreApiAdapter.EmployeeListListItem) map[string]coreApiAdapter.EmployeeListEmployee {
	res := make(map[string]coreApiAdapter.EmployeeListEmployee, len(list))
	for i := range list {
		if list[i].Employee.Code != "" {
			res[list[i].Employee.Code] = list[i].Employee
		}
	}
	return res
}
