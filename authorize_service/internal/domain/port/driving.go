package port

import (
	"context"

	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
)


type CasbinUsecase interface {
	Enforce(ctx context.Context, auth model.CasbinAuth) error
	ListRules(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error)
	GetRule(ctx context.Context, id int) (*model.CasbinRule, error)
	CreateRule(ctx context.Context, rule model.CasbinRule) error
	DeleteRule(ctx context.Context, id int) error
	UpdateRuleField(ctx context.Context, id int, field model.UpdateField, value string) error
}

type OTPUsecase interface {
	CreateOtpByUser(ctx context.Context, username string) (string, error)
	Authorize(ctx context.Context, otpCode string, username string) error
}
