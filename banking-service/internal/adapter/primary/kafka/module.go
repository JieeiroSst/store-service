package kafka

import (
	"github.com/JieeiroSst/banking-service/pkg/idempotency"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewConsumerGroup),
	fx.Provide(idempotency.NewIdempotencyGuard),
	fx.Provide(NewTransactionConsumer),
	fx.Invoke(Run),
)
