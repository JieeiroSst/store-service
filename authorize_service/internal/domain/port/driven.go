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

type OTPPort interface {
	GenerateOTP(username string) (string, error)
	ValidateOTP(otpCode, username string) error
}
