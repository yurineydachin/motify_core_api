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

type PayslipAgentEmployee struct {
	*models.Payslip
	*models.Agent
	*models.Employee
}

func (service *PayslipService) GetListByUserID(ctx context.Context, userID, limit, offset uint64) ([]PayslipAgentEmployee, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []PayslipAgentEmployee{}
	err := service.db.Select(&res, `
        SELECT
            id_payslip, p_fk_employee, p_title, p_currency, p_amount, p_updated_at, p_created_at,
            id_agent, a_name, a_company_id, a_description, a_logo, a_bg_image, a_address, a_phone, a_email, a_site, a_updated_at, a_created_at,
            id_employee, e_fk_agent, e_fk_user, e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role, e_updated_at, e_created_at
        FROM motify_payslip
        INNER JOIN motify_agent_employees ON id_employee = p_fk_employee
        INNER JOIN motify_agents ON id_agent = e_fk_agent
        WHERE e_fk_user = ?
        ORDER BY p_created_at DESC
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
            p.id_payslip, p.p_fk_employee, p.p_title, p.p_currency, p.p_amount, p.p_updated_at, p.p_created_at
        FROM motify_payslip p
        WHERE p.p_fk_employee = ?
        ORDER BY p.p_created_at DESC
        LIMIT ?
        OFFSET ?
    `, employeeID, limit, offset)
	return res, err
}

func (service *PayslipService) GetPayslipByID(ctx context.Context, modelID uint64) (*models.Payslip, error) {
	res := models.Payslip{}
	err := service.db.Get(&res, `
        SELECT id_payslip, p_fk_employee, p_title, p_currency, p_amount, p_data, p_updated_at, p_created_at
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
            INSERT INTO motify_payslip (p_fk_employee, p_title, p_currency, p_amount, p_data)
            VALUES (:p_fk_employee, :p_title, :p_currency, :p_amount, :p_data)
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
                p_fk_employee = :p_fk_employee,
                p_title = :p_title,
                p_currency = :p_currency,
                p_amount = :p_amount,
                p_data = :p_data
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
