package notifier

import (
	"context"
	"fmt"
	"net/smtp"
	"os"

	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
)

type SMTPNotifier struct {
	host, port, from, username, password string
}

func NewSMTPNotifier() notificationapp.Notifier {
	return &SMTPNotifier{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		from:     os.Getenv("SMTP_FROM"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (n *SMTPNotifier) Channel() notificationapp.Channel { return notificationapp.ChannelEmail }

func (n *SMTPNotifier) Send(ctx context.Context, recipient, templateCode string, payload map[string]any) error {
	if n.host == "" {
		return fmt.Errorf("SMTP_HOST not configured")
	}
	addr := n.host + ":" + n.port
	var auth smtp.Auth
	if n.username != "" {
		auth = smtp.PlainAuth("", n.username, n.password, n.host)
	}

	subject := "Notification: " + templateCode
	body := renderTemplate(templateCode, payload)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", n.from, recipient, subject, body)

	return smtp.SendMail(addr, auth, n.from, []string{recipient}, []byte(msg))
}

func renderTemplate(templateCode string, payload map[string]any) string {
	return fmt.Sprintf("Template: %s\nData: %v", templateCode, payload)
}
