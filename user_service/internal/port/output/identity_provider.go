package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type IdentityProvider interface {
	AuthCodeURL(ctx context.Context, state string) (string, error)
	Exchange(ctx context.Context, code string) (domain.ExternalIdentity, error)
}
