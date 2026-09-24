// Package auth verifies access tokens issued by user-service (HS256 JWT with
// "sub", "username" and "role" claims) and enforces role based access on gRPC methods.
package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Claims struct {
	Subject  string
	Username string
	Role     string
}

type ctxKey struct{}

// FromContext returns the caller's claims, or nil for an anonymous call.
func FromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

// Level is the minimum privilege a method demands.
type Level int

const (
	Public   Level = iota // no token needed; a valid token is still parsed if sent
	Customer              // any authenticated user
	Staff                 // staff or admin
	Admin                 // admin only
)

type Config struct {
	SecretKey string
	AdminRole string // role claim that maps to admin, default "admin"
	StaffRole string // role claim that maps to staff, default "staff"
}

type Authenticator struct {
	secret []byte
	admin  string
	staff  string
	// Methods maps a full gRPC method name to its level; unlisted methods
	// require an authenticated user.
	methods map[string]Level
}

func New(cfg Config, methods map[string]Level) (*Authenticator, error) {
	if cfg.SecretKey == "" {
		return nil, errors.New("auth: jwt secret key is empty")
	}
	a := &Authenticator{secret: []byte(cfg.SecretKey), admin: cfg.AdminRole, staff: cfg.StaffRole, methods: methods}
	if a.admin == "" {
		a.admin = "admin"
	}
	if a.staff == "" {
		a.staff = "staff"
	}
	return a, nil
}

func (a *Authenticator) IsAdmin(c *Claims) bool { return c != nil && c.Role == a.admin }
func (a *Authenticator) IsStaff(c *Claims) bool {
	return c != nil && (c.Role == a.staff || c.Role == a.admin)
}

// Parse validates a bearer token. Only HS256 is accepted, which stops tokens
// signed with "none" or an unexpected algorithm.
func (a *Authenticator) Parse(token string) (*Claims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return a.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	m, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token")
	}
	c := &Claims{}
	c.Subject, _ = m["sub"].(string)
	c.Username, _ = m["username"].(string)
	c.Role, _ = m["role"].(string)
	if c.Subject == "" {
		return nil, errors.New("invalid token")
	}
	return c, nil
}

// UnaryInterceptor authenticates the call and applies the method's level.
func (a *Authenticator) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		level, listed := a.methods[info.FullMethod]
		if !listed {
			level = Customer
		}

		var claims *Claims
		if tok := bearer(ctx); tok != "" {
			c, err := a.Parse(tok)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
			}
			claims = c
			ctx = context.WithValue(ctx, ctxKey{}, c)
		}

		switch {
		case level == Public:
		case claims == nil:
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		case level == Staff && !a.IsStaff(claims), level == Admin && !a.IsAdmin(claims):
			return nil, status.Error(codes.PermissionDenied, "insufficient role")
		}
		return next(ctx, req)
	}
}

func bearer(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, v := range md.Get("authorization") {
		if len(v) > 7 && strings.EqualFold(v[:7], "bearer ") {
			return strings.TrimSpace(v[7:])
		}
	}
	return ""
}
