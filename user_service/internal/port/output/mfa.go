package output

import (
	"context"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type MFAProvider interface {
	Status(ctx context.Context, userID int) (domain.MFAStatus, error)
	SetupTOTP(ctx context.Context, userID int, account string) (domain.TOTPSetup, error)
	EnableTOTP(ctx context.Context, userID int, code string) ([]string, error)
	DisableTOTP(ctx context.Context, userID int, code string) error
	RegenerateRecoveryCodes(ctx context.Context, userID int, code string) ([]string, error)
	VerifyTOTP(ctx context.Context, userID int, code string) error
}

type MFAChallengeStore interface {
	Save(ctx context.Context, token string, challenge domain.MFAChallenge, ttl time.Duration) error
	Get(ctx context.Context, token string) (domain.MFAChallenge, error)
	Delete(ctx context.Context, token string) (bool, error)
}
