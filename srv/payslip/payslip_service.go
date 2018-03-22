package payslip_service

import (
	"context"
	"fmt"

	"motify_core_api/models"
	"motify_core_api/resources/database"
)

const (
	defaultLimit = 30
)

type PayslipService struct {
	db *database.DbAdapter
}

func NewPayslipService(db *database.DbAdapter) *PayslipService {
	return &PayslipService{
		db: db,
	}
}

func (service *PayslipService) GetListByUserID(ctx context.Context, userID, limit, offset uint64) ([]*models.Payslip, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.Payslip{}
	err := service.db.Select(&res, `
        SELECT
            p.id_payslip, p.fk_employee, p.currency, p.amount, p.updated_at, p.created_at
        FROM motify_payslip p
        INNER JOIN motify_agent_employees e ON e.id_employee = p.fk_employee
        WHERE e.fk_user = ?
        ORDER BY p.created_at DESC
        LIMIT ?
        OFFSET ?
    `, userID, limit, offset)
	return res, err
}

func (service *PayslipService) GetListByEmployeeID(ctx context.Context, employeeID, limit, offset uint64) ([]*models.Payslip, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.Payslip{}
	err := service.db.Select(&res, `
        SELECT
            p.id_payslip, p.fk_employee, p.currency, p.amount, p.updated_at, p.created_at
        FROM motify_payslip p
        WHERE p.fk_employee = ?
        ORDER BY p.created_at DESC
        LIMIT ?
        OFFSET ?
    `, employeeID, limit, offset)
	return res, err
}

func (service *PayslipService) GetPayslipByID(ctx context.Context, modelID uint64) (*models.Payslip, error) {
	res := models.Payslip{}
	err := service.db.Get(&res, `
        SELECT id_payslip, fk_employee, currency, amount, data, updated_at, created_at
        FROM motify_payslip WHERE id_payslip = ?
    `, modelID)
	return &res, err
}

func (service *PayslipService) SetPayslip(ctx context.Context, model *models.Payslip) (uint64, error) {
	if model.ID > 0 {
		return service.updatePayslip(ctx, model)
	}
	return service.createPayslip(ctx, model)
}

func (service *PayslipService) createPayslip(ctx context.Context, model *models.Payslip) (uint64, error) {
	if model.EmployeeFK == 0 {
		return 0, fmt.Errorf("Insert DB exec error: no fk_employee")
	}
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_payslip (fk_employee, currency, amount, data)
            VALUES (:fk_employee, :currency, :amount, :data)
        `, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}

func (service *PayslipService) updatePayslip(ctx context.Context, model *models.Payslip) (uint64, error) {
	updateRes, err := service.db.Exec(`
            UPDATE motify_payslip SET
                fk_employee = :fk_employee,
                currency = :currency,
                amount = :amount,
                data = :data
            WHERE
                id_payslip = :id_payslip
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

func (service *PayslipService) DeletePayslip(ctx context.Context, modelID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_payslip
            WHERE
                id_payslip = :id_payslip
        `, map[string]interface{}{
		"id_payslip": modelID,
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
