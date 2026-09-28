package userservice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUserNotFound    = errors.New("user not found")
	ErrUpstream        = errors.New("user-service unavailable")
)

type Identity struct {
	UserID   string
	Username string
	Role     string   
	Roles    []string 
}

func (i Identity) HasRole(role string) bool {
	for _, r := range i.Roles {
		if r == role {
			return true
		}
	}
	return i.Role == role
}

func (i Identity) IsAdmin() bool { return i.Role == "admin" || i.Role == "super_admin" }

type User struct {
	ID       string
	Username string
	Email    string
	Name     string
	Phone    string
	Address  string
	Active   bool
}

const identityTTL = 30 * time.Second

type cacheEntry struct {
	id      Identity
	expires time.Time
}

type Client struct {
	baseURL string
	http    *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
		cache:   map[string]cacheEntry{},
	}
}

func (c *Client) Authenticate(ctx context.Context, token string) (Identity, error) {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return Identity{}, ErrUnauthenticated
	}

	c.mu.Lock()
	e, ok := c.cache[token]
	c.mu.Unlock()
	if ok && time.Now().Before(e.expires) {
		return e.id, nil
	}

	userID, err := c.validate(ctx, token)
	if err != nil {
		return Identity{}, err
	}

	claims := readClaims(token)
	id := Identity{UserID: userID, Username: claims.Username, Role: claims.Role, Roles: claims.Roles}

	c.mu.Lock()
	if len(c.cache) > 10_000 {
		c.cache = map[string]cacheEntry{}
	}
	c.cache[token] = cacheEntry{id: id, expires: time.Now().Add(identityTTL)}
	c.mu.Unlock()
	return id, nil
}

func (c *Client) validate(ctx context.Context, token string) (string, error) {
	if c.baseURL == "" {
		return "", fmt.Errorf("%w: base url not configured", ErrUpstream)
	}
	body, _ := json.Marshal(map[string]string{"session_token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/validate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return "", ErrUnauthenticated
	case resp.StatusCode != http.StatusOK:
		return "", fmt.Errorf("%w: validate returned %d", ErrUpstream, resp.StatusCode)
	}

	var out struct {
		Valid       bool   `json:"valid"`
		UserID      string `json:"user_id"`
		UserIDCamel string `json:"userId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	id := out.UserID
	if id == "" {
		id = out.UserIDCamel
	}
	if !out.Valid || id == "" {
		return "", ErrUnauthenticated
	}
	return id, nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (User, error) {
	if c.baseURL == "" {
		return User{}, fmt.Errorf("%w: base url not configured", ErrUpstream)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/user?user_id="+url.QueryEscape(userID), nil)
	if err != nil {
		return User{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return User{}, ErrUserNotFound
	case resp.StatusCode != http.StatusOK:
		return User{}, fmt.Errorf("%w: find user returned %d", ErrUpstream, resp.StatusCode)
	}

	var out struct {
		Users struct {
			ID       json.Number `json:"id"`
			Username string      `json:"username"`
			Email    string      `json:"email"`
			Name     string      `json:"name"`
			Phone    string      `json:"phone"`
			Address  string      `json:"address"`
			Checked  bool        `json:"checked"`
		} `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return User{}, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	u := out.Users
	if u.ID == "" || u.ID == "0" {
		return User{}, ErrUserNotFound
	}
	return User{
		ID: u.ID.String(), Username: u.Username, Email: u.Email,
		Name: u.Name, Phone: u.Phone, Address: u.Address, Active: u.Checked,
	}, nil
}

type tokenClaims struct {
	Username string   `json:"username"`
	Role     string   `json:"role"`
	Roles    []string `json:"roles"`
}

func readClaims(token string) tokenClaims {
	var c tokenClaims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return c
	}
	_ = json.Unmarshal(raw, &c)
	return c
}
