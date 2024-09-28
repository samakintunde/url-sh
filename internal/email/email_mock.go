package email

import (
	"fmt"
	"log/slog"
)

type mockEmailService struct{}

func NewMockEmailService() Emailer {
	return &mockEmailService{}
}

func (s *mockEmailService) Send(to []string, subject, body string) error {
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body)

	slog.Info("Email sent", "to", to, "msg", []byte(msg))
	return nil
}

func (s *mockEmailService) SendVerificationMail(email string, code string) error {
	return s.Send([]string{email}, "Verify your email", code)
}

func (s *mockEmailService) SendVerificationCompleteMail(email string) error {
	msg := fmt.Sprintf("Your email has been successfully verified")
	return s.Send([]string{email}, "Email verification successful", msg)
}

func (s *mockEmailService) SendPasswordResetMail(email, token string) error {
	msg := fmt.Sprintf("Here's your password reseet token: %s", token)
	return s.Send([]string{email}, "Reset your password", msg)
}
