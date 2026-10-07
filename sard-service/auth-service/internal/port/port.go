package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/domain"
)

type ChallengeRepository interface {
	Create(ctx context.Context, c *domain.Challenge) error
	Update(ctx context.Context, c *domain.Challenge) error
	Get(ctx context.Context, id string) (*domain.Challenge, error)
	ListPendingByUser(ctx context.Context, userID int64, now time.Time) ([]*domain.Challenge, error)
}

type Locker interface {
	WithLock(ctx context.Context, key string, fn func(ctx context.Context) error) error
}

type CardDirectory interface {
	ResolvePAN(ctx context.Context, pan, expiry string) (domain.CardRef, error)
}

type SessionValidator interface {
	Validate(ctx context.Context, sessionToken string) (int64, error)
}

type OTPMessage struct {
	UserID      int64
	ChallengeID string
	OTP         string
	MaskedPAN   string
	Amount      int64
	Currency    string
	Merchant    string
	ExpiresAt   time.Time
}

type OTPSender interface {
	Send(ctx context.Context, msg OTPMessage) error
}

type UserDirectory interface {
	Username(ctx context.Context, userID int64) (string, error)
}

type OTPProvider interface {
	Issue(ctx context.Context, username string) (code string, expiresAt time.Time, err error)
	Verify(ctx context.Context, username, code string) (bool, error)
}

type Clock interface {
	Now() time.Time
}

type Metrics interface {
	ObserveChallenge(event string)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
