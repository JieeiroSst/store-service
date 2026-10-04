package application

import (
	"crypto/subtle"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
)

type Authorizer struct {
	defaultAllow bool
	management   string
	read         []string
}

func NewAuthorizer(cfg *config.Config) *Authorizer {
	return &Authorizer{
		defaultAllow: cfg.ACL.DefaultPolicy != "deny",
		management:   cfg.ACL.ManagementToken,
		read:         cfg.ACL.ReadTokens,
	}
}

func (a *Authorizer) DefaultPolicy() string {
	if a.defaultAllow {
		return "allow"
	}
	return "deny"
}

func (a *Authorizer) Authorize(token string, write bool) error {
	if a.defaultAllow || (token != "" && equal(token, a.management)) {
		return nil
	}
	if !write {
		for _, t := range a.read {
			if equal(token, t) {
				return nil
			}
		}
	}
	return domain.ErrPermissionDenied
}

func equal(a, b string) bool {
	return b != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
