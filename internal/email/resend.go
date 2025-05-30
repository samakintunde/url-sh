package email

import (
	"fmt"

	"github.com/resend/resend-go/v2"
)

type ResendService struct {
	config EmailSMTPConfig
	client *resend.Client
}

func NewResendService(config EmailSMTPConfig, apiKey string) *ResendService {
	return &ResendService{
		config: config,
		client: resend.NewClient(apiKey),
	}
}

func (e *ResendService) Send(to []string, subject, body string) error {
	params := &resend.SendEmailRequest{
		From:    e.config.From,
		To:      to,
		Subject: subject,
		Html:    body,
	}

	_, err := e.client.Emails.Send(params)

	if err != nil {
		return err
	}

	return nil
}

func (s *ResendService) SendVerificationMail(email, code string) error {
	msg := fmt.Sprintf("Your verification code is: %s", code)
	return s.Send([]string{email}, "Verify your Account", msg)
}

func (s *ResendService) SendVerificationCompleteMail(email string) error {
	msg := fmt.Sprintf("Your email has been successfully verified")
	return s.Send([]string{email}, "Email verification successful", msg)
}

func (s *ResendService) SendPasswordResetMail(email, token string) error {
	msg := fmt.Sprintf("Here's your password reseet token: %s", token)
	return s.Send([]string{email}, "Reset your password", msg)
}
