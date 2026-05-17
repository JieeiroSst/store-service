package otp

import (
	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"go.uber.org/fx"
)

func newOTPPortFromConfig(cfg *config.Config) port.OTPPort {
	return NewOTPAdapter(cfg.Secret.JwtSecretKey)
}

var Module = fx.Options(
	fx.Provide(newOTPPortFromConfig),
)
