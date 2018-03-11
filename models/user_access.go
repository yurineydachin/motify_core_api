package models

const (
	UserAccessEmail = "email"
	UserAccessFb    = "fb"
	UserAccessVk    = "vk"
)

type UserAccess struct {
	ID        uint64 `db:"id_user_access"`
	UserFK    uint64 `db:"fk_user"`
	Type      string `db:"type_access"`
	Email     string `db:"email"`
	Phone     string `db:"phone"`
	Password  string `db:"password"`
	UpdatedAt string `db:"updated_at"`
	CreatedAt string `db:"created_at"`
	IsHashed bool `db:"-"`
}
