package userservice

import (
	"github.com/JIeeiroSst/threads-service/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) *Client {
		return New(cfg.UserService.BaseURL, cfg.UserService.TimeoutDuration())
	}),
)
