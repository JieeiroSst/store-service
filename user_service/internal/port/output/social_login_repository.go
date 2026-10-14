package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type SocialLoginRepository interface {
	Find(ctx context.Context, provider, subject string) (domain.SocialLogin, error)
	Create(ctx context.Context, login domain.SocialLogin) error
}
