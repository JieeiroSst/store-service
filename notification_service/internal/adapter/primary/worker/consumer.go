package worker

import (
	"context"
	"log"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/rabbitmq"
	"go.uber.org/fx"
)

type Consumer struct {
	mq           rabbitmq.RabbitMQ
	notification port.NotificationUsecase
}

func NewConsumer(mq rabbitmq.RabbitMQ, notification port.NotificationUsecase) *Consumer {
	return &Consumer{mq: mq, notification: notification}
}

func (c *Consumer) handle(notification model.Notification) error {
	return c.notification.Dispatch(context.Background(), &notification)
}

func registerLifecycle(lc fx.Lifecycle, c *Consumer) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := c.mq.StartConsumer(c.handle); err != nil {
					log.Printf("rabbitmq consumer error: %v", err)
				}
			}()
			return nil
		},
	})
}

var Module = fx.Options(
	fx.Provide(NewConsumer),
	fx.Invoke(registerLifecycle),
)
