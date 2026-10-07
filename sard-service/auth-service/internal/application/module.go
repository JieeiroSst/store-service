package application

import (
	"github.com/JIeeiroSst/auth-service/config"
	"github.com/JIeeiroSst/auth-service/internal/domain"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) Settings {
		return Settings{
			Policy: domain.Policy{
				OTPTTL:       cfg.Challenge.OTPTTL,
				MaxAttempts:  cfg.Challenge.MaxAttempts,
				MaxResends:   cfg.Challenge.MaxResends,
				ValidityTime: cfg.Challenge.ValidityTime,
			},
			ExposeOTP: cfg.Challenge.ExposeOTP,
		}
	}),
	fx.Provide(NewAuthenticationService),
)
