package models

type Payslip struct {
	ID         uint64  `db:"id_payslip"`
	EmployeeFK uint64  `db:"p_fk_employee"`
	Title      string  `db:"p_title"`
	Currency   string  `db:"p_currency"`
	Amount     float64 `db:"p_amount"`
	Data       []byte  `db:"p_data"`
	UpdateAt   string  `db:"p_updated_at"`
	CreatedAt  string  `db:"p_created_at"`
}

func (p *Payslip) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_payslip":    p.ID,
		"p_fk_employee": p.EmployeeFK,
		"p_title":       p.Title,
		"p_currency":    p.Currency,
		"p_amount":      p.Amount,
		"p_data":        p.Data,
	}
}

type PayslipExtended struct {
	Payslip

	CompanyName string `db:"company_name"`
	Role        string `db:"role"`
}

func (ext *PayslipExtended) ToPayslip() *Payslip {
	return &ext.Payslip
}
