package userservice

import (
	"github.com/JIeeiroSst/livestream-service/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) *Client {
		return New(cfg.Auth.UserServiceURL, cfg.Auth.UserServiceTimeoutDuration())
	}),
)
