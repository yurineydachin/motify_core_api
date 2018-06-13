package push_service

import (
	"context"

	"github.com/anachronistic/apns"

	"motify_core_api/godep_libs/service/logger"
)

type apnservice struct {
	client *apns.Client
}

func newAPNS(gateway, certificateBase64, keyBase64 string) *apnservice {
	return &apnservice{
		client: apns.BareClient(gateway, certificateBase64, keyBase64),
	}
}

func (s *apnservice) send(ctx context.Context, message, pushToken string) error {
	payload := apns.NewPayload()
	payload.Alert = message
	pn := apns.NewPushNotification()
	pn.DeviceToken = pushToken
	pn.AddPayload(payload)
	resp := s.client.Send(pn)
	alert, err := pn.PayloadString()
	logger.Error(ctx, "Alert:", alert)
	logger.Error(ctx, "Alert, err:", err)
	logger.Error(ctx, "Success:", resp.Success)
	logger.Error(ctx, "Error:", resp.Error)
	return resp.Error
}
