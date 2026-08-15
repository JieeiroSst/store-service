package publisher

import (
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"go.uber.org/fx"
)

var Module = fx.Module("kafka-publisher",
	fx.Provide(fx.Annotate(NewKafkaPublisher, fx.As(new(outbox.EventPublisher)))),
)
