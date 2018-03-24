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
	ID               uint64  `db:"id_user_access"`
	UserFK           uint64  `db:"ua_fk_user"`
	Type             string  `db:"ua_type"`
	IsHashedEmail    bool    `db:"-"`
	Email            *string `db:"ua_email"`
	IsHashedPhone    bool    `db:"-"`
	Phone            *string `db:"ua_phone"`
	IsHashedPassword bool    `db:"-"`
	Password         string  `db:"ua_password"`
	UpdatedAt        string  `db:"ua_updated_at"`
	CreatedAt        string  `db:"ua_created_at"`
}

func (access *UserAccess) SetEmail(value string) {
	access.Email = &value
	access.IsHashedEmail = false
}

func (access *UserAccess) SetPhone(value string) {
	access.Phone = &value
	access.IsHashedPhone = false
}

func (access *UserAccess) SetPassword(value string) {
	access.Password = value
	access.IsHashedPassword = false
}

func (access *UserAccess) MarkAllHashed() {
	access.IsHashedEmail = true
	access.IsHashedPhone = true
	access.IsHashedPassword = true
}

func (access UserAccess) ToArgs() map[string]interface{} {
	args := map[string]interface{}{
		"id_user_access": access.ID,
		"ua_fk_user":     access.UserFK,
		"ua_type":        access.Type,
	}
	if access.Email != nil {
		if access.IsHashedEmail {
			args["ua_email"] = *access.Email
		} else {
			args["ua_email"] = utils.Hash(*access.Email)
		}
	}
	if access.Phone != nil {
		if access.IsHashedPhone {
			args["ua_phone"] = *access.Phone
		} else {
			args["ua_phone"] = utils.Hash(*access.Phone)
		}
	}
	if access.IsHashedPassword {
		args["ua_password"] = access.Password
	} else {
		args["ua_password"] = utils.Hash(access.Password)
	}
	return args
}
