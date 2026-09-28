package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"go.uber.org/zap"
)

const sendGridURL = "https://api.sendgrid.com/v3/mail/send"

type smtpPool struct {
	mu      sync.Mutex
	conns   []*smtp.Client
	maxSize int
	addr    string
	auth    smtp.Auth
	tlsCfg  *tls.Config
}

func (p *smtpPool) acquire() (*smtp.Client, error) {
	for {
		p.mu.Lock()
		if len(p.conns) == 0 {
			p.mu.Unlock()
			return p.dial()
		}
		c := p.conns[len(p.conns)-1]
		p.conns = p.conns[:len(p.conns)-1]
		p.mu.Unlock()
		if err := c.Reset(); err == nil {
			return c, nil
		}
		_ = c.Close()
	}
}

func (p *smtpPool) release(c *smtp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.conns) < p.maxSize {
		p.conns = append(p.conns, c)
		return
	}
	_ = c.Quit()
}

func (p *smtpPool) dial() (*smtp.Client, error) {
	c, err := smtp.Dial(p.addr)
	if err != nil {
		return nil, fmt.Errorf("smtp dial %s: %w", p.addr, err)
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(p.tlsCfg); err != nil {
			_ = c.Close()
			return nil, fmt.Errorf("starttls: %w", err)
		}
	}
	if p.auth != nil {
		if err := c.Auth(p.auth); err != nil {
			_ = c.Close()
			return nil, fmt.Errorf("smtp auth: %w", err)
		}
	}
	return c, nil
}

func (p *smtpPool) closeAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.conns {
		_ = c.Quit()
	}
	p.conns = nil
}

type Sender struct {
	cfg        config.EmailConfig
	pool       *smtpPool
	httpClient *http.Client
	logger     *zap.Logger
}

func New(cfg config.EmailConfig, log *zap.Logger) *Sender {
	if log == nil {
		log = zap.NewNop()
	}
	s := &Sender{
		cfg:    cfg,
		logger: log,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 90 * time.Second,
			},
		},
	}

	if !s.useSendGrid() {
		var auth smtp.Auth
		if cfg.SMTPUser != "" {
			auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
		}
		s.pool = &smtpPool{
			addr:    fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
			auth:    auth,
			tlsCfg:  &tls.Config{ServerName: cfg.SMTPHost, MinVersion: tls.VersionTLS12},
			maxSize: 5,
		}
	}
	return s
}

func (s *Sender) useSendGrid() bool { return strings.EqualFold(s.cfg.Provider, "sendgrid") }

func (s *Sender) Type() model.ChannelType { return model.ChannelEmail }

func (s *Sender) Send(ctx context.Context, msg port.Message) error {
	addr, err := mail.ParseAddress(msg.Recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient %q: %w", msg.Recipient, err)
	}
	msg.Recipient = addr.Address
	if s.useSendGrid() {
		return s.sendViaSendGrid(ctx, msg)
	}
	return s.sendViaSMTP(ctx, msg)
}

func (s *Sender) Close() {
	if s.pool != nil {
		s.pool.closeAll()
	}
}

func (s *Sender) sendViaSMTP(ctx context.Context, msg port.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is not configured")
	}
	client, err := s.pool.acquire()
	if err != nil {
		return err
	}
	if err := s.smtpSend(client, msg); err != nil {
		_ = client.Close()
		return err
	}
	s.pool.release(client)
	return nil
}

func (s *Sender) smtpSend(client *smtp.Client, msg port.Message) error {
	if err := client.Mail(s.cfg.FromAddr); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err := client.Rcpt(msg.Recipient); err != nil {
		return fmt.Errorf("RCPT TO <%s>: %w", msg.Recipient, err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := wc.Write(buildMIMEMessage(s.cfg.FromAddr, msg, time.Now())); err != nil {
		_ = wc.Close()
		return fmt.Errorf("write message: %w", err)
	}
	// Close is where the server accepts or rejects the message.
	if err := wc.Close(); err != nil {
		return fmt.Errorf("end DATA: %w", err)
	}
	return nil
}

func buildMIMEMessage(from string, msg port.Message, now time.Time) []byte {
	contentType := "text/plain; charset=utf-8"
	if isHTMLBody(msg.Body) {
		contentType = "text/html; charset=utf-8"
	}

	var buf bytes.Buffer
	writeHeader := func(k, v string) { buf.WriteString(k + ": " + v + "\r\n") }
	writeHeader("From", headerSafe(from))
	writeHeader("To", headerSafe(msg.Recipient))
	writeHeader("Subject", mime.QEncoding.Encode("utf-8", headerSafe(msg.Subject)))
	writeHeader("Date", now.Format(time.RFC1123Z))
	writeHeader("MIME-Version", "1.0")
	writeHeader("Content-Type", contentType)
	writeHeader("Content-Transfer-Encoding", "quoted-printable")
	buf.WriteString("\r\n")

	qp := quotedprintable.NewWriter(&buf)
	_, _ = qp.Write([]byte(msg.Body))
	_ = qp.Close()
	return buf.Bytes()
}

func headerSafe(v string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(v)
}

func isHTMLBody(body string) bool {
	lower := strings.ToLower(body)
	for _, tag := range []string{"<html", "<body", "<p>", "<div", "<h1", "<br", "<table"} {
		if strings.Contains(lower, tag) {
			return true
		}
	}
	return false
}

type sendGridPayload struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
}

type sendGridPersonalization struct {
	To []sendGridEmail `json:"to"`
}

type sendGridEmail struct {
	Email string `json:"email"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (s *Sender) sendViaSendGrid(ctx context.Context, msg port.Message) error {
	if s.cfg.SendGridKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY is not configured")
	}
	contentType := "text/plain"
	if isHTMLBody(msg.Body) {
		contentType = "text/html"
	}

	body, err := json.Marshal(sendGridPayload{
		Personalizations: []sendGridPersonalization{{To: []sendGridEmail{{Email: msg.Recipient}}}},
		From:             sendGridEmail{Email: s.cfg.FromAddr},
		Subject:          headerSafe(msg.Subject),
		Content:          []sendGridContent{{Type: contentType, Value: msg.Body}},
	})
	if err != nil {
		return fmt.Errorf("marshal sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendGridURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build sendgrid request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.SendGridKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sendgrid http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("sendgrid API error %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}
