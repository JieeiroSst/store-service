package http

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/auth"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Authenticator struct {
	apiKey     string
	apiKeyRole auth.Role
	oidc       *oidcVerifier
}

func NewAuthenticator(cfg *config.Config) *Authenticator {
	a := &Authenticator{apiKey: cfg.Server.APIKey, apiKeyRole: auth.Admin}
	if r, ok := auth.ParseRole(cfg.Auth.APIKeyRole, ""); ok {
		a.apiKeyRole = r
	}
	if cfg.Auth.OIDCIssuer != "" {
		a.oidc = newOIDCVerifier(cfg.Auth)
	}
	if a.apiKey == "" && a.oidc == nil {
		logrus.Warn("authentication is disabled: no API_KEY and no OIDC_ISSUER; every caller is an administrator")
	}
	return a
}

func (a *Authenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.apiKey == "" && a.oidc == nil {
			c.Request = c.Request.WithContext(auth.WithPrincipal(c.Request.Context(),
				auth.Principal{Subject: "anonymous", Roles: []auth.Role{auth.Admin}, Service: true, IP: c.ClientIP()}))
			c.Next()
			return
		}

		p, err := a.authenticate(c)
		if err != nil {
			status := http.StatusUnauthorized
			if err == errAuthUnavailable {
				status = http.StatusServiceUnavailable
			}
			c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
			return
		}
		p.IP = c.ClientIP()
		c.Request = c.Request.WithContext(auth.WithPrincipal(c.Request.Context(), p))
		c.Next()
	}
}

var (
	errUnauthenticated = fmt.Errorf("authentication required: send a bearer token or X-API-Key")
	errAuthUnavailable = fmt.Errorf("the identity provider is unavailable")
)

func (a *Authenticator) authenticate(c *gin.Context) (auth.Principal, error) {
	if h := c.GetHeader("Authorization"); a.oidc != nil && strings.HasPrefix(h, "Bearer ") {
		return a.oidc.verify(c.Request.Context(), strings.TrimPrefix(h, "Bearer "))
	}
	if key := c.GetHeader("X-API-Key"); a.apiKey != "" && key != "" {
		if subtle.ConstantTimeCompare([]byte(key), []byte(a.apiKey)) != 1 {
			return auth.Principal{}, fmt.Errorf("invalid API key")
		}
		return auth.Principal{Subject: "api-key", Name: "API key", Roles: []auth.Role{a.apiKeyRole}, Service: true}, nil
	}
	return auth.Principal{}, errUnauthenticated
}

type oidcVerifier struct {
	cfg config.AuthConfig

	mu      sync.Mutex
	keys    keyfunc.Keyfunc
	retryAt time.Time
}

func newOIDCVerifier(cfg config.AuthConfig) *oidcVerifier {
	if cfg.OIDCJWKSURL == "" {
		cfg.OIDCJWKSURL = cfg.OIDCIssuer + "/protocol/openid-connect/certs"
	}
	return &oidcVerifier{cfg: cfg}
}

func (v *oidcVerifier) keyfunc() (keyfunc.Keyfunc, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.keys != nil {
		return v.keys, nil
	}
	if time.Now().Before(v.retryAt) {
		return nil, errAuthUnavailable
	}
	failFirst := false
	k, err := keyfunc.NewDefaultOverrideCtx(context.Background(), []string{v.cfg.OIDCJWKSURL}, keyfunc.Override{
		HTTPTimeout:               5 * time.Second,
		NoErrorReturnFirstHTTPReq: &failFirst,
		RefreshInterval:           15 * time.Minute,
		RefreshUnknownKID:         rate.NewLimiter(rate.Every(time.Minute), 1),
	})
	if err != nil {
		v.retryAt = time.Now().Add(10 * time.Second)
		logrus.WithError(err).Warn("cannot load the OIDC signing keys")
		return nil, errAuthUnavailable
	}
	v.keys = k
	return k, nil
}

var signingMethods = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384", "ES512"}

func (v *oidcVerifier) verify(_ context.Context, token string) (auth.Principal, error) {
	kf, err := v.keyfunc()
	if err != nil {
		return auth.Principal{}, err
	}
	opts := []jwt.ParserOption{
		jwt.WithValidMethods(signingMethods),
		jwt.WithIssuer(v.cfg.OIDCIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30 * time.Second),
	}
	if v.cfg.OIDCAudience != "" {
		opts = append(opts, jwt.WithAudience(v.cfg.OIDCAudience))
	}
	claims := jwt.MapClaims{}
	if _, err := jwt.ParseWithClaims(token, claims, kf.Keyfunc, opts...); err != nil {
		return auth.Principal{}, fmt.Errorf("invalid bearer token")
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return auth.Principal{}, fmt.Errorf("invalid bearer token")
	}
	p := auth.Principal{Subject: sub, Name: displayName(claims)}
	for _, name := range claimStrings(claims, v.cfg.RolesClaim) {
		if r, ok := auth.ParseRole(name, v.cfg.RolePrefix); ok {
			p.Roles = append(p.Roles, r)
		}
	}
	return p, nil
}

func displayName(claims jwt.MapClaims) string {
	for _, k := range []string{"preferred_username", "name", "email"} {
		if s, _ := claims[k].(string); s != "" {
			return s
		}
	}
	return ""
}

func claimStrings(claims map[string]any, path string) []string {
	var cur any = map[string]any(claims)
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[part]
	}
	switch v := cur.(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, x := range v {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
