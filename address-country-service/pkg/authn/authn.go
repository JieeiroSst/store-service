package authn

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	InternalTokenHeader = "X-Internal-Token"
	identityTTL         = 30 * time.Second
	defaultUserService  = "http://user-api-svc"
)

type statusError struct {
	code int
	msg  string
}

func (e statusError) Error() string   { return e.msg }
func (e statusError) StatusCode() int { return e.code }

var (
	ErrUnauthenticated error = statusError{http.StatusUnauthorized, "authentication required"}
	ErrForbidden       error = statusError{http.StatusForbidden, "forbidden"}
	ErrUnavailable     error = statusError{http.StatusServiceUnavailable, "authentication service unavailable"}
)

var errAnonymous = errors.New("anonymous")

type Principal struct {
	UserID   string
	Username string
	Role     string
	Roles    []string
	Service  bool
}

func (p Principal) IsAdmin() bool {
	if p.Service {
		return true
	}
	for _, r := range append([]string{p.Role}, p.Roles...) {
		if r == "admin" || r == "super_admin" {
			return true
		}
	}
	return false
}

func (p Principal) Owns(userID string) bool {
	return p.IsAdmin() || (p.UserID != "" && p.UserID == userID)
}

type cacheEntry struct {
	p       Principal
	expires time.Time
}

type Authenticator struct {
	baseURL  string
	http     *http.Client
	internal [][]byte
	warnOnce sync.Once
	mu       sync.Mutex
	cache    map[string]cacheEntry
}

func New(userServiceURL string, internalTokens []string, timeout time.Duration) *Authenticator {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	a := &Authenticator{
		baseURL: strings.TrimRight(userServiceURL, "/"),
		http:    &http.Client{Timeout: timeout},
		cache:   map[string]cacheEntry{},
	}
	for _, t := range internalTokens {
		if t = strings.TrimSpace(t); t != "" {
			a.internal = append(a.internal, []byte(t))
		}
	}
	return a
}

func FromEnv() *Authenticator {
	url := os.Getenv("USER_SERVICE_URL")
	if url == "" {
		url = defaultUserService
	}
	return New(url, strings.Split(os.Getenv("INTERNAL_API_TOKENS"), ","), 3*time.Second)
}

func (a *Authenticator) Enforcing() bool { return len(a.internal) > 0 }

func (a *Authenticator) Authenticate(ctx context.Context, authorization, internalToken string) (Principal, error) {
	if internalToken = strings.TrimSpace(internalToken); internalToken != "" {
		if a.validInternal(internalToken) {
			return Principal{Service: true}, nil
		}
		return Principal{}, ErrUnauthenticated
	}
	if token, ok := strings.CutPrefix(strings.TrimSpace(authorization), "Bearer "); ok && strings.TrimSpace(token) != "" {
		return a.user(ctx, strings.TrimSpace(token))
	}
	if !a.Enforcing() {
		a.warnOnce.Do(func() {
			log.Printf("authn: INTERNAL_API_TOKENS is not set, requests without credentials are accepted as internal calls")
		})
		return Principal{Service: true}, nil
	}
	return Principal{}, errAnonymous
}

func (a *Authenticator) Request(r *http.Request) (Principal, error) {
	p, err := a.Authenticate(r.Context(), r.Header.Get("Authorization"), r.Header.Get(InternalTokenHeader))
	if errors.Is(err, errAnonymous) {
		return Principal{}, ErrUnauthenticated
	}
	return p, err
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := a.Request(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
	})
}

func (a *Authenticator) Optional(r *http.Request) (Principal, bool, error) {
	p, err := a.Authenticate(r.Context(), r.Header.Get("Authorization"), r.Header.Get(InternalTokenHeader))
	switch {
	case errors.Is(err, errAnonymous):
		return Principal{}, false, nil
	case err != nil:
		return Principal{}, false, err
	}
	return p, true, nil
}

func (a *Authenticator) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok, err := a.Optional(r)
		switch {
		case err != nil:
			WriteError(w, err)
		case ok:
			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func Require(ctx context.Context) (Principal, error) {
	p, ok := FromContext(ctx)
	if !ok {
		return Principal{}, ErrUnauthenticated
	}
	return p, nil
}

func RequireAdmin(ctx context.Context) (Principal, error) {
	p, err := Require(ctx)
	if err != nil {
		return p, err
	}
	if !p.IsAdmin() {
		return p, ErrForbidden
	}
	return p, nil
}

func RequireOwner(ctx context.Context, userID string) (Principal, error) {
	p, err := Require(ctx)
	if err != nil {
		return p, err
	}
	if !p.Owns(userID) {
		return p, ErrForbidden
	}
	return p, nil
}

func (a *Authenticator) validInternal(token string) bool {
	match := 0
	for _, want := range a.internal {
		match |= subtle.ConstantTimeCompare([]byte(token), want)
	}
	return match == 1
}

func (a *Authenticator) user(ctx context.Context, token string) (Principal, error) {
	a.mu.Lock()
	e, ok := a.cache[token]
	a.mu.Unlock()
	if ok && time.Now().Before(e.expires) {
		return e.p, nil
	}
	userID, err := a.validate(ctx, token)
	if err != nil {
		return Principal{}, err
	}
	c := readClaims(token)
	p := Principal{UserID: userID, Username: c.Username, Role: c.Role, Roles: c.Roles}
	a.mu.Lock()
	if len(a.cache) > 10_000 {
		a.cache = map[string]cacheEntry{}
	}
	a.cache[token] = cacheEntry{p: p, expires: time.Now().Add(identityTTL)}
	a.mu.Unlock()
	return p, nil
}

func (a *Authenticator) validate(ctx context.Context, token string) (string, error) {
	if a.baseURL == "" {
		log.Printf("authn: user-service url is not configured")
		return "", ErrUnavailable
	}
	body, _ := json.Marshal(map[string]string{"session_token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/v1/validate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.http.Do(req)
	if err != nil {
		log.Printf("authn: validate session: %v", err)
		return "", ErrUnavailable
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return "", ErrUnauthenticated
	case resp.StatusCode != http.StatusOK:
		log.Printf("authn: validate session returned %d", resp.StatusCode)
		return "", ErrUnavailable
	}
	var out struct {
		Valid       bool   `json:"valid"`
		UserID      string `json:"user_id"`
		UserIDCamel string `json:"userId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		log.Printf("authn: decode validate response: %v", err)
		return "", ErrUnavailable
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

type claims struct {
	Username string   `json:"username"`
	Role     string   `json:"role"`
	Roles    []string `json:"roles"`
}

func readClaims(token string) claims {
	var c claims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return c
	}
	_ = json.Unmarshal(payload, &c)
	return c
}

type contextKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	return p, ok
}

func Status(err error) int {
	var se statusError
	if errors.As(err, &se) {
		return se.code
	}
	return http.StatusUnauthorized
}

func Message(err error) string {
	var se statusError
	if errors.As(err, &se) {
		return se.msg
	}
	return "authentication required"
}

func WriteError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(Status(err))
	_ = json.NewEncoder(w).Encode(map[string]string{"error": Message(err)})
}

func SetInternalToken(h http.Header) {
	if t := strings.TrimSpace(os.Getenv("INTERNAL_API_TOKEN")); t != "" {
		h.Set(InternalTokenHeader, t)
	}
}
