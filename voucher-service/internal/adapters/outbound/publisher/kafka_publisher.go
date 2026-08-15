package publisher

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	kafkago "github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafkago.Writer
}

func NewKafkaPublisher(writer *kafkago.Writer) outbox.EventPublisher {
	return &KafkaPublisher{writer: writer}
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic, key string, payload []byte) error {
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	})
}
