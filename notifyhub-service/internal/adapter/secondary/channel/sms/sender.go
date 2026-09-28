package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"go.uber.org/zap"
)

const (
	twilioBaseURL = "https://api.twilio.com/2010-04-01"
	vonageURL     = "https://rest.nexmo.com/sms/json"
	maxSMSLength  = 1600
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
}

type Sender struct {
	cfg    config.SMSConfig
	client *http.Client
	logger *zap.Logger
}

func New(cfg config.SMSConfig, log *zap.Logger) *Sender {
	if log == nil {
		log = zap.NewNop()
	}
	return &Sender{
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
}

func (s *Sender) Type() model.ChannelType { return model.ChannelSMS }

func (s *Sender) Send(ctx context.Context, msg port.Message) error {
	result, err := s.send(ctx, msg)
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

func (s *Sender) send(ctx context.Context, msg port.Message) (*SendResult, error) {
	switch strings.ToLower(s.cfg.Provider) {
	case "twilio":
		return s.sendTwilio(ctx, msg)
	case "vonage", "nexmo":
		return s.sendVonage(ctx, msg)
	default:
		if s.cfg.TwilioSID != "" {
			return s.sendTwilio(ctx, msg)
		}
		if s.cfg.VonageKey != "" {
			return s.sendVonage(ctx, msg)
		}
		return nil, errors.New("no SMS provider configured (set SMS_PROVIDER to twilio or vonage)")
	}
}

type twilioResponse struct {
	SID          string  `json:"sid"`
	Status       string  `json:"status"`
	ErrorCode    *int    `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	Message      string  `json:"message"`
	Code         int     `json:"code"`
}

func (s *Sender) sendTwilio(ctx context.Context, msg port.Message) (*SendResult, error) {
	if s.cfg.TwilioSID == "" || s.cfg.TwilioToken == "" {
		return nil, errors.New("twilio credentials not configured")
	}

	form := url.Values{}
	form.Set("To", msg.Recipient)
	form.Set("From", s.cfg.TwilioFrom)
	form.Set("Body", truncateSMS(msg.Body, maxSMSLength))
	if cb := msg.Data["status_callback"]; cb != "" {
		form.Set("StatusCallback", cb)
	}
	if msid := msg.Data["messaging_service_sid"]; msid != "" {
		form.Del("From")
		form.Set("MessagingServiceSid", msid)
	}

	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages.json", twilioBaseURL, url.PathEscape(s.cfg.TwilioSID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("twilio build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(s.cfg.TwilioSID, s.cfg.TwilioToken)

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
	decodeErr := json.Unmarshal(raw, &tr)

	if resp.StatusCode >= 400 {
		errMsg := fmt.Sprintf("http %d", resp.StatusCode)
		switch {
		case tr.Message != "":
			errMsg += fmt.Sprintf(": %s (code %d)", tr.Message, tr.Code)
		case tr.ErrorMessage != nil:
			errMsg += ": " + *tr.ErrorMessage
		}
		return nil, fmt.Errorf("twilio API error: %s", errMsg)
	}
	if decodeErr != nil {
		return nil, fmt.Errorf("twilio decode response: %w", decodeErr)
	}

	if tr.Status == "failed" || tr.Status == "undelivered" {
		errMsg := "message status=" + tr.Status
		if tr.ErrorMessage != nil {
			errMsg += ": " + *tr.ErrorMessage
		}
		return nil, errors.New(errMsg)
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
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	From      string `json:"from"`
	To        string `json:"to"`
	Text      string `json:"text"`
	Type      string `json:"type"` // text | unicode
}

type vonageResponse struct {
	Messages []vonageMessage `json:"messages"`
}

type vonageMessage struct {
	To        string `json:"to"`
	MessageID string `json:"message-id"`
	Status    string `json:"status"` // "0" = success
	ErrorText string `json:"error-text"`
}

func (s *Sender) sendVonage(ctx context.Context, msg port.Message) (*SendResult, error) {
	if s.cfg.VonageKey == "" || s.cfg.VonageSecret == "" {
		return nil, errors.New("vonage credentials not configured")
	}

	msgType := "text"
	for _, r := range msg.Body {
		if r > 127 {
			msgType = "unicode"
			break
		}
	}

	body, err := json.Marshal(vonageRequest{
		APIKey:    s.cfg.VonageKey,
		APISecret: s.cfg.VonageSecret,
		From:      s.cfg.VonageFrom,
		To:        normalizeE164(msg.Recipient),
		Text:      truncateSMS(msg.Body, maxSMSLength),
		Type:      msgType,
	})
	if err != nil {
		return nil, fmt.Errorf("vonage marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, vonageURL, bytes.NewReader(body))
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
		return nil, errors.New("vonage: empty messages array in response")
	}

	m := vr.Messages[0]
	if m.Status != "0" {
		return nil, fmt.Errorf("vonage send failed (status=%s): %s", m.Status, m.ErrorText)
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
	phone = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(phone)
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}
	return phone
}
