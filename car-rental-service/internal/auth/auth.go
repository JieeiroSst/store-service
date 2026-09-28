package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/JIeeiroSst/car-rental-service/internal/userservice"
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
	AdminRole string // authorize-service role that maps to admin; "super_admin" always counts
	StaffRole string // authorize-service role that maps to staff, default "operator"
}

type Identities interface {
	Authenticate(ctx context.Context, token string) (userservice.Identity, error)
}

type Authenticator struct {
	users   Identities
	admin   string
	staff   string
	methods map[string]Level
}

func New(cfg Config, users Identities, methods map[string]Level) (*Authenticator, error) {
	if users == nil {
		return nil, errors.New("auth: user-service client is required")
	}
	a := &Authenticator{users: users, admin: cfg.AdminRole, staff: cfg.StaffRole, methods: methods}
	if a.admin == "" {
		a.admin = "admin"
	}
	if a.staff == "" {
		a.staff = "operator"
	}
	return a, nil
}

func (a *Authenticator) IsAdmin(c *Claims) bool {
	return c != nil && (c.Role == a.admin || c.Role == "super_admin")
}

func (a *Authenticator) IsStaff(c *Claims) bool {
	return c != nil && (c.Role == a.staff || a.IsAdmin(c))
}

func (a *Authenticator) Parse(ctx context.Context, token string) (*Claims, error) {
	id, err := a.users.Authenticate(ctx, token)
	if err != nil {
		return nil, err
	}
	return &Claims{Subject: id.UserID, Username: id.Username, Role: id.Role}, nil
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
			c, err := a.Parse(ctx, tok)
			if errors.Is(err, userservice.ErrUpstream) {
				return nil, status.Error(codes.Unavailable, "authentication service unavailable")
			}
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
