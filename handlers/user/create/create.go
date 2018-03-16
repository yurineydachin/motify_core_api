package user_create

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
	return "Create user"
}

func (*Handler) Description() string {
	return "create user and return user with token if success"
}
