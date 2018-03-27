package integration_service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"motify_core_api/models"
	"motify_core_api/resources/database"
)

var (
	testIntegrationID    = uint64(0)
	testIntegrationIDStr = "0"
)

func getService(t *testing.T) *IntegrationService {
	db, err := database.NewDbAdapter([]string{"root:123456@tcp(localhost:3306)/motify_core_api"}, "Europe/Moscow", false)
	if assert.Nil(t, err, "DB adapter init error") &&
		assert.NotNil(t, db, "DB adapter is empty") {
		return NewIntegrationService(db)
	}
	return nil
}

func TestCreateDBAdapterAndService(t *testing.T) {
	assert.NotNil(t, getService(t), "service is nil")
}

func TestSetIntegration_Create(t *testing.T) {
	service := getService(t)
	integration := &models.Integration{
		Hash: "test 1234",
	}
	var err error
	testIntegrationID, err = service.SetIntegration(context.Background(), integration)
	testIntegrationIDStr = fmt.Sprintf("%d", testIntegrationID)
	if assert.Nil(t, err) {
		assert.Equal(t, testIntegrationID > 0, true)
	}
}

func TestGetIntegrationByID(t *testing.T) {
	service := getService(t)
	integration, err := service.GetIntegrationByID(context.Background(), testIntegrationID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, integration) {
		assert.Equal(t, integration.ID, testIntegrationID)
		assert.Equal(t, integration.Hash, "test 1234")
	}
}

func TestSetIntegration_Update(t *testing.T) {
	service := getService(t)
	integration := &models.Integration{
		ID:   testIntegrationID,
		Hash: "test " + testIntegrationIDStr,
	}
	integrationID, err := service.SetIntegration(context.Background(), integration)
	if assert.Nil(t, err) {
		assert.Equal(t, testIntegrationID, integration.ID)
	}

	integrationNew, err := service.GetIntegrationByID(context.Background(), integrationID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, integrationNew) {
		assert.Equal(t, integrationNew.ID, testIntegrationID)
		assert.Equal(t, integrationNew.Hash, "test "+testIntegrationIDStr)
	}
}

func TestDeleteIntegration(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeleteIntegration(context.Background(), testIntegrationID))
}
