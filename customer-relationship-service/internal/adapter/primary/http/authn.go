package http

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/userservice"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/auth"
	"github.com/gin-gonic/gin"
)

type Authenticator struct {
	apiKey     string
	apiKeyRole auth.Role
	rolePrefix string
	users      *userservice.Client
}

func NewAuthenticator(cfg *config.Config) *Authenticator {
	a := &Authenticator{
		apiKey:     cfg.Server.APIKey,
		apiKeyRole: auth.Admin,
		rolePrefix: cfg.Auth.RolePrefix,
		users:      userservice.New(cfg.Auth.UserServiceURL, cfg.Auth.UserServiceTimeout),
	}
	if r, ok := auth.ParseRole(cfg.Auth.APIKeyRole, ""); ok {
		a.apiKeyRole = r
	}
	return a
}

func (a *Authenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		p, err := a.authenticate(c)
		if err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, errAuthUnavailable) {
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
	errUnauthenticated = fmt.Errorf("authentication required: send a user-service bearer token or X-API-Key")
	errAuthUnavailable = fmt.Errorf("user-service is unavailable")
)

func (a *Authenticator) authenticate(c *gin.Context) (auth.Principal, error) {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		id, err := a.users.Authenticate(c.Request.Context(), strings.TrimPrefix(h, "Bearer "))
		if errors.Is(err, userservice.ErrUpstream) {
			return auth.Principal{}, errAuthUnavailable
		}
		if err != nil {
			return auth.Principal{}, fmt.Errorf("invalid bearer token")
		}
		p := auth.Principal{Subject: id.UserID, Name: id.Username}
		if r, ok := a.crmRole(id); ok {
			p.Roles = []auth.Role{r}
		}
		return p, nil
	}
	if key := c.GetHeader("X-API-Key"); a.apiKey != "" && key != "" {
		if subtle.ConstantTimeCompare([]byte(key), []byte(a.apiKey)) != 1 {
			return auth.Principal{}, fmt.Errorf("invalid API key")
		}
		return auth.Principal{Subject: "api-key", Name: "API key", Roles: []auth.Role{a.apiKeyRole}, Service: true}, nil
	}
	return auth.Principal{}, errUnauthenticated
}

func (a *Authenticator) crmRole(id userservice.Identity) (auth.Role, bool) {
	var best auth.Principal
	for _, name := range append([]string{id.Role}, id.Roles...) {
		if r, ok := a.mapRole(name); ok {
			best.Roles = append(best.Roles, r)
		}
	}
	if len(best.Roles) == 0 {
		return "", false
	}
	return best.Highest(), true
}

func (a *Authenticator) mapRole(name string) (auth.Role, bool) {
	if a.rolePrefix != "" {
		if r, ok := auth.ParseRole(name, a.rolePrefix); ok {
			return r, true
		}
	}
	switch name {
	case "admin", "super_admin":
		return auth.Admin, true
	case "operator":
		return auth.Staff, true
	}
	return "", false
}
