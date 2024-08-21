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

type emailService struct {
	config EmailSMTPConfig
}

func NewEmailService(config EmailSMTPConfig) Emailer {
	return emailService{
		config: config,
	}
}

func (e emailService) Send(to []string, subject, body string) error {
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body)

	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	return smtp.SendMail(addr, auth, e.config.From, to, []byte(msg))
}
