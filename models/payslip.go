package models

type Payslip struct {
	ID         uint64  `db:"id_paylist"`
	EmployeeFK uint64  `db:"fk_employee"`
	Currency   string  `db:"currency"`
	Amount     float64 `db:"amount"`
	Data       []byte  `db:"data"`
	UpdateAt   string  `db:"updated_at"`
	CreatedAt  string  `db:"created_at"`
}

type PayslipExtended struct {
	Payslip

	CompanyName string `db:"company_name"`
	Role        string `db:"role"`
}

func (ext *PayslipExtended) ToPayslip() *Payslip {
	return &ext.Payslip
}
