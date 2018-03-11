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
