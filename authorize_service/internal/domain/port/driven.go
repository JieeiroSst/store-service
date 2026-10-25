package port

import (
	"context"
	"time"

	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
)

type CasbinRepository interface {
	CasbinRuleAll(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error)
	CasbinRuleByID(ctx context.Context, id int) (*model.CasbinRule, error)
	CreateCasbinRule(ctx context.Context, rule model.CasbinRule) error
	DeleteCasbinRule(ctx context.Context, id int) error
	UpdateCasbinRuleField(ctx context.Context, id int, field model.UpdateField, value string) error
}

type CachePort interface {
	GetInterface(ctx context.Context, key string, value interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	GetInt(ctx context.Context, key string) (int, error)
	SetInt(ctx context.Context, key string, value int) error
}

type TOTPRepository interface {
	Get(ctx context.Context, userID string) (model.UserTOTP, error)
	SavePending(ctx context.Context, userID, sealedSecret string) error
	Enable(ctx context.Context, userID string, step int64, recoveryCodes string, at time.Time) error
	MarkUsed(ctx context.Context, userID string, step int64) (bool, error)
	SwapRecoveryCodes(ctx context.Context, userID, old, new string) (bool, error)
	Delete(ctx context.Context, userID string) error
}

type SecretBox interface {
	Seal(plaintext []byte) (string, error)
	Open(sealed string) ([]byte, error)
}

type QRCodeEncoder interface {
	PNG(content string, size int) ([]byte, error)
}

type FailureThrottle interface {
	Blocked(ctx context.Context, key string) (bool, error)
	RecordFailure(ctx context.Context, key string) error
	Reset(ctx context.Context, key string) error
}

type OTPPort interface {
	GenerateOTP(username string) (string, time.Time, error)
	ValidateOTP(otpCode, username string) error
}
