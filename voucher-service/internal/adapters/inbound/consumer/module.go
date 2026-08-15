package consumer

import (
	"context"

	auditapp "github.com/JIeeiroSst/voucher-service/internal/application/audit"
	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type runner interface {
	Run(ctx context.Context)
	Close() error
}

func registerConsumers(lc fx.Lifecycle, factory ReaderFactory, orderSvc orderapp.OrderService, auditSvc auditapp.AuditService, log *zap.Logger) {
	consumers := []runner{
		NewPaymentSettledConsumer(factory, orderSvc, log),
		NewAuditConsumer("voucher.events", factory, auditSvc, log),
		NewAuditConsumer("order.events", factory, auditSvc, log),
	}

	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			for _, r := range consumers {
				go r.Run(ctx)
			}
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			for _, r := range consumers {
				_ = r.Close()
			}
			return nil
		},
	})
}

var Module = fx.Module("kafka-consumers",
	fx.Invoke(registerConsumers),
)
