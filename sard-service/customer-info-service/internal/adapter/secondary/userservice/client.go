package userservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-info-service/config"
	"github.com/JIeeiroSst/customer-info-service/internal/domain"
	"github.com/JIeeiroSst/customer-info-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) port.UserDirectory {
		return NewClient(cfg.UserService.BaseURL, cfg.UserService.Token, cfg.UserService.Timeout)
	}),
)

const maxResponseBytes = 1 << 20

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: timeout},
	}
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Sex      string `json:"sex"`
	Active   bool   `json:"active"`
}

func (c *Client) GetUser(ctx context.Context, id int64) (domain.UserProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/v1/users/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		return domain.UserProfile{}, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("%w: user-service: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("%w: user-service: %v", domain.ErrUnavailable, err)
	}
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return domain.UserProfile{}, domain.ErrUserNotFound
	case resp.StatusCode != http.StatusOK:
		return domain.UserProfile{}, fmt.Errorf("%w: user-service returned %s", domain.ErrUnavailable, resp.Status)
	}
	var u userResponse
	if err := json.Unmarshal(body, &u); err != nil {
		return domain.UserProfile{}, fmt.Errorf("%w: user-service: decode: %v", domain.ErrUnavailable, err)
	}
	if u.ID != id {
		return domain.UserProfile{}, fmt.Errorf("%w: user-service returned user %d for id %d", domain.ErrUnavailable, u.ID, id)
	}
	return domain.UserProfile{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Name:     u.Name,
		Phone:    u.Phone,
		Address:  u.Address,
		Sex:      u.Sex,
		Active:   u.Active,
	}, nil
}
