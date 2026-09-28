package userservice

import (
	"time"

	"github.com/JIeeiroSst/basket-service/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) *Client {
		return New(cfg.UserService.BaseURL, time.Duration(cfg.UserService.TimeoutSeconds)*time.Second)
	}),
	fx.Provide(NewDirectory),
)
