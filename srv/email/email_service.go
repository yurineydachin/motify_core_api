package email_service

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-gomail/gomail"
)

type Service struct {
	host         string
	port         uint
	userEmail    string
	userPassword string
}

func NewService(host string, port uint, userEmail, userPassword string) *Service {
	return &Service{
		host:         host,
		port:         port,
		userEmail:    userEmail,
		userPassword: userPassword,
	}
}

func (s *Service) SendEmployeeInvite(ctx context.Context, to string, from string, qrCode []byte) error {
	imgBase64Str := base64.StdEncoding.EncodeToString(qrCode)

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Hello, you've got new motification")
	m.SetBody("text/html", `
    <html><body>
    <h1>Hola!</h1>
    <p>Thank you for sugning up for Motify. We're really happy to have you!</p>
    <p>Click the link below to get your agent account set up.</p>
    <img src="data:image/png;base64,`+imgBase64Str+`" /> 
    </body></html>
    `)

	if s.userEmail == "" || s.userPassword == "" {
		d := gomail.Dialer{Host: s.host, Port: int(s.port)}
		if err := d.DialAndSend(m); err != nil {
			return err
		}
		return nil
	}

	d := gomail.NewDialer(s.host, int(s.port), s.userEmail, s.userPassword)
	send, err := d.Dial()
	if err != nil {
		return fmt.Errorf("Dial: %s", err.Error())
	}
	if err := gomail.Send(send, m); err != nil {
		return fmt.Errorf("Send: %s", err.Error())
	}

	return nil
}
