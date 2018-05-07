package employee_update

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	EmployeeHash       string   `key:"employee_hash" description:"Employee hash"`
	Code               *string  `key:"employee_code" description:"Employee code"`
	Name               *string  `key:"name" description:"Name"`
	Role               *string  `key:"role" description:"Role"`
	Email              *string  `key:"email" description:"Email"`
	HireDate           *string  `key:"hire_date" description:"Hire date"`
	NumberOfDepandants *uint    `key:"number_of_dependants" description:"number of depandants"`
	GrossBaseSalary    *float64 `key:"gross_base_salary" description:"Gross base salary"`
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
	ERROR_PARSING_HASH   error `text:"Error parsing hash"`
	EMPLOYEE_NOT_UPDATED error `text:"Error updating employee"`
	EMPLOYEE_NOT_FOUND   error `text:"Employee not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Employee/Update/V1")
	cache.DisableTransportCache(ctx)

	integrationID := apiToken.GetExtraID()

	t, err := wrapToken.ParseEmployee(opts.EmployeeHash)
	employeeID := t.GetID()
	if err != nil {
		logger.Error(ctx, "Error parse employee hash: ", err)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if t.GetExtraID() != integrationID {
		logger.Error(ctx, "Wrong employee hash (integration_id not equal): %d != %d", t.GetExtraID(), integrationID)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if employeeID == 0 {
		logger.Error(ctx, "Wrong employee hash (employeeID = 0)")
		return nil, v1Errors.ERROR_PARSING_HASH
	}

	coreOpts := coreApiAdapter.EmployeeUpdateV1Args{
		ID:                 &employeeID,
		Code:               opts.Code,
		Name:               opts.Name,
		Role:               opts.Role,
		Email:              opts.Email,
		HireDate:           opts.HireDate,
		NumberOfDepandants: opts.NumberOfDepandants,
		GrossBaseSalary:    opts.GrossBaseSalary,
	}
	data, err := handler.coreApi.EmployeeUpdateV1(ctx, coreOpts)
	if err != nil {
		return nil, v1Errors.EMPLOYEE_NOT_UPDATED
	}
	if data == nil || data.Employee == nil {
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
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
