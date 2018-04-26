package employee_update

import (
	"context"
	"strings"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	ID                 *uint64  `key:"id_employee" description:"Employee ID"`
	AgentFK            *uint64  `key:"fk_agent" description:"Agent ID"`
	UserFK             *uint64  `key:"fk_user" description:"User ID"`
	Code               *string  `key:"employee_code" description:"Employee code"`
	Name               *string  `key:"name" description:"Name"`
	Role               *string  `key:"role" description:"Role"`
	Email              *string  `key:"email" description:"Email"`
	HireDate           *string  `key:"hire_date" description:"Hire date"`
	NumberOfDepandants *uint    `key:"number_of_dependants" description:"number of depandants"`
	GrossBaseSalary    *float64 `key:"gross_base_salary" description:"Gross base salary"`
}

type V1Res struct {
	Agent    *Agent    `json:"agent" description:"Agent"`
	Employee *Employee `json:"employee" description:"Employee"`
	User     *User     `json:"user" description:"User"`
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

type User struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Awatar      string `json:"awatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type V1ErrorTypes struct {
	NOT_ENOUGH_PARAMS       error `text:"need params id_employee or fk_agent and fk_user to find employee"`
	AGENT_NOT_FOUND         error `text:"agent not found"`
	EMPLOYEE_NOT_FOUND      error `text:"employee not found"`
	USER_NOT_FOUND          error `text:"user not found"`
	UPDATE_FAILED           error `text:"updating employee is failed"`
	EMPLOYEE_NOT_UPDATED    error `text:"updated employee not found"`
	EMPLOYEE_ALREADY_EXISTS error `text:"employee already exists for this agent and user"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Employee/Update/V1")
	cache.DisableTransportCache(ctx)

	var employee *models.Employee
	var err error
	if opts.ID != nil && *opts.ID > 0 {
		employee, err = handler.agentService.GetEmployeeByID(ctx, *opts.ID)
	} else if opts.AgentFK != nil && *opts.AgentFK > 0 && opts.UserFK != nil && *opts.UserFK > 0 {
		employee, err = handler.agentService.GetEmployeeByAgentAndMobileUser(ctx, *opts.AgentFK, *opts.UserFK)
	} else {
		return nil, v1Errors.NOT_ENOUGH_PARAMS
	}
	if err != nil {
		logger.Error(ctx, "Failed loading from DB: %v", err)
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}
	if employee == nil {
		logger.Error(ctx, "Failed loading employee is nil")
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}

	needUpdate := false
	agentID := employee.AgentFK
	if opts.AgentFK != nil && *opts.AgentFK > 0 && employee.AgentFK != *opts.AgentFK {
		agentID = *opts.AgentFK
		needUpdate = true
	}

	agent, err := handler.agentService.GetAgentByID(ctx, agentID)
	if err != nil {
		logger.Error(ctx, "Failed loading agent %d: %v", agentID, err)
		return nil, v1Errors.AGENT_NOT_FOUND
	}
	if agent == nil {
		logger.Error(ctx, "Failed loading agent (%d) is nil", agentID)
		return nil, v1Errors.AGENT_NOT_FOUND
	}
	employee.AgentFK = agent.ID

	var userRes *User
	userID := uint64(0)
	if opts.UserFK != nil && *opts.UserFK > 0 {
		userID = *opts.UserFK
		if employee.UserFK == nil {
			needUpdate = true
		} else if employee.UserFK != nil && *employee.UserFK != *opts.UserFK {
			needUpdate = true
		}
	} else if employee.UserFK != nil && *employee.UserFK > 0 {
		userID = *employee.UserFK
	}
	if userID > 0 {
		user, err := handler.userService.GetUserByID(ctx, userID)
		if err != nil {
			logger.Error(ctx, "Failed loading %v", err)
			return nil, v1Errors.USER_NOT_FOUND
		}
		if user == nil {
			logger.Error(ctx, "Failed loading user is nil")
			return nil, v1Errors.USER_NOT_FOUND
		}
		employee.UserFK = &user.ID
		userRes = &User{
			ID:          user.ID,
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Awatar:      user.Awatar,
			Phone:       user.Phone,
			Email:       user.Email,
			UpdatedAt:   user.UpdatedAt,
			CreatedAt:   user.CreatedAt,
		}
	}

	if opts.Name != nil && *opts.Name != "" && employee.Name != *opts.Name {
		employee.Name = *opts.Name
		needUpdate = true
	}
	if opts.Code != nil && *opts.Code != "" && employee.Code != *opts.Code {
		employee.Code = *opts.Code
		needUpdate = true
	}
	if opts.Role != nil && *opts.Role != "" && employee.Role != *opts.Role {
		employee.Role = *opts.Role
		needUpdate = true
	}
	if opts.Email != nil && *opts.Email != "" && employee.Email != *opts.Email {
		employee.Email = *opts.Email
		needUpdate = true
	}
	if opts.HireDate != nil && *opts.HireDate != "" && employee.HireDate != *opts.HireDate {
		employee.HireDate = *opts.HireDate
		needUpdate = true
	}
	if opts.NumberOfDepandants != nil && *opts.NumberOfDepandants > 0 && employee.NumberOfDepandants != *opts.NumberOfDepandants {
		employee.NumberOfDepandants = *opts.NumberOfDepandants
		needUpdate = true
	}
	if opts.GrossBaseSalary != nil && *opts.GrossBaseSalary > 0 && employee.GrossBaseSalary != *opts.GrossBaseSalary {
		employee.GrossBaseSalary = *opts.GrossBaseSalary
		needUpdate = true
	}

	if needUpdate {
		employeeID, err := handler.agentService.SetEmployee(ctx, employee)
		if err != nil {
			if strings.Index(err.Error(), "uniq_fk_agent_fk_user") > -1 {
				return nil, v1Errors.EMPLOYEE_ALREADY_EXISTS
			}
			logger.Error(ctx, "Failed updating employee: %v", err)
			return nil, v1Errors.UPDATE_FAILED
		}
		if employeeID == 0 {
			logger.Error(ctx, "Failed updating employee: employeeID is 0")
			return nil, v1Errors.UPDATE_FAILED
		}

		employee, err = handler.agentService.GetEmployeeByID(ctx, employeeID)
		if err != nil {
			logger.Error(ctx, "Failed loading from DB: %v", err)
			return nil, v1Errors.EMPLOYEE_NOT_UPDATED
		}
		if agent == nil {
			logger.Error(ctx, "Failed loading employee is nil")
			return nil, v1Errors.EMPLOYEE_NOT_UPDATED
		}
	}

	return &V1Res{
		Agent: &Agent{
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
		User: userRes,
	}, nil
}
