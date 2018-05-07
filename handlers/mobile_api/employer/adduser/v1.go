package employer_adduser

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Code string `key:"magic_code" description:"Magic code to find employee"`
}

type V1Res struct {
	Agent    *Agent    `json:"agent" description:"Agent"`
	Employee *Employee `json:"employee" description:"Employee"`
	User     *User     `json:"user" description:"User"`
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

type User struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type V1ErrorTypes struct {
	ERROR_PARSE_MAGIC_CODE error `text:"Error parse magic code"`
	EMPLOYEE_UPDATE_FAILED error `text:"User creating failed"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "User/Update/V1")
	cache.DisableTransportCache(ctx)

	employeeToken, err := wrapToken.ParseEmployee(opts.Code)
	if err != nil {
		logger.Error(ctx, "Error parse magic code: ", err)
		return nil, v1Errors.ERROR_PARSE_MAGIC_CODE
	}

	employeeID := uint64(employeeToken.GetID())
	userID := uint64(apiToken.GetID())
	coreOpts := coreApiAdapter.EmployeeUpdateV1Args{
		ID:     &employeeID,
		UserFK: &userID,
	}

	updateData, err := handler.coreApi.EmployeeUpdateV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: AGENT_NOT_FOUND" {
			return nil, v1Errors.EMPLOYEE_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: EMPLOYEE_NOT_FOUND" {
			return nil, v1Errors.EMPLOYEE_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: EMPLOYEE_NOT_UPDATED" {
			return nil, v1Errors.EMPLOYEE_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: USER_NOT_FOUND" {
			return nil, v1Errors.EMPLOYEE_UPDATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: UPDATE_FAILED" {
			return nil, v1Errors.EMPLOYEE_UPDATE_FAILED
		}
		return nil, err
	}
	if updateData.User == nil || updateData.Agent == nil || updateData.Employee == nil {
		return nil, v1Errors.EMPLOYEE_UPDATE_FAILED
	}
	if updateData.Employee.UserFK == nil || *updateData.Employee.UserFK != userID {
		return nil, v1Errors.EMPLOYEE_UPDATE_FAILED
	}

	user := updateData.User
	agent := updateData.Agent
	employee := updateData.Employee
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
		User: &User{
			Hash:        wrapToken.NewMobileUser(user.ID).Fixed().String(),
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Avatar:      user.Avatar,
			Phone:       user.Phone,
			Email:       user.Email,
		},
	}, nil
}
