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

type PaySlipService struct {
	db *database.DbAdapter
}

func NewPaySlipService(db *database.DbAdapter) *PaySlipService {
	return &PaySlipService{
		db: db,
	}
}

func (service *PaySlipService) GetListByUserID(ctx context.Context, userID, limit, offset uint64) ([]*models.PayslipExtended, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.PayslipExtended{}
	err := service.db.Select(&res, `
        SELECT
            p.id_payslip, p.fk_employee, p.currency, p.amount, p.updated_at, p.created_at,
            a.name as company_name,
            e.role
        INNER JOIN motify_agent_employees e ON e.id_employee = p.fk_employee
        INNER JOIN motify_agents a ON a.id_agent = e.fk_agent
        FROM motify_payslip p
        WHERE e.fk_user = ?
        LIMIT ?
        OFFSET ?
        ORDER BY p.created_at DESC
    `, userID, limit, offset)
	return res, err
}

func (service *PaySlipService) GetDetail(ctx context.Context, modelID uint64) (*models.PayslipExtended, error) {
	res := &models.PayslipExtended{}
	err := service.db.Get(&res, `
        SELECT
            p.id_payslip, p.fk_employee, p.currency, p.amount, p.data, p.updated_at, p.created_at,
            a.name as company_name,
            e.role
        INNER JOIN motify_agent_employees e ON e.id_employee = p.fk_employee
        INNER JOIN motify_agents a ON a.id_agent = e.fk_agent
        FROM motify_payslip p
        WHERE p.id_payslip = ?
    `, modelID)
	return res, err
}

func (service *PaySlipService) GetPayslipByID(ctx context.Context, modelID uint64) (*models.Payslip, error) {
	res := models.Payslip{}
	err := service.db.Get(&res, `
        SELECT id_payslip, fk_employee, currency, amount, data, updated_at, created_at
        FROM motify_payslip WHERE id_payslip = ?
    `, modelID)
	return &res, err
}

func (service *PaySlipService) SetPayslip(ctx context.Context, model *models.Payslip) (uint64, error) {
	if model.ID > 0 {
		return service.updatePayslip(ctx, model)
	}
	return service.createPayslip(ctx, model)
}

func (service *PaySlipService) createPayslip(ctx context.Context, model *models.Payslip) (uint64, error) {
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

func (service *PaySlipService) updatePayslip(ctx context.Context, model *models.Payslip) (uint64, error) {
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

func (service *PaySlipService) DeletePayslip(ctx context.Context, modelID uint64) error {
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
