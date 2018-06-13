package models

type Device struct {
	ID        uint64 `db:"id_push_device"`
	UserFK    uint64 `db:"d_fk_user"`
	Device    string `db:"d_device"`
	Token     string `db:"d_token"`
	UpdatedAt string `db:"d_updated_at"`
	CreatedAt string `db:"d_created_at"`
}

func (pd Device) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_device": pd.ID,
		"d_fk_user": pd.UserFK,
		"d_token":   pd.Token,
	}
}
