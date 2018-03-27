package models

type User struct {
	ID            uint64  `db:"id_user"`
	IntegrationFK *uint64 `db:"u_fk_integration"`
	Name          string  `db:"u_name"`
	Short         string  `db:"u_short"`
	Description   string  `db:"u_description"`
	Awatar        string  `db:"u_awatar"`
	Phone         string  `db:"u_phone"`
	Email         string  `db:"u_email"`
	UpdatedAt     string  `db:"u_updated_at"`
	CreatedAt     string  `db:"u_created_at"`
}

func (user *User) ToArgs() map[string]interface{} {
	res := map[string]interface{}{
		"id_user":       user.ID,
		"u_name":        user.Name,
		"u_short":       user.Short,
		"u_description": user.Description,
		"u_awatar":      user.Awatar,
		"u_phone":       user.Phone,
		"u_email":       user.Email,
	}
	if user.IntegrationFK != nil && *user.IntegrationFK > 0 {
		res["u_fk_integration"] = *user.IntegrationFK
	}
	return res
}
