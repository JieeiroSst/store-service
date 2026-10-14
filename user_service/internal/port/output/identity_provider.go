package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type IdentityProvider interface {
	AuthCodeURL(ctx context.Context, state, verifier string) (string, error)
	Exchange(ctx context.Context, code, verifier string) (domain.ExternalIdentity, error)
	VerifyToken(ctx context.Context, token string) (domain.ExternalIdentity, error)
}
