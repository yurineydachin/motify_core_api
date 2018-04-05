package employee_invite

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/utils/qrcode"
	wrapToken "motify_core_api/utils/token"
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
	logger.Debug(ctx, "Employee/Invite/V1")
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

	email := employee.Email
	if opts.Email != nil && *opts.Email != "" {
		email = *opts.Email
	}

	magicCode := wrapToken.NewEmployeeQR(employee.ID).String()
	code, err := qrcode.Generate(magicCode, 0)
	status := "Email not sended"
	if err != nil {
		logger.Error(ctx, "Error generate QR code: %v", err)
		status = "Error generate QR code"
	} else if email != "" && handler.emailFrom != "" {
		err = handler.emailService.SendEmployeeInvite(ctx, email, handler.emailFrom, code)
		if err != nil {
			logger.Error(ctx, "Error sending email: %v", err)
			status = "Error sending email"
		} else {
			status = "OK"
		}
	} else {
		status = "Email is empty"
		logger.Error(ctx, "Email not sended: some email is empty: email '%s', handler.emailFrom '%s'", email, handler.emailFrom)
	}

	return &V1Res{
		Result: status,
		Code:   magicCode,
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
