package models

type Agent struct {
	ID          uint64 `db:"id_agent"`
	Name        string `db:"a_name"`
	CompanyID   string `db:"a_company_id"`
	Description string `db:"a_description"`
	Logo        string `db:"a_logo"`
	Background  string `db:"a_bg_image"`
	Address     string `db:"a_address"`
	Phone       string `db:"a_phone"`
	Email       string `db:"a_email"`
	Site        string `db:"a_site"`
	UpdatedAt   string `db:"a_updated_at"`
	CreatedAt   string `db:"a_created_at"`
}

func (agent *Agent) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_agent":      agent.ID,
		"a_name":        agent.Name,
		"a_company_id":  agent.CompanyID,
		"a_description": agent.Description,
		"a_logo":        agent.Logo,
		"a_bg_image":    agent.Background,
		"a_address":     agent.Address,
		"a_phone":       agent.Phone,
		"a_email":       agent.Email,
		"a_site":        agent.Site,
	}
}

type AgentWithSetting struct {
	Agent
	AgentSetting
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
