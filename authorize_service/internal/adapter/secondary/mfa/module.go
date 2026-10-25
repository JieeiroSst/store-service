package mfa

import (
	"context"
	"time"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/config"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func newRepository(db *gorm.DB) (port.TOTPRepository, error) {
	if err := db.AutoMigrate(&model.UserTOTP{}); err != nil {
		return nil, err
	}
	return NewRepository(db), nil
}

func newSecretBox(cfg *config.Config) (port.SecretBox, error) {
	key := cfg.MFA.EncryptionKey
	if key == "" {
		logger.WithContext(context.Background()).Warn("mfa: encryption_key is not configured, falling back to jwt_secret_key; rotating it will invalidate enrolled authenticators")
		key = cfg.Secret.JwtSecretKey
	}
	return NewSecretBox(key)
}

func newQRCode() port.QRCodeEncoder { return QRCode{} }

func newThrottle(lc fx.Lifecycle, cfg *config.Config) port.FailureThrottle {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Cache.Host})
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return rdb.Close() }})
	return NewThrottle(rdb, 5, 15*time.Minute)
}

var Module = fx.Options(
	fx.Provide(newRepository, newSecretBox, newQRCode, newThrottle),
)
