package push_service

import (
	"context"

	"motify_core_api/models"
)

const (
	DeviceIOS     = "ios"
	DeviceAndroid = "android"
)

type Service struct {
	apns *apnservice
}

func New() *Service {
	return &Service{}
}

func (s *Service) AddAPNS(gateway, certificateBase64, keyBase64 string) *Service {
	s.apns = newAPNS(gateway, certificateBase64, keyBase64)
	return s
}

func (s *Service) AddGMS(gateway, certificateBase64, keyBase64 string) *Service {
	return s
}

func (s *Service) Send(ctx context.Context, message, device, pushToken string) error {
	if device == DeviceIOS && s.apns != nil {
		return s.apns.send(ctx, message, pushToken)
	}
	return nil
}

func (s *Service) SendMessages(ctx context.Context, message string, devices []*models.Device) error {
	for _, device := range devices {
		s.Send(ctx, message, device.Device, device.Token)
	}
	return nil
}
