package user_login

import (
	"motify_core_api/srv/user"
)

type Handler struct {
	userService *user_service.UserService
}

func New(userService *user_service.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

func (*Handler) Caption() string {
	return "Login user"
}

func (*Handler) Description() string {
	return "Try login user and return user with token if success"
}
