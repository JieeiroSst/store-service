package userservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
	"github.com/JIeeiroSst/toggle-service/internal/infrastructure/config"
)

type client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) port.UserDirectory {
	return &client{
		baseURL:    strings.TrimRight(cfg.UserService.BaseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type signUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type signUpResponse struct {
	Message string     `json:"message"`
	User    remoteUser `json:"user"`
}

type findUserResponse struct {
	User remoteUser `json:"users"`
}

type remoteUser struct {
	ID       int32        `json:"id"`
	Username string       `json:"username"`
	Email    string       `json:"email"`
	Roles    []remoteRole `json:"roles"`
}

type remoteRole struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (c *client) Register(ctx context.Context, email, username, password string) (*model.User, error) {
	body, err := json.Marshal(signUpRequest{Username: username, Password: password, Email: email})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/user/sign-up", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user_service sign-up: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, apperr.ErrConflict
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("user_service sign-up: unexpected status %d", resp.StatusCode)
	}

	var out signUpResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("user_service sign-up: decode response: %w", err)
	}
	return toDomainUser(out.User), nil
}

func (c *client) Login(ctx context.Context, username, password string) (*model.User, error) {
	body, err := json.Marshal(loginRequest{Username: username, Password: password})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/login", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user_service login: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
		return nil, apperr.ErrInvalidCredentials
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("user_service login: unexpected status %d", resp.StatusCode)
	}

	return c.findByUsername(ctx, username)
}

func (c *client) findByUsername(ctx context.Context, username string) (*model.User, error) {
	q := url.Values{"username": []string{username}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/user?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user_service find user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperr.ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("user_service find user: unexpected status %d", resp.StatusCode)
	}

	var out findUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("user_service find user: decode response: %w", err)
	}
	return toDomainUser(out.User), nil
}

func toDomainUser(u remoteUser) *model.User {
	isAdmin := false
	for _, r := range u.Roles {
		if strings.EqualFold(r.Name, "admin") {
			isAdmin = true
			break
		}
	}
	return &model.User{
		ID:       strconv.Itoa(int(u.ID)),
		Username: u.Username,
		Email:    u.Email,
		IsAdmin:  isAdmin,
	}
}
