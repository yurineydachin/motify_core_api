package models

type PushDevice struct {
	ID        uint64 `db:"id_push_device"`
	UserFK    uint64 `db:"pd_fk_user"`
	Token     string `db:"pd_token"`
	UpdatedAt string `db:"pd_updated_at"`
	CreatedAt string `db:"pd_created_at"`
}

func (pd PushDevice) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_push_device": pd.ID,
		"pd_fk_user":     pd.UserFK,
		"pd_token":       pd.Token,
	}
}
