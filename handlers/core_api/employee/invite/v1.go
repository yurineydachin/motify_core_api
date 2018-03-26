package employee_invite

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"
)

type V1Args struct {
	ID    uint64  `key:"id_employee" description:"Employee ID"`
	Email *string `key:"email" description:"Email"`
}

type V1Res struct {
	Result   string    `json:"result" description:"Result status"`
	Code     string    `json:"magic_code" description:"Magic code for generating QR code"`
	Employee *Employee `json:"employee" description:"Employee"`
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

type V1ErrorTypes struct {
	MISSED_REQUIRED_FIELDS error `text:"Missed required fields. You should set id_employee or fk_agent and fk_user to find employee"`
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

	employee, err := handler.agentService.GetEmployeeByID(ctx, opts.ID)
	if err != nil {
		logger.Error(ctx, "Failed loading from DB: %v", err)
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}
	if employee == nil {
		logger.Error(ctx, "Failed loading employee is nil")
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}

	return &V1Res{
		Result: "OK",
		Code:   token.NewTokenV1(employee.ID, 2).String(),
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
	}, nil
}
