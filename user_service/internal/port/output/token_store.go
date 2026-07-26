package output

import (
	"context"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

// TokenStore persists issued access/refresh token pairs so a stateless JWT
// access token can still be revoked on logout, and a refresh token can be
// validated and rotated when it's redeemed.
type TokenStore interface {
	SaveSession(ctx context.Context, session domain.Session, accessTTL, refreshTTL time.Duration) error
	GetSessionByAccessToken(ctx context.Context, accessToken string) (domain.Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (domain.Session, error)
	DeleteSession(ctx context.Context, accessToken, refreshToken string) error
}
