package email

type Emailer interface {
	Send(to []string, subject, body string) error
	SendVerificationMail(email, code string) error
	SendVerificationCompleteMail(email string) error
	SendPasswordResetMail(email, token string) error
}
