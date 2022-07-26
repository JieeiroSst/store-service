package email

import (
	"context"

	gomail "gopkg.in/mail.v2"
)

type Email interface {
	SendEmail(ctx context.Context, body Body, emailFrom string, emailTo string) error
}

type Config struct {
	NameEmail string
	Password  string
}

type Body struct {
	Subject string
	Body    string
}

type email struct {
	Config
	gomail *gomail.Dialer
}

func NewEmail(config Config) Email {
	gomail := gomail.NewDialer("smtp.gmail.com", 587, config.NameEmail, config.Password)
	return &email{
		gomail: gomail,
	}
}

func (e *email) SendEmail(ctx context.Context, body Body, emailFrom string, emailTo string) error {
	email := gomail.NewMessage()

	email.SetHeader("From", emailFrom)
	email.SetHeader("To", emailTo)
	email.SetHeader("Subject", body.Subject)
	email.SetBody("text/plain", body.Body)
	if err := e.gomail.DialAndSend(email); err != nil {
		return err
	}
	return nil
}
