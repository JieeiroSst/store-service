package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/channel"
	"github.com/JIeeiroSst/notifyhub-service/internal/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"go.uber.org/zap"
)

type DeliveryStatus string

const (
	DeliveryQueued    DeliveryStatus = "queued"
	DeliveryDelivered DeliveryStatus = "delivered"
	DeliveryFailed    DeliveryStatus = "failed"
	DeliveryUnknown   DeliveryStatus = "unknown"
)

type SendResult struct {
	MessageSID string
	Status     DeliveryStatus
	Provider   string
	To         string
	Error      string
}

type Sender struct {
	cfg    config.SMSConfig
	client *http.Client
	mu     sync.Mutex
	logger *zap.Logger

	rateTicker *time.Ticker
	rateTokens chan struct{}
}

func New(cfg config.SMSConfig, log *zap.Logger) *Sender {
	if log == nil {
		log = zap.NewNop()
	}

	s := &Sender{
		cfg:    cfg,
		logger: log,
		client: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
	return s
}

func (s *Sender) Type() model.ChannelType { return model.ChannelSMS }

func (s *Sender) Send(ctx context.Context, msg channel.Message) error {
	result, err := s.sendWithResult(ctx, msg)
	if err != nil {
		return err
	}
	s.logger.Info("sms sent",
		zap.String("to", msg.Recipient),
		zap.String("sid", result.MessageSID),
		zap.String("status", string(result.Status)),
		zap.String("provider", result.Provider),
	)
	return nil
}

func (s *Sender) sendWithResult(ctx context.Context, msg channel.Message) (*SendResult, error) {
	provider := strings.ToLower(s.cfg.Provider)
	switch provider {
	case "twilio":
		return s.sendTwilio(ctx, msg)
	case "vonage", "nexmo":
		return s.sendVonage(ctx, msg)
	default:
		// Default to Twilio if SID is set, otherwise Vonage
		if s.cfg.TwilioSID != "" {
			return s.sendTwilio(ctx, msg)
		}
		if s.cfg.VonageKey != "" {
			return s.sendVonage(ctx, msg)
		}
		return nil, fmt.Errorf("no SMS provider configured (set SMS_PROVIDER to twilio or vonage)")
	}
}

type twilioResponse struct {
	SID          string  `json:"sid"`
	Status       string  `json:"status"`
	ErrorCode    *int    `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	To           string  `json:"to"`
	From         string  `json:"from"`
	Body         string  `json:"body"`
	Price        *string `json:"price"`
	PriceUnit    string  `json:"price_unit"`
	DateCreated  string  `json:"date_created"`
}

func (s *Sender) sendTwilio(ctx context.Context, msg channel.Message) (*SendResult, error) {
	if s.cfg.TwilioSID == "" || s.cfg.TwilioToken == "" {
		return nil, fmt.Errorf("twilio credentials not configured")
	}

	endpoint := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		s.cfg.TwilioSID,
	)

	// Twilio uses form-encoded body
	form := url.Values{}
	form.Set("To", msg.Recipient)
	form.Set("From", s.cfg.TwilioFrom)
	form.Set("Body", truncateSMS(msg.Body, 1600))

	if msg.Data != nil {
		if cb, ok := msg.Data["status_callback"]; ok && cb != "" {
			form.Set("StatusCallback", cb)
		}
		if msid, ok := msg.Data["messaging_service_sid"]; ok && msid != "" {
			form.Del("From")
			form.Set("MessagingServiceSid", msid)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("twilio build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.cfg.TwilioSID, s.cfg.TwilioToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twilio http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil {
		return nil, fmt.Errorf("twilio read response: %w", err)
	}

	var tr twilioResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, fmt.Errorf("twilio decode response: %w — body: %s", err, string(raw))
	}

	// HTTP 4xx/5xx
	if resp.StatusCode >= 400 {
		errMsg := fmt.Sprintf("http %d", resp.StatusCode)
		if tr.ErrorMessage != nil {
			errMsg += ": " + *tr.ErrorMessage
		}
		if tr.ErrorCode != nil {
			errMsg += fmt.Sprintf(" (code %d)", *tr.ErrorCode)
		}
		return nil, fmt.Errorf("twilio API error: %s", errMsg)
	}

	// Explicit failure status
	switch tr.Status {
	case "failed", "undelivered":
		errMsg := fmt.Sprintf("message status=%s", tr.Status)
		if tr.ErrorMessage != nil {
			errMsg += ": " + *tr.ErrorMessage
		}
		return &SendResult{
			MessageSID: tr.SID,
			Status:     DeliveryFailed,
			Provider:   "twilio",
			To:         msg.Recipient,
			Error:      errMsg,
		}, fmt.Errorf(errMsg)
	}

	return &SendResult{
		MessageSID: tr.SID,
		Status:     mapTwilioStatus(tr.Status),
		Provider:   "twilio",
		To:         msg.Recipient,
	}, nil
}

func mapTwilioStatus(s string) DeliveryStatus {
	switch s {
	case "delivered":
		return DeliveryDelivered
	case "queued", "sending", "sent", "accepted", "scheduled":
		return DeliveryQueued
	case "failed", "undelivered":
		return DeliveryFailed
	default:
		return DeliveryUnknown
	}
}

type vonageRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
	Type string `json:"type"` // text | unicode
}

type vonageResponse struct {
	Messages []vonageMessage `json:"messages"`
}

type vonageMessage struct {
	To               string `json:"to"`
	MessageID        string `json:"message-id"`
	Status           string `json:"status"` // "0" = success
	ErrorText        string `json:"error-text"`
	RemainingBalance string `json:"remaining-balance"`
	MessagePrice     string `json:"message-price"`
	Network          string `json:"network"`
}

func (s *Sender) sendVonage(ctx context.Context, msg channel.Message) (*SendResult, error) {
	if s.cfg.VonageKey == "" || s.cfg.VonageSecret == "" {
		return nil, fmt.Errorf("vonage credentials not configured")
	}

	msgType := "text"
	// Use unicode type for non-ASCII content
	for _, r := range msg.Body {
		if r > 127 {
			msgType = "unicode"
			break
		}
	}

	payload := vonageRequest{
		From: s.cfg.TwilioFrom, 
		To:   normalizeE164(msg.Recipient),
		Text: truncateSMS(msg.Body, 1600),
		Type: msgType,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("vonage marshal: %w", err)
	}

	endpoint := "https://rest.nexmo.com/sms/json"
	q := url.Values{}
	q.Set("api_key", s.cfg.VonageKey)
	q.Set("api_secret", s.cfg.VonageSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		endpoint+"?"+q.Encode(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("vonage build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vonage http: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("vonage http %d: %s", resp.StatusCode, string(raw))
	}

	var vr vonageResponse
	if err := json.Unmarshal(raw, &vr); err != nil {
		return nil, fmt.Errorf("vonage decode: %w", err)
	}

	if len(vr.Messages) == 0 {
		return nil, fmt.Errorf("vonage: empty messages array in response")
	}

	m := vr.Messages[0]
	if m.Status != "0" {
		return &SendResult{
			MessageSID: m.MessageID,
			Status:     DeliveryFailed,
			Provider:   "vonage",
			To:         msg.Recipient,
			Error:      m.ErrorText,
		}, fmt.Errorf("vonage send failed (status=%s): %s", m.Status, m.ErrorText)
	}

	return &SendResult{
		MessageSID: m.MessageID,
		Status:     DeliveryQueued,
		Provider:   "vonage",
		To:         msg.Recipient,
	}, nil
}

func truncateSMS(body string, maxLen int) string {
	runes := []rune(body)
	if len(runes) <= maxLen {
		return body
	}
	return string(runes[:maxLen-3]) + "..."
}

func normalizeE164(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}
	return phone
}