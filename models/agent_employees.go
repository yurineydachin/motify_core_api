package models

type Employee struct {
	ID                 uint64  `db:"id_employee"`
	AgentFK            uint64  `db:"e_fk_agent"`
	UserFK             *uint64 `db:"e_fk_user"`
	Code               string  `db:"e_code"`
	Name               string  `db:"e_name"`
	Role               string  `db:"e_role"`
	Email              string  `db:"e_email"`
	HireDate           string  `db:"e_hire_date"`
	NumberOfDepandants uint    `db:"e_number_of_dependants"`
	GrossBaseSalary    float64 `db:"e_gross_base_salary"`
	UpdatedAt          string  `db:"e_updated_at"`
	CreatedAt          string  `db:"e_created_at"`
}

func (emp *Employee) ToArgs() map[string]interface{} {
	res := map[string]interface{}{
		"id_employee":            emp.ID,
		"e_fk_agent":             emp.AgentFK,
		"e_code":                 emp.Code,
		"e_name":                 emp.Name,
		"e_email":                emp.Email,
		"e_hire_date":            emp.HireDate,
		"e_number_of_dependants": emp.NumberOfDepandants,
		"e_gross_base_salary":    emp.GrossBaseSalary,
		"e_role":                 emp.Role,
	}
	if emp.UserFK != nil && *emp.UserFK > 0 {
		res["fk_user"] = *emp.UserFK
	}
	return res
}
