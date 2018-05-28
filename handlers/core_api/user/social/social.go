package user_social

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
	return "Login or register user via social net"
}

func (*Handler) Description() string {
	return "Login or register user via social net: FB or Google"
}
