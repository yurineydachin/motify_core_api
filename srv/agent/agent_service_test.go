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
	testAgentID         = uint64(0)
	testAgentIDStr      = "0"
	testAgentSettingsID = uint64(0)
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

func TestDeleteAgent(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeleteAgent(context.Background(), testAgentID))
}
