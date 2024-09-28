package email

import (
	"errors"
	"fmt"
	"net/smtp"
)

var (
	ErrSendingEmail = errors.New("couldn't send email")
)

type EmailSMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type EmailService struct {
	config EmailSMTPConfig
}

func NewEmailService(config EmailSMTPConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

func (e *EmailService) Send(to []string, subject, body string) error {
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body)

	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	return smtp.SendMail(addr, auth, e.config.From, to, []byte(msg))
}

func (s *EmailService) SendVerificationMail(email, code string) error {
	msg := fmt.Sprintf("Your verification code is: %s", code)
	return s.Send([]string{email}, "Verify your Account", msg)
}

func (s *EmailService) SendVerificationCompleteMail(email string) error {
	msg := fmt.Sprintf("Your email has been successfully verified")
	return s.Send([]string{email}, "Email verification successful", msg)
}

func (s *EmailService) SendPasswordResetMail(email, token string) error {
	msg := fmt.Sprintf("Here's your password reseet token: %s", token)
	return s.Send([]string{email}, "Reset your password", msg)
}
