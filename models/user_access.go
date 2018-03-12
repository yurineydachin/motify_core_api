package models

import (
	"motify_core_api/utils"
)

const (
	UserAccessEmail = "email"
	UserAccessFb    = "fb"
	UserAccessVk    = "vk"
)

type UserAccess struct {
	ID               uint64 `db:"id_user_access"`
	UserFK           uint64 `db:"fk_user"`
	Type             string `db:"type_access"`
	IsHashedEmail    bool   `db:"-"`
	Email            string `db:"email"`
	IsHashedPhone    bool   `db:"-"`
	Phone            string `db:"phone"`
	IsHashedPassword bool   `db:"-"`
	Password         string `db:"password"`
	UpdatedAt        string `db:"updated_at"`
	CreatedAt        string `db:"created_at"`
}

func (access *UserAccess) MarkAllHashed() {
	access.IsHashedEmail = true
	access.IsHashedPhone = true
	access.IsHashedPassword = true
}

func (access UserAccess) ToArgs() map[string]interface{} {
	args := map[string]interface{}{
		"id_user_access": access.ID,
		"fk_user":        access.UserFK,
		"type_access":    access.Type,
	}
	if access.IsHashedEmail {
		args["email"] = access.Email
	} else {
		args["email"] = utils.Hash(access.Email)
	}
	if access.IsHashedPhone {
		args["phone"] = access.Phone
	} else {
		args["phone"] = utils.Hash(access.Phone)
	}
	if access.IsHashedPassword {
		args["password"] = access.Password
	} else {
		args["password"] = utils.Hash(access.Password)
	}
	return args
}
