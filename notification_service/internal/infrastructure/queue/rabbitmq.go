package queue

import (
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/rabbitmq"
	"go.uber.org/fx"
)

func NewRabbitMQ(cfg *config.Config) (rabbitmq.RabbitMQ, error) {
	return rabbitmq.GetInstance(cfg.Rabbit)
}

var Module = fx.Options(
	fx.Provide(NewRabbitMQ),
)
