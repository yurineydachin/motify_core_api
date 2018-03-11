package models

type Agent struct {
	ID          uint64 `db:"id_agent"`
	Name        string `db:"name"`
	CompanyID   string `db:"company_id"`
	Description string `db:"description"`
	Logo        string `db:"logo"`
	Background  string `db:"bg_image"`
	Address     string `db:"address"`
	Phone       string `db:"phone"`
	Email       string `db:"email"`
	Site        string `db:"site"`
	UpdatedAt   string `db:"updated_at"`
	CreatedAt   string `db:"created_at"`
}
