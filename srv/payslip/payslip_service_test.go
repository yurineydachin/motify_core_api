package payslip_service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"motify_core_api/models"
	"motify_core_api/resources/database"
	"motify_core_api/srv/agent"
)

var (
	testAgentID    = uint64(0)
	testEmployeeID = uint64(0)
	testPayslipID  = uint64(0)
)

func getService(t *testing.T) *PayslipService {
	db, err := database.NewDbAdapter([]string{"root:123456@tcp(localhost:3306)/motify_core_api"}, "Europe/Moscow", false)
	if assert.Nil(t, err, "DB adapter init error") &&
		assert.NotNil(t, db, "DB adapter is empty") {
		return NewPayslipService(db)
	}
	return nil
}

func getAgentService(t *testing.T) *agent_service.AgentService {
	db, err := database.NewDbAdapter([]string{"root:123456@tcp(localhost:3306)/motify_core_api"}, "Europe/Moscow", false)
	if assert.Nil(t, err, "DB adapter init error") &&
		assert.NotNil(t, db, "DB adapter is empty") {
		return agent_service.NewAgentService(db)
	}
	return nil
}

func TestCreateDBAdapterAndService(t *testing.T) {
	assert.NotNil(t, getService(t), "service is nil")
}

func TestSetAgent_Create(t *testing.T) {
	service := getAgentService(t)
	agent := &models.Agent{
		Name:      "agent test",
		CompanyID: "company id",
	}
	var err error
	testAgentID, err = service.SetAgent(context.Background(), agent)
	if assert.Nil(t, err) {
		assert.Equal(t, testAgentID > 0, true)
	}
}

func TestSetEmployee_Create(t *testing.T) {
	service := getAgentService(t)
	employee := &models.Employee{
		AgentFK: testAgentID,
		Code:    "employee code",
	}
	var err error
	testEmployeeID, err = service.SetEmployee(context.Background(), employee)
	if assert.Nil(t, err) {
		assert.Equal(t, testEmployeeID > 0, true)
	}
}

func TestSetPayslip_Create(t *testing.T) {
	service := getService(t)
	payslip := &models.Payslip{
		EmployeeFK: testEmployeeID,
		Currency:   "USD",
		Amount:     2500.00,
		Data:       []byte("some data"),
	}
	var err error
	testPayslipID, err = service.SetPayslip(context.Background(), payslip)
	if assert.Nil(t, err) {
		assert.Equal(t, testPayslipID > 0, true)
	}
}

func TestGetPayslipByID(t *testing.T) {
	service := getService(t)
	payslip, err := service.GetPayslipByID(context.Background(), testPayslipID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, payslip) {
		assert.Equal(t, payslip.ID, testPayslipID)
		assert.Equal(t, payslip.Currency, "USD")
		assert.Equal(t, payslip.Amount, 2500.00)
		assert.Equal(t, payslip.Data, []byte("some data"))
	}
}

func TestSetPayslip_Update(t *testing.T) {
	service := getService(t)
	payslip := &models.Payslip{
		ID:         testPayslipID,
		EmployeeFK: testEmployeeID,
		Currency:   "RUB",
		Amount:     12500.00,
		Data:       []byte("more data"),
	}
	payslipID, err := service.SetPayslip(context.Background(), payslip)
	if assert.Nil(t, err) {
		assert.Equal(t, testPayslipID, payslipID)
	}

	payslipNew, err := service.GetPayslipByID(context.Background(), payslipID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, payslipNew) {
		assert.Equal(t, payslipNew.ID, testPayslipID)
		assert.Equal(t, payslipNew.Currency, "RUB")
		assert.Equal(t, payslipNew.Amount, 12500.00)
		assert.Equal(t, payslipNew.Data, []byte("more data"))
	}
}

func TestDeletePayslip(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeletePayslip(context.Background(), testPayslipID))
}

func TestDeleteAgentAndEmployee(t *testing.T) {
	service := getAgentService(t)
	assert.Nil(t, service.DeleteEmployee(context.Background(), testEmployeeID))
	assert.Nil(t, service.DeleteAgent(context.Background(), testAgentID))
}
