package userservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/wishlists-service/internal/config"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/pkg/credential"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.UserService.URL, "/"),
		http:    &http.Client{Timeout: cfg.UserService.Timeout},
	}
}

type userResponse struct {
	ID       json.Number `json:"id"`
	Username string      `json:"username"`
	Name     string      `json:"name"`
	Active   bool        `json:"active"`
}

func (c *Client) GetUser(ctx context.Context, id string) (*model.Owner, error) {
	if n, err := strconv.ParseInt(id, 10, 32); err != nil || n <= 0 {
		return nil, model.ErrNotFound
	}
	cred := credential.From(ctx)
	if cred.Empty() {
		return nil, model.ErrUnauthenticated
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/users/"+url.PathEscape(id)+"/profile", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	cred.Apply(req.Header)
	resp, err := c.http.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%w: user-service: %v", model.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, model.ErrNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, model.ErrUnauthenticated
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		slog.Warn("user-service lookup failed", "status", resp.StatusCode, "user_id", id, "body", string(body))
		return nil, fmt.Errorf("%w: user-service returned %d", model.ErrUnavailable, resp.StatusCode)
	}
	var u userResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&u); err != nil {
		return nil, fmt.Errorf("%w: decode user-service response: %v", model.ErrUnavailable, err)
	}
	if u.ID.String() != id {
		return nil, fmt.Errorf("%w: user-service returned user %s for %s", model.ErrUnavailable, u.ID, id)
	}
	return &model.Owner{ID: id, Username: u.Username, Name: u.Name, Active: u.Active}, nil
}
