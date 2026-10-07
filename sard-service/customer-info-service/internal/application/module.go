package application

import (
	"github.com/JIeeiroSst/customer-info-service/config"
	"github.com/JIeeiroSst/customer-info-service/internal/domain"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) domain.KYCPolicy {
		return domain.KYCPolicy{MinAge: cfg.KYC.MinAge, MinConfidence: cfg.KYC.MinConfidence, RequireFaceMatch: cfg.KYC.RequireFaceMatch}
	}),
	fx.Provide(NewCustomerService),
)
