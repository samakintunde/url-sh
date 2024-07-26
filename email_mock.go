package main

import (
	"fmt"
	"log/slog"
)

type mockEmailService struct{}

func NewMockEmailService() mockEmailService {
	return mockEmailService{}
}

func (e mockEmailService) Send(to []string, subject, body string) error {
	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body)

	slog.Info("Email sent", "to", to, "msg", []byte(msg))
	return nil
}
