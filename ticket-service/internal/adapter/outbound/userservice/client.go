package userservice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/JIeeiroSst/ticket-service/config"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/httpx"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const (
	maxCacheEntries = 10_000
	maxLookups      = 256
	lookupWait      = 500 * time.Millisecond
)

type Client struct {
	http         *http.Client
	baseURL      string
	validatePath string
	userPath     string
	cacheTTL     time.Duration
	slots        chan struct{}

	mu    sync.Mutex
	cache map[[32]byte]cached
}

type cached struct {
	id      outbound.Identity
	expires time.Time
}

var (
	_ outbound.IdentityProvider = (*Client)(nil)
	_ outbound.UserDirectory    = (*Client)(nil)
)

func NewClient(cfg config.Config) *Client {
	return &Client{
		http:         &http.Client{Timeout: cfg.UserServiceTimeout, Transport: httpx.NewTransport(256)},
		slots:        make(chan struct{}, maxLookups),
		baseURL:      trimSlash(cfg.UserServiceURL),
		validatePath: cfg.UserServiceValidatePath,
		userPath:     cfg.UserServiceUserPath,
		cacheTTL:     cfg.AuthCacheTTL,
		cache:        map[[32]byte]cached{},
	}
}

func (c *Client) Resolve(ctx context.Context, token string) (outbound.Identity, error) {
	key := sha256.Sum256([]byte(token))
	if id, ok := c.cached(key); ok {
		return id, nil
	}

	release, err := c.enter(ctx)
	if err != nil {
		return outbound.Identity{}, err
	}
	defer release()
	userID, err := c.validate(ctx, token)
	if err != nil {
		return outbound.Identity{}, err
	}
	email, roles, err := c.profile(ctx, userID)
	if err != nil {
		return outbound.Identity{}, err
	}
	id := outbound.Identity{UserID: userID, Email: email, Roles: roles}
	c.store(key, id)
	return id, nil
}

func (c *Client) Lookup(ctx context.Context, userID int64) (outbound.Identity, error) {
	key := sha256.Sum256([]byte("user:" + strconv.FormatInt(userID, 10)))
	if id, ok := c.cached(key); ok {
		return id, nil
	}
	release, err := c.enter(ctx)
	if err != nil {
		return outbound.Identity{}, err
	}
	defer release()
	email, roles, err := c.profile(ctx, userID)
	if errors.Is(err, domain.ErrUnauthorized) {
		return outbound.Identity{}, domain.ErrNotFound
	}
	if err != nil {
		return outbound.Identity{}, err
	}
	id := outbound.Identity{UserID: userID, Email: email, Roles: roles}
	c.store(key, id)
	return id, nil
}

func (c *Client) enter(ctx context.Context) (func(), error) {
	select {
	case c.slots <- struct{}{}:
		return func() { <-c.slots }, nil
	default:
	}
	t := time.NewTimer(lookupWait)
	defer t.Stop()
	select {
	case c.slots <- struct{}{}:
		return func() { <-c.slots }, nil
	case <-t.C:
		return nil, fmt.Errorf("%w: too many identity lookups in flight", domain.ErrBusy)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) validate(ctx context.Context, token string) (int64, error) {
	body, _ := json.Marshal(map[string]string{"session_token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.validatePath, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	var out struct {
		Valid  bool   `json:"valid"`
		UserID string `json:"user_id"`
	}
	status, err := c.do(req, &out)
	if err != nil {
		return 0, err
	}
	if status == http.StatusUnauthorized || (status == http.StatusOK && !out.Valid) {
		return 0, domain.ErrUnauthorized
	}
	if status != http.StatusOK {
		return 0, upstream("validate returned %d", status)
	}
	id, err := strconv.ParseInt(out.UserID, 10, 64)
	if err != nil || id <= 0 {
		return 0, upstream("validate returned bad user_id %q", out.UserID)
	}
	return id, nil
}

func (c *Client) profile(ctx context.Context, userID int64) (string, []string, error) {
	u := c.baseURL + c.userPath + "?" + url.Values{"user_id": {strconv.FormatInt(userID, 10)}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", nil, err
	}
	var out struct {
		User struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
			Roles []struct {
				Name string `json:"name"`
			} `json:"roles"`
		} `json:"users"`
	}
	status, err := c.do(req, &out)
	if err != nil {
		return "", nil, err
	}
	if status != http.StatusOK {
		return "", nil, upstream("find user returned %d", status)
	}
	if out.User.ID == 0 {
		return "", nil, domain.ErrUnauthorized
	}
	roles := make([]string, len(out.User.Roles))
	for i, r := range out.User.Roles {
		roles[i] = r.Name
	}
	return out.User.Email, roles, nil
}

func (c *Client) do(req *http.Request, into any) (int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, upstream("%v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
			return 0, upstream("decode: %v", err)
		}
	}
	return resp.StatusCode, nil
}

func (c *Client) cached(key [32]byte) (outbound.Identity, bool) {
	if c.cacheTTL <= 0 {
		return outbound.Identity{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.cache[key]
	if !ok || time.Now().After(e.expires) {
		delete(c.cache, key)
		return outbound.Identity{}, false
	}
	return e.id, true
}

func (c *Client) store(key [32]byte, id outbound.Identity) {
	if c.cacheTTL <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.cache) >= maxCacheEntries {
		now := time.Now()
		for k, e := range c.cache {
			if now.After(e.expires) {
				delete(c.cache, k)
			}
		}
		if len(c.cache) >= maxCacheEntries {
			clear(c.cache)
		}
	}
	c.cache[key] = cached{id: id, expires: time.Now().Add(c.cacheTTL)}
}

func upstream(format string, a ...any) error {
	return fmt.Errorf("%w: user-service "+format, append([]any{domain.ErrUpstreamUnavailable}, a...)...)
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
