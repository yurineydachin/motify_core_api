package user_create

import (
	"motify_core_api/srv/email"
	"motify_core_api/srv/user"
)

type Handler struct {
	userService  *user_service.UserService
	emailService *email_service.Service
	emailFrom    string
}

func New(userService *user_service.UserService, emailService *email_service.Service, emailFrom string) *Handler {
	return &Handler{
		userService:  userService,
		emailService: emailService,
		emailFrom:    emailFrom,
	}
}

func (*Handler) Caption() string {
	return "Create user"
}

func (*Handler) Description() string {
	return "create user and return user with token if success"
}
