package integration_service

import (
	"context"
	"fmt"

	"motify_core_api/models"
	"motify_core_api/resources/database"
)

type IntegrationService struct {
	db *database.DbAdapter
}

func NewIntegrationService(db *database.DbAdapter) *IntegrationService {
	return &IntegrationService{
		db: db,
	}
}

func (service *IntegrationService) GetIntegrationByID(ctx context.Context, modelID uint64) (*models.Integration, error) {
	res := models.Integration{}
	err := service.db.Get(&res, `
        SELECT id_integration, i_hash, i_updated_at, i_created_at
        FROM motify_integrations WHERE id_integration = ?
    `, modelID)
	return &res, err
}

func (service *IntegrationService) SetIntegration(ctx context.Context, model *models.Integration) (uint64, error) {
	if model.ID > 0 {
		return service.updateIntegration(ctx, model)
	}
	return service.createIntegration(ctx, model)
}

func (service *IntegrationService) createIntegration(ctx context.Context, model *models.Integration) (uint64, error) {
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_integrations (i_hash)
            VALUES (:i_hash)
        `, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}

func (service *IntegrationService) updateIntegration(ctx context.Context, model *models.Integration) (uint64, error) {
	updateRes, err := service.db.Exec(`
            UPDATE motify_integrations SET
                i_hash = :i_hash
            WHERE
                id_integration = :id_integration
        `, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Update DB exec error: %v", err)
	}
	rowsCount, err := updateRes.RowsAffected()
	if rowsCount == 0 {
		return 0, fmt.Errorf("Update DB exec error: nothing changed")
	}
	return model.ID, nil
}

func (service *IntegrationService) DeleteIntegration(ctx context.Context, modelID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_integrations
            WHERE
                id_integration = :id_integration
        `, map[string]interface{}{
		"id_integration": modelID,
	})
	if err != nil {
		return fmt.Errorf("Insert DB exec error: %v", err)
	}
	rowsCount, err := deleteRes.RowsAffected()
	if rowsCount == 0 {
		return fmt.Errorf("Delete DB exec error: nothing changed")
	}
	return nil
}
