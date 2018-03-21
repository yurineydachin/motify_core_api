package user_update

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
	return "Update user"
}

func (*Handler) Description() string {
	return "Update user profile"
}
