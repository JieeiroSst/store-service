package publisher

import (
	"context"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/rabbitmq"
)

type notificationPublisher struct {
	mq rabbitmq.RabbitMQ
}

func NewNotificationPublisher(mq rabbitmq.RabbitMQ) port.NotificationPublisher {
	return &notificationPublisher{mq: mq}
}

func (p *notificationPublisher) Publish(ctx context.Context, notification *model.Notification) error {
	return p.mq.PublishToQueue(notification)
}
