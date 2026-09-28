package coupon

import (
	"github.com/JIeeiroSst/vending-machine-service/config"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewProvider),
)

func NewProvider(cfg *config.Config) port.CouponProvider {
	up := cfg.Upstream
	if up.CouponServiceURL == "" {
		return Disabled{}
	}
	return NewClient(httpx.New(up.CouponServiceURL, up.Timeout, up.ServiceName))
}
