package kafka

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewWriter(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) *kafkago.Writer {
	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(cfg.KafkaBrokers...),
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing kafka writer")
			return w.Close()
		},
	})

	return w
}

type ReaderFactory func(topic, groupID string) *kafkago.Reader

func NewReaderFactory(cfg *config.Config) ReaderFactory {
	return func(topic, groupID string) *kafkago.Reader {
		return kafkago.NewReader(kafkago.ReaderConfig{
			Brokers: cfg.KafkaBrokers,
			Topic:   topic,
			GroupID: groupID,
		})
	}
}
