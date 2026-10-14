package output

import (
	"context"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, s domain.LoginSession) error
	Get(ctx context.Context, id string) (domain.LoginSession, error)
	Touch(ctx context.Context, id string, lastActive, expiresAt time.Time) error
	ListActive(ctx context.Context, userID int, now time.Time) ([]domain.LoginSession, error)
	Revoke(ctx context.Context, ids []string, reason string, at time.Time) error
}
