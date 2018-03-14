package models

type Employee struct {
	ID                 uint64  `db:"id_employee"`
	AgentFK            uint64  `db:"fk_agent"`
	UserFK             *uint64 `db:"fk_user"`
	Code               string  `db:"employee_code"`
	HireDate           string  `db:"hire_date"`
	NumberOfDepandants uint    `db:"number_of_dependants"`
	GrossBaseSalary    float64 `db:"gross_base_salary"`
	Role               string  `db:"role"`
	UpdatedAt          string  `db:"updated_at"`
	CreatedAt          string  `db:"created_at"`
}

func (emp *Employee) ToArgs() map[string]interface{} {
	res := map[string]interface{}{
		"id_employee":          emp.ID,
		"fk_agent":             emp.AgentFK,
		"employee_code":        emp.Code,
		"hire_date":            emp.HireDate,
		"number_of_dependants": emp.NumberOfDepandants,
		"gross_base_salary":    emp.GrossBaseSalary,
		"role":                 emp.Role,
	}
	if emp.UserFK != nil && *emp.UserFK > 0 {
		res["fk_user"] = *emp.UserFK
	}
	return res
}
