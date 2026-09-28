package notification

import (
	"github.com/JIeeiroSst/vending-machine-service/config"
	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewNotifier),
)

func NewNotifier(cfg *config.Config) port.Notifier {
	up := cfg.Upstream
	if up.NotificationURL == "" {
		return Log{}
	}
	return NewSlack(httpx.New(up.NotificationURL, up.Timeout, up.ServiceName))
}
