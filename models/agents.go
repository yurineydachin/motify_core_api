package models

type Agent struct {
	ID            uint64 `db:"id_agent"`
	IntegrationFK uint64 `db:"a_fk_integration"`
	Name          string `db:"a_name"`
	CompanyID     string `db:"a_company_id"`
	Description   string `db:"a_description"`
	Logo          string `db:"a_logo"`
	Background    string `db:"a_bg_image"`
	Address       string `db:"a_address"`
	Phone         string `db:"a_phone"`
	Email         string `db:"a_email"`
	Site          string `db:"a_site"`
	UpdatedAt     string `db:"a_updated_at"`
	CreatedAt     string `db:"a_created_at"`
}

func (agent *Agent) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_agent":         agent.ID,
		"a_fk_integration": agent.IntegrationFK,
		"a_name":           agent.Name,
		"a_company_id":     agent.CompanyID,
		"a_description":    agent.Description,
		"a_logo":           agent.Logo,
		"a_bg_image":       agent.Background,
		"a_address":        agent.Address,
		"a_phone":          agent.Phone,
		"a_email":          agent.Email,
		"a_site":           agent.Site,
	}
}
