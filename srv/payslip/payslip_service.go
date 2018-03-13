package payslip_service

import (
	"motify_core_api/resources/database"
)

const (
	defaultLimit = 30
)

type PaySlipService struct {
	db database.DbAdapter
}

func NewPayslipService(db database.DbAdapter) *PaySlipService {
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
            p.id_paylist, p.fk_employee, p.currency, p.amount, p.updated_at, p.created_at,
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

func (service *PaySlipService) GetDetail(ctx context.Context, payslipID uint64) (*models.PayslipExtended, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.PayslipExtended{}
	err := service.db.Get(&res, `
        SELECT
            p.id_paylist, p.fk_employee, p.currency, p.amount, p.data, p.updated_at, p.created_at,
            a.name as company_name,
            e.role
        INNER JOIN motify_agent_employees e ON e.id_employee = p.fk_employee
        INNER JOIN motify_agents a ON a.id_agent = e.fk_agent
        FROM motify_payslip p
        WHERE p.id_paylist = ?
    `, payslipID)
	return res, err
}

func (service *PaySlipService) Create(ctx context.Context, payslip *models.Payslip) (uint64, error) {
	if payslip.EmployeeFK == 0 {
		return 0, fmt.Errorf("Insert DB exec error: no fk_user")
	}
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_payslip (fk_employee, currency, amount, data)
            VALUES (:fk_employee, :currency, :amount, :data)
        `, map[string]interface{}{
		"fk_employee": payslip.EmployeeFK,
		"currency":    payslip.Currency,
		"amount":      payslip.Amount,
		"data":        payslip.Data,
	})
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}
