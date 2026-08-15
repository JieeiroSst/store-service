package outbox

import (
	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newRelay(cfg *config.Config, repo Repository, publisher EventPublisher, log *zap.Logger) *Relay {
	return NewRelay(repo, publisher, cfg.OutboxRelayInterval, log)
}

var Module = fx.Module("outbox-relay",
	fx.Provide(newRelay),
	fx.Invoke(RegisterRelay),
)
