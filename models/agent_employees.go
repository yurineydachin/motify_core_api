package models

type Employee struct {
	AgentFK            uint64  `db:"fk_agent"`
	UserFK             uint64  `db:"fk_user"`
	Code               string  `db:"employee_code"`
	HireDate           string  `db:"hire_date"`
	NumberOfDepandants uint    `db:"number_of_dependants"`
	GrossBaseSalary    float64 `db:"gross_base_salary"`
	Role               string  `db:"role"`
	UpdatedAt          string  `db:"updated_at"`
	CreatedAt          string  `db:"created_at"`
}
