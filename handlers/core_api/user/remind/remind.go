package user_remind

import (
	"motify_core_api/srv/user"
)

type Handler struct {
	userService *user_service.UserService
	emailService *email_service.Service
	emailFrom    string
}

func New(userService *user_service.UserService, emailService *email_service.Service, emailFrom string) *Handler {
	return &Handler{
		userService: userService,
		emailService: emailService,
		emailFrom:    emailFrom,
	}
}

func (*Handler) Caption() string {
	return "Remind user password"
}

func (*Handler) Description() string {
	return "Remind user password"
}
