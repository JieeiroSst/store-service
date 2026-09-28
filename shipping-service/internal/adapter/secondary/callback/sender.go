package callback

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"go.uber.org/fx"
)

const (
	HeaderEventID   = "X-Shipping-Event-Id"
	HeaderSignature = "X-Shipping-Signature"
)

type Sender struct {
	secret  []byte
	allowed []string
	http    *http.Client
}

func New(cfg *config.Config) *Sender {
	return &Sender{
		secret:  []byte(cfg.Callback.Secret),
		allowed: cfg.Callback.AllowedHostList(),
		http: &http.Client{
			Timeout: cfg.Callback.TimeoutDuration(),
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (s *Sender) Allowed(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return errors.New("must be an absolute http(s) URL")
	}
	if len(s.allowed) == 0 {
		return errors.New("callbacks are disabled (CALLBACK_ALLOWED_HOSTS is empty)")
	}
	host := strings.ToLower(u.Hostname())
	for _, a := range s.allowed {
		if host == a || (strings.HasPrefix(a, ".") && strings.HasSuffix(host, a)) {
			return nil
		}
	}
	return fmt.Errorf("host %s is not in CALLBACK_ALLOWED_HOSTS", host)
}

func Sign(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (s *Sender) Send(ctx context.Context, target string, event model.CallbackEvent) error {
	if err := s.Allowed(target); err != nil {
		return err
	}
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderEventID, event.EventID)
	req.Header.Set(HeaderSignature, Sign(s.secret, body))

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("callback %s: %w", target, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("callback %s: status %d: %s", target, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.CallbackSender)))),
)
