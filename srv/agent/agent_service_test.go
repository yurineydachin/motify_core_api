package agent_service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"motify_core_api/models"
	"motify_core_api/resources/database"
)

var (
	testAgentID    = uint64(0)
	testAgentIDStr = "0"
	testSettingID  = uint64(0)
	testEmployeeID = uint64(0)
)

func getService(t *testing.T) *AgentService {
	db, err := database.NewDbAdapter([]string{"root:123456@tcp(localhost:3306)/motify_core_api"}, "Europe/Moscow", false)
	if assert.Nil(t, err, "DB adapter init error") &&
		assert.NotNil(t, db, "DB adapter is empty") {
		return NewAgentService(db)
	}
	return nil
}

func TestCreateDBAdapterAndService(t *testing.T) {
	assert.NotNil(t, getService(t), "service is nil")
}

func TestSetAgent_Create(t *testing.T) {
	service := getService(t)
	agent := &models.Agent{
		Name:        "agent test",
		CompanyID:   "company id",
		Description: "desc test",
		Logo:        "logo test",
		Background:  "bg_image",
		Phone:       "phone test",
		Email:       "email_test@text.com",
		Address:     "address test",
		Site:        "site test",
	}
	var err error
	testAgentID, err = service.SetAgent(context.Background(), agent)
	testAgentIDStr = fmt.Sprintf("%d", testAgentID)
	if assert.Nil(t, err) {
		assert.Equal(t, testAgentID > 0, true)
	}
}

func TestGetAgentByID(t *testing.T) {
	service := getService(t)
	agent, err := service.GetAgentByID(context.Background(), testAgentID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, agent) {
		assert.Equal(t, agent.ID, testAgentID)
		assert.Equal(t, agent.Name, "agent test")
	}
}

func TestSetAgent_Update(t *testing.T) {
	service := getService(t)
	agent := &models.Agent{
		ID:          testAgentID,
		Name:        "agent " + testAgentIDStr,
		CompanyID:   "company id " + testAgentIDStr,
		Description: "desc " + testAgentIDStr,
		Logo:        "logo " + testAgentIDStr,
		Phone:       "phone " + testAgentIDStr,
		Email:       "email_" + testAgentIDStr + "@text.com",
	}
	agentID, err := service.SetAgent(context.Background(), agent)
	if assert.Nil(t, err) {
		assert.Equal(t, testAgentID, agent.ID)
	}

	agentNew, err := service.GetAgentByID(context.Background(), agentID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, agentNew) {
		assert.Equal(t, agentNew.ID, testAgentID)
		assert.Equal(t, agentNew.Name, "agent "+testAgentIDStr)
	}
}

func TestSetEmployee_Create(t *testing.T) {
	service := getService(t)
	employee := &models.Employee{
		AgentFK:            testAgentID,
		Code:               "employee code",
		HireDate:           "hire date",
		NumberOfDepandants: 1,
		GrossBaseSalary:    1234.00,
		Role:               "role",
	}
	var err error
	testEmployeeID, err = service.SetEmployee(context.Background(), employee)
	if assert.Nil(t, err) {
		assert.Equal(t, testEmployeeID > 0, true)
	}
}

func TestGetEmployeeByID(t *testing.T) {
	service := getService(t)
	employee, err := service.GetEmployeeByID(context.Background(), testEmployeeID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, employee) {
		assert.Equal(t, employee.ID, testEmployeeID)
		assert.Equal(t, employee.Code, "employee code")
	}
}

func TestSetEmployee_Update(t *testing.T) {
	service := getService(t)
	employee := &models.Employee{
		ID:                 testEmployeeID,
		AgentFK:            testAgentID,
		Code:               "employee code 2",
		HireDate:           "hire date 2",
		NumberOfDepandants: 2,
		GrossBaseSalary:    2345.00,
		Role:               "role 2",
	}
	employeeID, err := service.SetEmployee(context.Background(), employee)
	if assert.Nil(t, err) {
		assert.Equal(t, testEmployeeID, employee.ID)
	}

	employeeNew, err := service.GetEmployeeByID(context.Background(), employeeID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, employeeNew) {
		assert.Equal(t, employeeNew.ID, testEmployeeID)
		assert.Equal(t, employeeNew.Code, "employee code 2")
	}
}

func TestSetSetting_Create(t *testing.T) {
	service := getService(t)
	setting := &models.AgentSetting{
		AgentFK: testAgentID,
		Role:    "role setting",
		IsNotificationEnabled: false,
		IsMainAgent:           false,
	}
	var err error
	testSettingID, err = service.SetSetting(context.Background(), setting)
	if assert.Nil(t, err) {
		assert.Equal(t, testSettingID > 0, true)
	}
}

func TestGetSettingByID(t *testing.T) {
	service := getService(t)
	setting, err := service.GetSettingByID(context.Background(), testSettingID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, setting) {
		assert.Equal(t, setting.ID, testSettingID)
		assert.Equal(t, setting.Role, "role setting")
		assert.Equal(t, setting.IsNotificationEnabled, false)
		assert.Equal(t, setting.IsMainAgent, false)
	}
}

func TestSetSetting_Update(t *testing.T) {
	service := getService(t)
	setting := &models.AgentSetting{
		ID:      testSettingID,
		AgentFK: testAgentID,
		Role:    "role setting 2",
		IsNotificationEnabled: true,
		IsMainAgent:           true,
	}
	settingID, err := service.SetSetting(context.Background(), setting)
	if assert.Nil(t, err) {
		assert.Equal(t, testSettingID, setting.ID)
	}

	settingNew, err := service.GetSettingByID(context.Background(), settingID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, settingNew) {
		assert.Equal(t, settingNew.ID, testSettingID)
		assert.Equal(t, settingNew.Role, "role setting 2")
		assert.Equal(t, settingNew.IsNotificationEnabled, true)
		assert.Equal(t, settingNew.IsMainAgent, true)
	}
}

func TestDeleteSetting(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeleteSetting(context.Background(), testSettingID))
}

func TestDeleteEmployee(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeleteEmployee(context.Background(), testEmployeeID))
}

func TestDeleteAgent(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeleteAgent(context.Background(), testAgentID))
}
