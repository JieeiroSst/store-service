package http

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	cacheMaxSize  = 10_000
	userIDKey     = "user_id"
	serviceKey    = "service"
	serviceHeader = "X-Service-Key"
)

type entry struct {
	userID string
	expiry time.Time
}

type Authenticator struct {
	enabled   bool
	svcKey    string
	validator port.TokenValidator
	ttl       time.Duration
	now       func() time.Time

	mu    sync.Mutex
	cache map[[32]byte]entry
}

func NewAuthenticator(cfg *config.Config, v port.TokenValidator) *Authenticator {
	return &Authenticator{
		enabled: cfg.Auth.Mode == config.AuthToken, svcKey: cfg.Auth.ServiceKey, validator: v, ttl: cfg.Auth.CacheTTL,
		now: time.Now, cache: map[[32]byte]entry{},
	}
}

func (a *Authenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if k := c.GetHeader(serviceHeader); k != "" {
			if a.svcKey == "" || subtle.ConstantTimeCompare([]byte(k), []byte(a.svcKey)) != 1 {
				a.reject(c, model.ErrUnauthenticated)
				return
			}
			c.Set(serviceKey, true)
			c.Next()
			return
		}
		if !a.enabled {
			c.Next()
			return
		}
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token = strings.TrimSpace(token); !ok || token == "" {
			a.reject(c, model.ErrUnauthenticated)
			return
		}
		key := sha256.Sum256([]byte(token))
		id, ok := a.cached(key)
		if !ok {
			var err error
			if id, err = a.validator.Validate(c.Request.Context(), token); err != nil {
				a.reject(c, err)
				return
			}
			a.store(key, id)
		}
		c.Set(userIDKey, id)
		c.Next()
	}
}

func (a *Authenticator) reject(c *gin.Context, err error) {
	if errors.Is(err, model.ErrUnauthenticated) {
		c.Header("WWW-Authenticate", "Bearer")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	logrus.WithError(err).Error("authentication unavailable")
	c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "authentication is unavailable"})
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

func isService(c *gin.Context) bool {
	v, _ := c.Get(serviceKey)
	ok, _ := v.(bool)
	return ok
}
