package models

type AgentSettings struct {
	AgentFK               uint64 `db:"fk_agent"`
	UserFK                uint64 `db:"fk_user"`
	Role                  string `db:"role"`
	IsNotificationEnabled bool   `db:"notifications_enabled"`
	IsMainAgent           bool   `db:"is_main_agent"`
	UpdatedAt             string `db:"updated_at"`
	CreatedAt             string `db:"created_at"`
}
