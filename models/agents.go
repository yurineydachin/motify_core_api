package models

type Agent struct {
	ID          uint64 `db:"id_agent"`
	Name        string `db:"name"`
	CompanyID   string `db:"company_id"`
	Description string `db:"description"`
	Logo        string `db:"logo"`
	Background  string `db:"bg_image"`
	Address     string `db:"address"`
	Phone       string `db:"phone"`
	Email       string `db:"email"`
	Site        string `db:"site"`
	UpdatedAt   string `db:"updated_at"`
	CreatedAt   string `db:"created_at"`
}

type AgentWithSettings struct {
	Agent
	AgentSettings
	/*
		Role                  string `db:"role"`
		IsNotificationEnabled bool   `db:"notifications_enabled"`
		IsMainAgent           bool   `db:"is_main_agent"`
	*/
}

type AgentWithEmployee struct {
	Agent
	Employee
	/*
		UserFK             uint64  `db:"fk_user"`
		Code               string  `db:"employee_code"`
		HireDate           string  `db:"hire_date"`
		NumberOfDepandants uint    `db:"number_of_dependants"`
		GrossBaseSalary    float64 `db:"gross_base_salary"`
		Role               string  `db:"role"`
	*/
}
