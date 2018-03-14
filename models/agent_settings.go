package models

type AgentSetting struct {
	ID                    uint64  `db:"id_setting"`
	AgentFK               uint64  `db:"fk_agent"`
	AgentProcessedFK      *uint64 `db:"fk_agent_processed"`
	UserFK                *uint64 `db:"fk_user"`
	Role                  string  `db:"role"`
	IsNotificationEnabled bool    `db:"notifications_enabled"`
	IsMainAgent           bool    `db:"is_main_agent"`
	UpdatedAt             string  `db:"updated_at"`
	CreatedAt             string  `db:"created_at"`
}

func (s *AgentSetting) ToArgs() map[string]interface{} {
	res := map[string]interface{}{
		"id_setting":            s.ID,
		"fk_agent":              s.AgentFK,
		"role":                  s.Role,
		"notifications_enabled": s.IsNotificationEnabled,
		"is_main_agent":         s.IsMainAgent,
	}
	if s.UserFK != nil && *s.UserFK > 0 {
		res["fk_user"] = *s.UserFK
	}
	if s.AgentProcessedFK != nil && *s.AgentProcessedFK > 0 {
		res["fk_agent_processed"] = *s.AgentProcessedFK
	}
	return res
}
