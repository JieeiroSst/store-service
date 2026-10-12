package userservice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

var ErrUpstream = errors.New("user-service unavailable")

const (
	negativeTTL = time.Minute
	maxCache    = 10_000
)

type existsEntry struct {
	exists  bool
	expires time.Time
}

type Client struct {
	baseURL       string
	internalToken string
	jidField      string
	allowPassword bool
	cacheTTL      time.Duration
	http          *http.Client
	mu            sync.Mutex
	exists        map[string]existsEntry
}

var _ port.ExternalAuthenticator = (*Client)(nil)

func New(cfg config.UserServiceConfig) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		baseURL:       strings.TrimRight(cfg.URL, "/"),
		internalToken: cfg.InternalToken,
		jidField:      cfg.JIDField,
		allowPassword: cfg.AllowPasswordLogin,
		cacheTTL:      cfg.CacheTTL,
		http:          &http.Client{Timeout: timeout},
		exists:        map[string]existsEntry{},
	}
}

func (c *Client) CheckPassword(ctx context.Context, user, luser, lserver, password string) (bool, error) {
	if password == "" {
		return false, nil
	}
	if looksLikeJWT(password) {
		return c.checkToken(ctx, luser, password)
	}
	if !c.allowPassword || c.jidField != "username" {
		return false, nil
	}
	return c.checkLogin(ctx, user, password)
}

func (c *Client) checkToken(ctx context.Context, luser, token string) (bool, error) {
	userID, ok, err := c.validate(ctx, token)
	if err != nil || !ok {
		return false, err
	}
	switch c.jidField {
	case "user_id":
		return userID == luser, nil
	default:
		name, valid := domain.Nodeprep(tokenUsername(token))
		return valid && name != "" && name == luser, nil
	}
}

func (c *Client) validate(ctx context.Context, token string) (string, bool, error) {
	status, body, err := c.post(ctx, "/api/v1/validate", map[string]string{"session_token": token})
	if err != nil {
		return "", false, err
	}
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return "", false, nil
	case status != http.StatusOK:
		return "", false, fmt.Errorf("%w: validate returned %d", ErrUpstream, status)
	}
	var out struct {
		Valid       bool   `json:"valid"`
		UserID      string `json:"user_id"`
		UserIDCamel string `json:"userId"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	id := out.UserID
	if id == "" {
		id = out.UserIDCamel
	}
	return id, out.Valid && id != "", nil
}

func (c *Client) checkLogin(ctx context.Context, user, password string) (bool, error) {
	status, body, err := c.post(ctx, "/api/v1/login", map[string]string{"username": user, "password": password})
	if err != nil {
		return false, err
	}
	switch {
	case status == http.StatusUnauthorized || status == http.StatusNotFound || status == http.StatusBadRequest:
		return false, nil
	case status != http.StatusOK:
		return false, fmt.Errorf("%w: login returned %d", ErrUpstream, status)
	}
	var out struct {
		SessionToken      string `json:"sessionToken"`
		SessionTokenSnake string `json:"session_token"`
	}
	_ = json.Unmarshal(body, &out)
	token := out.SessionToken
	if token == "" {
		token = out.SessionTokenSnake
	}
	if token != "" {
		_, _, _ = c.post(context.WithoutCancel(ctx), "/api/v1/logout", map[string]string{"session_token": token})
	}
	return true, nil
}

func (c *Client) EnsureUser(ctx context.Context, account domain.Account) (bool, error) {
	status, body, err := c.post(ctx, "/user/sign-up", map[string]string{
		"username": account.Username,
		"password": account.Password,
		"email":    account.Email,
		"name":     account.Name,
	})
	if err != nil {
		return false, err
	}
	switch {
	case status == http.StatusOK || status == http.StatusCreated:
		if l, ok := domain.Nodeprep(account.Username); ok {
			c.mu.Lock()
			delete(c.exists, l)
			c.mu.Unlock()
		}
		return true, nil
	case status == http.StatusConflict:
		return false, nil
	case status >= 400 && status < 500:
		var out struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(body, &out)
		return false, fmt.Errorf("%w: sign-up returned %d: %s", domain.ErrAccountRejected, status, out.Message)
	}
	return false, fmt.Errorf("%w: sign-up returned %d", ErrUpstream, status)
}

func (c *Client) UserExists(ctx context.Context, luser, lserver string) (bool, error) {
	if luser == "" {
		return false, nil
	}
	now := time.Now()
	c.mu.Lock()
	e, ok := c.exists[luser]
	c.mu.Unlock()
	if ok && now.Before(e.expires) {
		return e.exists, nil
	}
	exists, err := c.lookup(ctx, luser)
	if err != nil {
		if ok {
			return e.exists, nil
		}
		return false, err
	}
	ttl := c.cacheTTL
	if !exists {
		ttl = negativeTTL
	}
	if ttl > 0 {
		c.mu.Lock()
		if len(c.exists) >= maxCache {
			c.exists = map[string]existsEntry{}
		}
		c.exists[luser] = existsEntry{exists: exists, expires: now.Add(ttl)}
		c.mu.Unlock()
	}
	return exists, nil
}

func (c *Client) lookup(ctx context.Context, luser string) (bool, error) {
	if c.jidField == "user_id" {
		status, _, err := c.get(ctx, "/internal/v1/users/"+url.PathEscape(luser), true)
		if err != nil {
			return false, err
		}
		switch status {
		case http.StatusOK:
			return true, nil
		case http.StatusNotFound, http.StatusBadRequest:
			return false, nil
		}
		return false, fmt.Errorf("%w: internal user lookup returned %d", ErrUpstream, status)
	}
	status, body, err := c.get(ctx, "/user?username="+url.QueryEscape(luser), false)
	if err != nil {
		return false, err
	}
	if status == http.StatusNotFound {
		return false, nil
	}
	if status != http.StatusOK {
		return false, fmt.Errorf("%w: find user returned %d", ErrUpstream, status)
	}
	var out struct {
		Users []struct {
			Username string `json:"username"`
		} `json:"users"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return false, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	for _, u := range out.Users {
		if name, ok := domain.Nodeprep(u.Username); ok && name == luser {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) post(ctx context.Context, path string, payload any) (int, []byte, error) {
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *Client) get(ctx context.Context, path string, internal bool) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, nil, err
	}
	if internal && c.internalToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.internalToken)
	}
	return c.do(req)
}

func (c *Client) do(req *http.Request) (int, []byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	return resp.StatusCode, body, nil
}

func looksLikeJWT(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 || len(s) < 20 {
		return false
	}
	for _, p := range parts[:2] {
		if _, err := base64.RawURLEncoding.DecodeString(p); err != nil {
			return false
		}
	}
	return true
}

func tokenUsername(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Username string `json:"username"`
	}
	_ = json.Unmarshal(raw, &claims)
	return claims.Username
}
