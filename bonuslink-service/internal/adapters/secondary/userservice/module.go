package userservice

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/bonuslink-service/internal/config"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) *Client {
		return New(cfg.UserService.BaseURL, cfg.UserService.Timeout)
	}),
)
