package userservice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Directory struct {
	client  pb.UserServiceClient
	timeout time.Duration
	ttl     time.Duration

	mu    sync.Mutex
	cache map[string]cached
}

type cached struct {
	user    model.User
	expires time.Time
}

func NewDirectory(client pb.UserServiceClient, cfg *config.Config) *Directory {
	return &Directory{
		client:  client,
		timeout: cfg.UserService.Timeout,
		ttl:     cfg.UserService.AuthCacheTTL,
		cache:   map[string]cached{},
	}
}

func (d *Directory) Login(ctx context.Context, username, password string) (model.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	res, err := d.client.Login(ctx, &pb.LoginRequest{Username: username, Password: password})
	if err != nil {
		return model.Session{}, mapError(err, port.ErrInvalidLogin)
	}
	return model.Session{AccessToken: res.SessionToken, RefreshToken: res.RefreshToken, ExpiresIn: res.ExpiryTime}, nil
}

func (d *Directory) Refresh(ctx context.Context, refreshToken string) (model.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	res, err := d.client.RefreshToken(ctx, &pb.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return model.Session{}, mapError(err, port.ErrUnauthenticated)
	}
	return model.Session{AccessToken: res.NewSessionToken, RefreshToken: res.NewRefreshToken, ExpiresIn: res.ExpiryTime}, nil
}

func (d *Directory) Validate(ctx context.Context, token string) (model.User, error) {
	if u, ok := d.lookup(token); ok {
		return u, nil
	}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	res, err := d.client.ValidateSession(ctx, &pb.ValidateRequest{SessionToken: token})
	if err != nil {
		return model.User{}, mapError(err, port.ErrUnauthenticated)
	}
	if !res.Valid {
		return model.User{}, port.ErrUnauthenticated
	}
	id, err := strconv.Atoi(res.UserId)
	if err != nil {
		return model.User{}, port.ErrUnauthenticated
	}
	username := usernameClaim(token)
	if username == "" {
		return model.User{}, port.ErrUnauthenticated
	}

	u := model.User{ID: id, Username: username}
	d.store(token, u)
	return u, nil
}

func (d *Directory) lookup(token string) (model.User, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	c, ok := d.cache[token]
	if !ok {
		return model.User{}, false
	}
	if time.Now().After(c.expires) {
		delete(d.cache, token)
		return model.User{}, false
	}
	return c.user, true
}

func (d *Directory) store(token string, u model.User) {
	if d.ttl <= 0 {
		return
	}
	now := time.Now()
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.cache) >= 1024 {
		for t, c := range d.cache {
			if now.After(c.expires) {
				delete(d.cache, t)
			}
		}
	}
	d.cache[token] = cached{user: u, expires: now.Add(d.ttl)}
}

func usernameClaim(token string) string {
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
	if json.Unmarshal(raw, &claims) != nil {
		return ""
	}
	return claims.Username
}

func mapError(err error, fallback error) error {
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
		return port.ErrUnavailable
	}
	return fallback
}
