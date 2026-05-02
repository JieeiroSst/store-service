package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/http"
	"net/smtp"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/channel"
	"github.com/JIeeiroSst/notifyhub-service/internal/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"go.uber.org/zap"
)

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type emailMessage struct {
	channel.Message
	CC          []string
	BCC         []string
	ReplyTo     string
	Attachments []Attachment
}

type smtpPool struct {
	mu      sync.Mutex
	conns   []*smtp.Client
	maxSize int
	addr    string
	auth    smtp.Auth
	tlsCfg  *tls.Config
}

func newSMTPPool(addr string, auth smtp.Auth, tlsCfg *tls.Config, size int) *smtpPool {
	return &smtpPool{
		addr:    addr,
		auth:    auth,
		tlsCfg:  tlsCfg,
		maxSize: size,
	}
}

func (p *smtpPool) acquire() (*smtp.Client, error) {
	p.mu.Lock()
	if len(p.conns) > 0 {
		c := p.conns[len(p.conns)-1]
		p.conns = p.conns[:len(p.conns)-1]
		p.mu.Unlock()
		if err := c.Noop(); err == nil {
			return c, nil
		}
		c.Close()
	} else {
		p.mu.Unlock()
	}
	return p.dial()
}

func (p *smtpPool) release(c *smtp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.conns) < p.maxSize {
		p.conns = append(p.conns, c)
	} else {
		c.Close()
	}
}

func (p *smtpPool) dial() (*smtp.Client, error) {
	c, err := smtp.Dial(p.addr)
	if err != nil {
		return nil, fmt.Errorf("smtp dial %s: %w", p.addr, err)
	}

	if p.tlsCfg != nil {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(p.tlsCfg); err != nil {
				c.Close()
				return nil, fmt.Errorf("starttls: %w", err)
			}
		}
	}

	if p.auth != nil {
		if err := c.Auth(p.auth); err != nil {
			c.Close()
			return nil, fmt.Errorf("smtp auth: %w", err)
		}
	}
	return c, nil
}

func (p *smtpPool) closeAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.conns {
		c.Close()
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

	if strings.ToLower(cfg.Provider) != "sendgrid" {
		host, _ := splitHostPort(cfg.SMTPHost, cfg.SMTPPort)
		addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
		auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, host)
		tlsCfg := &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		}
		s.pool = newSMTPPool(addr, auth, tlsCfg, 5)
	}

	return s
}

func (s *Sender) Type() model.ChannelType { return model.ChannelEmail }

func (s *Sender) Send(ctx context.Context, msg channel.Message) error {
	em := emailMessage{Message: msg}
	switch strings.ToLower(s.cfg.Provider) {
	case "sendgrid":
		return s.sendViaSendGrid(ctx, em)
	default:
		return s.sendViaSMTP(ctx, em)
	}
}

func (s *Sender) SendWithOptions(ctx context.Context, msg emailMessage) error {
	switch strings.ToLower(s.cfg.Provider) {
	case "sendgrid":
		return s.sendViaSendGrid(ctx, msg)
	default:
		return s.sendViaSMTP(ctx, msg)
	}
}

func (s *Sender) Close() {
	if s.pool != nil {
		s.pool.closeAll()
	}
}

func (s *Sender) sendViaSMTP(ctx context.Context, msg emailMessage) error {
	client, err := s.pool.acquire()
	if err != nil {
		s.logger.Warn("smtp pool acquire failed, falling back to direct send", zap.Error(err))
		return s.sendViaSMTPDirect(msg)
	}
	defer s.pool.release(client)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return s.smtpSend(client, msg)
}

func (s *Sender) sendViaSMTPDirect(msg emailMessage) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	host, _ := splitHostPort(s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, host)

	if len(msg.Attachments) == 0 && len(msg.CC) == 0 && len(msg.BCC) == 0 {
		body := buildSimpleBody(msg)
		allTo := append([]string{msg.Recipient}, msg.CC...)
		allTo = append(allTo, msg.BCC...)
		return smtp.SendMail(addr, auth, s.cfg.FromAddr, allTo, body)
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	return s.smtpSend(client, msg)
}

func (s *Sender) smtpSend(client *smtp.Client, msg emailMessage) error {
	if err := client.Mail(s.cfg.FromAddr); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}

	allRecipients := []string{msg.Recipient}
	allRecipients = append(allRecipients, msg.CC...)
	allRecipients = append(allRecipients, msg.BCC...)

	for _, rcpt := range allRecipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("RCPT TO <%s>: %w", rcpt, err)
		}
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	defer wc.Close()

	rawMsg := buildMIMEMessage(s.cfg.FromAddr, msg)
	_, err = wc.Write(rawMsg)
	return err
}

func buildMIMEMessage(from string, msg emailMessage) []byte {
	var buf bytes.Buffer
	isHTML := isHTMLBody(msg.Body)

	if len(msg.Attachments) == 0 && len(msg.CC) == 0 && len(msg.BCC) == 0 && msg.ReplyTo == "" {
		buf.WriteString(buildSimpleBodyStrWithFrom(from, msg, isHTML))
		return buf.Bytes()
	}

	mw := multipart.NewWriter(&buf)

	buf2 := &bytes.Buffer{}
	buf2.WriteString("From: " + from + "\r\n")
	buf2.WriteString("To: " + msg.Recipient + "\r\n")
	if len(msg.CC) > 0 {
		buf2.WriteString("Cc: " + strings.Join(msg.CC, ", ") + "\r\n")
	}
	if msg.ReplyTo != "" {
		buf2.WriteString("Reply-To: " + msg.ReplyTo + "\r\n")
	}
	buf2.WriteString("Subject: " + msg.Subject + "\r\n")
	buf2.WriteString("MIME-Version: 1.0\r\n")
	buf2.WriteString("Content-Type: multipart/mixed; boundary=" + mw.Boundary() + "\r\n")
	buf2.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	buf2.WriteString("\r\n")

	var bodyHeaders textproto.MIMEHeader
	if isHTML {
		bodyHeaders = textproto.MIMEHeader{"Content-Type": {"text/html; charset=utf-8"}, "Content-Transfer-Encoding": {"quoted-printable"}}
	} else {
		bodyHeaders = textproto.MIMEHeader{"Content-Type": {"text/plain; charset=utf-8"}, "Content-Transfer-Encoding": {"quoted-printable"}}
	}
	bodyPart, _ := mw.CreatePart(bodyHeaders)
	qpw := quotedprintable.NewWriter(bodyPart)
	qpw.Write([]byte(msg.Body))
	qpw.Close()

	for _, att := range msg.Attachments {
		attHeaders := textproto.MIMEHeader{
			"Content-Type":              {att.ContentType + "; name=" + att.Filename},
			"Content-Disposition":       {"attachment; filename=" + att.Filename},
			"Content-Transfer-Encoding": {"base64"},
		}
		attPart, _ := mw.CreatePart(attHeaders)
		attPart.Write(att.Data)
	}
	mw.Close()

	return append(buf2.Bytes(), buf.Bytes()...)
}

func buildSimpleBody(msg emailMessage) []byte {
	return buildSimpleBodyWithFrom("", msg, isHTMLBody(msg.Body))
}

func buildSimpleBodyWithFrom(from string, msg emailMessage, isHTML bool) []byte {
	var sb strings.Builder
	if from != "" {
		sb.WriteString("From: " + from + "\r\n")
	}
	sb.WriteString("To: " + msg.Recipient + "\r\n")
	sb.WriteString("Subject: " + msg.Subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	if isHTML {
		sb.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	} else {
		sb.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	}
	sb.WriteString("\r\n")
	sb.WriteString(msg.Body)
	return []byte(sb.String())
}

func buildSimpleBodyStrWithFrom(from string, msg emailMessage, isHTML bool) string {
	var sb strings.Builder
	if from != "" {
		sb.WriteString("From: " + from + "\r\n")
	}
	sb.WriteString("To: " + msg.Recipient + "\r\n")
	sb.WriteString("Subject: " + msg.Subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	if isHTML {
		sb.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	} else {
		sb.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	}
	sb.WriteString("\r\n")
	sb.WriteString(msg.Body)
	return sb.String()
}

func isHTMLBody(body string) bool {
	lower := strings.ToLower(body)
	return strings.Contains(lower, "<html") ||
		strings.Contains(lower, "<body") ||
		strings.Contains(lower, "<p>") ||
		strings.Contains(lower, "<div") ||
		strings.Contains(lower, "<h1")
}

func splitHostPort(host string, port int) (string, string) {
	if strings.Contains(host, ":") {
		h, p, err := net.SplitHostPort(host)
		if err == nil {
			return h, p
		}
	}
	return host, fmt.Sprintf("%d", port)
}

type sendGridPayload struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
	Attachments      []sendGridAttachment      `json:"attachments,omitempty"`
	ReplyToList      []sendGridEmail           `json:"reply_to_list,omitempty"`
}

type sendGridPersonalization struct {
	To  []sendGridEmail `json:"to"`
	CC  []sendGridEmail `json:"cc,omitempty"`
	BCC []sendGridEmail `json:"bcc,omitempty"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sendGridAttachment struct {
	Content     string `json:"content"`
	Type        string `json:"type"`
	Filename    string `json:"filename"`
	Disposition string `json:"disposition"`
}

func (s *Sender) sendViaSendGrid(ctx context.Context, msg emailMessage) error {
	if s.cfg.SendGridKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY is not configured")
	}

	// Determine content type
	contentType := "text/plain"
	if isHTMLBody(msg.Body) {
		contentType = "text/html"
	}

	p := &sendGridPersonalization{
		To: []sendGridEmail{{Email: msg.Recipient}},
	}
	for _, cc := range msg.CC {
		p.CC = append(p.CC, sendGridEmail{Email: cc})
	}
	for _, bcc := range msg.BCC {
		p.BCC = append(p.BCC, sendGridEmail{Email: bcc})
	}

	payload := sendGridPayload{
		Personalizations: []sendGridPersonalization{*p},
		From:             sendGridEmail{Email: s.cfg.FromAddr},
		Subject:          msg.Subject,
		Content: []sendGridContent{
			{Type: contentType, Value: msg.Body},
		},
	}

	if msg.ReplyTo != "" {
		payload.ReplyToList = []sendGridEmail{{Email: msg.ReplyTo}}
	}

	// Add attachments
	for _, att := range msg.Attachments {
		payload.Attachments = append(payload.Attachments, sendGridAttachment{
			Content:     encodeBase64(att.Data),
			Type:        att.ContentType,
			Filename:    att.Filename,
			Disposition: "attachment",
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.sendgrid.com/v3/mail/send",
		bytes.NewReader(body),
	)
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

	s.logger.Info("email sent via sendgrid",
		zap.String("to", msg.Recipient),
		zap.String("subject", msg.Subject),
	)
	return nil
}

func encodeBase64(data []byte) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var buf strings.Builder
	for i := 0; i < len(data); i += 3 {
		b0 := data[i]
		var b1, b2 byte
		if i+1 < len(data) {
			b1 = data[i+1]
		}
		if i+2 < len(data) {
			b2 = data[i+2]
		}
		buf.WriteByte(chars[b0>>2])
		buf.WriteByte(chars[((b0&3)<<4)|(b1>>4)])
		if i+1 < len(data) {
			buf.WriteByte(chars[((b1&0xf)<<2)|(b2>>6)])
		} else {
			buf.WriteByte('=')
		}
		if i+2 < len(data) {
			buf.WriteByte(chars[b2&0x3f])
		} else {
			buf.WriteByte('=')
		}
	}
	return buf.String()
}
