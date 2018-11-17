package models

import (
	"fmt"

	"motify_core_api/utils"
)

const (
	UserAccessEmail  = "email"
	UserAccessPhone  = "phone"
	UserAccessFb     = "fb"
	UserAccessGoogle = "google"
)

type UserAccess struct {
	ID               uint64  `db:"id_user_access"`
	UserFK           uint64  `db:"ua_fk_user"`
	IntegrationFK    *uint64 `db:"-"`
	Type             string  `db:"ua_type"`
	IsHashedLogin    bool    `db:"-"`
	Login            string  `db:"ua_login"`
	IsHashedPassword bool    `db:"-"`
	Password         string  `db:"ua_password"`
	UpdatedAt        string  `db:"ua_updated_at"`
	CreatedAt        string  `db:"ua_created_at"`
}

func LoginSufix(integrationID *uint64) string {
	if integrationID != nil && *integrationID > 0 {
		return fmt.Sprintf("_int_%d", *integrationID)
	}
	return ""
}

func (access *UserAccess) SetLogin(value string) {
	value += LoginSufix(access.IntegrationFK)
	access.Login = value
	access.IsHashedLogin = false
}

func (access *UserAccess) SetPassword(value string) {
	access.Password = value
	access.IsHashedPassword = false
}

func (access *UserAccess) MarkAllHashed() {
	access.IsHashedLogin = true
	access.IsHashedPassword = true
}

func (access UserAccess) ToArgs() map[string]interface{} {
	args := map[string]interface{}{
		"id_user_access": access.ID,
		"ua_fk_user":     access.UserFK,
		"ua_type":        access.Type,
	}
	if access.IsHashedLogin {
		args["ua_login"] = access.Login
	} else {
		args["ua_login"] = utils.HashLogin(access.Login)
	}
	if access.IsHashedPassword {
		args["ua_password"] = access.Password
	} else {
		args["ua_password"] = utils.HashPass(access.Password)
	}
	return args
}
