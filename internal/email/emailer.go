package email

type Emailer interface {
	Send(to []string, subject, body string) error
}
