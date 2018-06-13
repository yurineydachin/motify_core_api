package device_service

import (
	"context"
	"fmt"

	"motify_core_api/models"
	"motify_core_api/resources/database"
)

type Service struct {
	db *database.DbAdapter
}

func New(db *database.DbAdapter) *Service {
	return &Service{
		db: db,
	}
}

func (service *Service) GetListByEmployeeID(ctx context.Context, modelID uint64) ([]*models.Device, error) {
	res := []*models.Device{}
	err := service.db.Select(&res, `
        SELECT id_device, d_fk_user, d_token, d_updated_at, d_created_at
        FROM motify_device
	INNER JOIN motify_users ON id_user = d_fk_user
	INNER JOIN motify_agent_employees
	WHERE id_employee = ?
    `, modelID)
	return res, err
}

func (service *Service) Set(ctx context.Context, model *models.Device) (uint64, error) {
	if model.ID > 0 {
		return service.update(ctx, model)
	}
	return service.create(ctx, model)
}

func (service *Service) create(ctx context.Context, model *models.Device) (uint64, error) {
	if model.UserFK == 0 {
		return 0, fmt.Errorf("Insert DB exec error: no fk_user")
	}
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_device (d_fk_user, d_token)
            VALUES (:d_fk_user, :d_token)
        `, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	if id == 0 {
		return 0, fmt.Errorf("Insert DB error: id = 0")
	}
	return uint64(id), err
}

func (service *Service) update(ctx context.Context, model *models.Device) (uint64, error) {
	updateRes, err := service.db.Exec(`
            UPDATE motify_device SET
                d_token = :d_token
            WHERE id_device = :id_device
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

func (service *Service) DeleteByID(ctx context.Context, modelID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_device
            WHERE id_device = :id_device
        `, map[string]interface{}{
		"id_device": modelID,
	})
	if err != nil {
		return fmt.Errorf("Delete DB exec error: %v", err)
	}
	rowsCount, err := deleteRes.RowsAffected()
	if rowsCount == 0 {
		return fmt.Errorf("Delete DB exec error: nothing changed")
	}
	return nil
}
