package employee_details

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	EmployeeHash string `key:"employee_hash" description:"Employee hash"`
}

type V1Res struct {
	Employee *Employee `json:"employee" description:"Employee"`
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
	ERROR_PARSING_HASH error `text:"Error parsing hash"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Employee/List/V1")
	cache.DisableTransportCache(ctx)

	integrationID := apiToken.GetExtraID()

	t, err := wrapToken.ParseEmployee(opts.EmployeeHash)
	if err != nil {
		logger.Error(ctx, "Error parse agent hash: ", err)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if t.GetExtraID() != integrationID {
		logger.Error(ctx, "Wrong agent hash (integration_id not equal): %d != %d", t.GetExtraID(), integrationID)
		return nil, v1Errors.ERROR_PARSING_HASH
	}
	employeeID := t.GetID()

	coreOpts := coreApiAdapter.EmployeeDetailsV1Args{
		ID: &employeeID,
	}

	data, err := handler.coreApi.EmployeeDetailsV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}
	if data == nil || data.Employee == nil {
		return nil, nil
	}

	employee := data.Employee
	res := V1Res{
		Employee: &Employee{
			Hash:               wrapToken.NewEmployee(employee.ID, integrationID).Fixed().String(),
			Code:               employee.Code,
			Name:               employee.Name,
			Role:               employee.Role,
			Email:              employee.Email,
			HireDate:           employee.HireDate,
			NumberOfDepandants: employee.NumberOfDepandants,
			GrossBaseSalary:    employee.GrossBaseSalary,
		},
	}
	return &res, nil
}
