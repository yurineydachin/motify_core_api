package user_approve_send

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
	return "Approve user email"
}

func (*Handler) Description() string {
	return "Approve user email"
}
