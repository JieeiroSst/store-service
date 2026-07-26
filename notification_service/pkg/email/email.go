package email

import (
	"crypto/tls"

	gomail "gopkg.in/mail.v2"
)

type Client struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewClient(host string, port int, username, password, from string) *Client {
	return &Client{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (c *Client) Send(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", c.from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(c.host, c.port, c.username, c.password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}
