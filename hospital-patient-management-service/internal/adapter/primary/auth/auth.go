// Package auth authenticates callers with the session tokens user-service
// issues. The REST gateway calls the gRPC handlers in-process, so gRPC
// interceptors never see REST traffic: the same Authenticator therefore guards
// both entry points (HTTP middleware and a unary interceptor).
package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	cacheMaxSize = 10_000

	grpcPrefix = "/hospital."
	healthPath = "/health"
)

type ctxKey struct{}

func UserID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

type entry struct {
	userID string
	expiry time.Time
}

type Authenticator struct {
	enabled   bool
	validator port.TokenValidator
	now       func() time.Time
	ttl       time.Duration

	mu    sync.Mutex
	cache map[[32]byte]entry
}

func New(cfg *config.Config, v port.TokenValidator) *Authenticator {
	return &Authenticator{
		enabled:   cfg.Auth.Mode == config.AuthToken,
		validator: v,
		now:       time.Now,
		ttl:       cfg.Auth.CacheTTL,
		cache:     map[[32]byte]entry{},
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, header string) (context.Context, error) {
	if !a.enabled {
		return ctx, nil
	}
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return ctx, model.ErrUnauthenticated
	}
	token = strings.TrimSpace(token)

	ctx = port.WithBearer(ctx, token)

	key := sha256.Sum256([]byte(token))
	if id, ok := a.cached(key); ok {
		return context.WithValue(ctx, ctxKey{}, id), nil
	}
	id, err := a.validator.Validate(ctx, token)
	if err != nil {
		return ctx, err
	}
	a.store(key, id)
	return context.WithValue(ctx, ctxKey{}, id), nil
}

func (a *Authenticator) cached(key [32]byte) (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, ok := a.cache[key]
	if !ok || !a.now().Before(e.expiry) {
		delete(a.cache, key)
		return "", false
	}
	return e.userID, true
}

func (a *Authenticator) store(key [32]byte, id string) {
	if a.ttl <= 0 {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.cache) >= cacheMaxSize {
		clear(a.cache)
	}
	a.cache[key] = entry{userID: id, expiry: a.now().Add(a.ttl)}
}

func (a *Authenticator) HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == healthPath {
			next.ServeHTTP(w, r)
			return
		}
		ctx, err := a.Authenticate(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			writeHTTPError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authenticator) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
		if !strings.HasPrefix(info.FullMethod, grpcPrefix) {
			return h(ctx, req)
		}
		var header string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if v := md.Get("authorization"); len(v) > 0 {
				header = v[0]
			}
		}
		ctx, err := a.Authenticate(ctx, header)
		if err != nil {
			return nil, grpcError(err)
		}
		return h(ctx, req)
	}
}

func grpcError(err error) error {
	if errors.Is(err, model.ErrUnauthenticated) {
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}
	logrus.WithError(err).Error("authentication unavailable")
	return status.Error(codes.Unavailable, "authentication is unavailable")
}

func writeHTTPError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if errors.Is(err, model.ErrUnauthenticated) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":16,"message":"unauthenticated"}`))
		return
	}
	logrus.WithError(err).Error("authentication unavailable")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`{"code":14,"message":"authentication is unavailable"}`))
}

var Module = fx.Options(fx.Provide(New))
