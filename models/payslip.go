package models

type Payslip struct {
	ID         uint64  `db:"id_payslip"`
	EmployeeFK uint64  `db:"fk_employee"`
	Title      string  `db:"title"`
	Currency   string  `db:"currency"`
	Amount     float64 `db:"amount"`
	Data       []byte  `db:"data"`
	UpdateAt   string  `db:"updated_at"`
	CreatedAt  string  `db:"created_at"`
}

func (p *Payslip) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_payslip":  p.ID,
		"fk_employee": p.EmployeeFK,
		"title":       p.Title,
		"currency":    p.Currency,
		"amount":      p.Amount,
		"data":        p.Data,
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
