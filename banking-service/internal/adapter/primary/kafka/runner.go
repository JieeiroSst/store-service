package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/JieeiroSst/banking-service/config"
	"github.com/Shopify/sarama"
	"go.uber.org/fx"
)

func NewConsumerGroup(cfg *config.Config) (sarama.ConsumerGroup, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Return.Errors = true
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("kafka: connect consumer group: %w", err)
	}
	return group, nil
}

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Group   sarama.ConsumerGroup
	Handler *TransactionConsumer
	Config  *config.Config
}

func Run(p Params) {
	ctx, cancel := context.WithCancel(context.Background())
	topics := []string{p.Config.Kafka.TransactionTopic}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				for {
					if err := p.Group.Consume(ctx, topics, p.Handler); err != nil {
						if errors.Is(err, sarama.ErrClosedConsumerGroup) || ctx.Err() != nil {
							return
						}
						log.Printf("kafka: consumer group session error: %v", err)
					}
					if ctx.Err() != nil {
						return
					}
				}
			}()

			go func() {
				for err := range p.Group.Errors() {
					log.Printf("kafka: consumer error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return p.Group.Close()
		},
	})
}
