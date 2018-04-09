package models

type AgentSetting struct {
	ID                    uint64  `db:"id_setting"`
	AgentFK               uint64  `db:"s_fk_agent"`
	AgentProcessedFK      *uint64 `db:"s_fk_agent_processed"`
	UserFK                *uint64 `db:"s_fk_user"`
	Role                  string  `db:"s_role"`
	IsNotificationEnabled bool    `db:"s_notifications_enabled"`
	IsMainAgent           bool    `db:"s_is_main_agent"`
	UpdatedAt             string  `db:"s_updated_at"`
	CreatedAt             string  `db:"s_created_at"`
}

func (s *AgentSetting) ToArgs() map[string]interface{} {
	res := map[string]interface{}{
		"id_setting":              s.ID,
		"s_fk_agent":              s.AgentFK,
		"s_role":                  s.Role,
		"s_notifications_enabled": s.IsNotificationEnabled,
		"s_is_main_agent":         s.IsMainAgent,
	}
	if s.UserFK != nil && *s.UserFK > 0 {
		res["s_fk_user"] = *s.UserFK
	}
	if s.AgentProcessedFK != nil && *s.AgentProcessedFK > 0 {
		res["s_fk_agent_processed"] = *s.AgentProcessedFK
	}
	return res
}

type AgentSettingLeft struct {
	ID                    *uint64 `db:"id_setting"`
	AgentFK               *uint64 `db:"s_fk_agent"`
	AgentProcessedFK      *uint64 `db:"s_fk_agent_processed"`
	UserFK                *uint64 `db:"s_fk_user"`
	Role                  *string `db:"s_role"`
	IsNotificationEnabled *bool   `db:"s_notifications_enabled"`
	IsMainAgent           *bool   `db:"s_is_main_agent"`
	UpdatedAt             *string `db:"s_updated_at"`
	CreatedAt             *string `db:"s_created_at"`
}

func (s *AgentSettingLeft) ToAgentSetting() *AgentSetting {
	if s == nil || s.ID == nil || *s.ID == 0 {
		return nil
	}
	res := &AgentSetting{
		ID:               *s.ID,
		AgentProcessedFK: s.AgentProcessedFK,
		UserFK:           s.UserFK,
	}
	if s.AgentFK != nil {
		res.AgentFK = *s.AgentFK
	}
	if s.Role != nil {
		res.Role = *s.Role
	}
	if s.IsNotificationEnabled != nil {
		res.IsNotificationEnabled = *s.IsNotificationEnabled
	}
	if s.IsMainAgent != nil {
		res.IsMainAgent = *s.IsMainAgent
	}
	if s.UpdatedAt != nil {
		res.UpdatedAt = *s.UpdatedAt
	}
	if s.CreatedAt != nil {
		res.CreatedAt = *s.CreatedAt
	}
	return res
}
