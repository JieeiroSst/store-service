package kafka

import "go.uber.org/fx"

var Module = fx.Module("kafka-client", fx.Provide(NewWriter, NewReaderFactory))
