package user_device

import (
	"motify_core_api/srv/device"
)

type Handler struct {
	deviceService *device_service.Service
}

func New(deviceService *device_service.Service) *Handler {
	return &Handler{
		deviceService: deviceService,
	}
}

func (*Handler) Caption() string {
	return "Add device to user"
}

func (*Handler) Description() string {
	return "Add device to user"
}
