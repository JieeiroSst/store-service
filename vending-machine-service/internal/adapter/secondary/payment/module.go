package payment

import (
	"github.com/JIeeiroSst/vending-machine-service/config"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewGateway),
)

func NewGateway(cfg *config.Config) port.PaymentGateway {
	up := cfg.Upstream
	var wallet port.PaymentGateway
	if up.WalletServiceURL != "" {
		wallet = NewWallet(httpx.New(up.WalletServiceURL, up.Timeout, up.ServiceName))
	}
	return NewRouter(wallet, Simulated{})
}
