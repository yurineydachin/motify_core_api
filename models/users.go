package models

type User struct {
	ID          uint64 `db:"id_user"`
	Name        string `db:"name"`
	Short       string `db:"p_description"`
	Description string `db:"description"`
	Awatar      string `db:"awatar"`
	Phone       string `db:"phone"`
	Email       string `db:"email"`
	UpdatedAt   string `db:"updated_at"`
	CreatedAt   string `db:"created_at"`
}

func (user *User) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_user":       user.ID,
		"user_name":     user.Name,
		"p_description": user.Short,
		"description":   user.Description,
		"awatar":        user.Awatar,
		"phone":         user.Phone,
		"email":         user.Email,
	}
}
