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
