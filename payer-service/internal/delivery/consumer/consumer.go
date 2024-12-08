package consumer

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/payer-service/dto"
	"github.com/JIeeiroSst/payer-service/internal/usecase"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	usecase *usecase.Usecase
	nats    *nats.Conn
}

func NewConsumer(usecase *usecase.Usecase, nats *nats.Conn) *Consumer {
	return &Consumer{
		usecase: usecase,
		nats:    nats,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.nats.Subscribe("payer.transaction", func(msg *nats.Msg) {
		var createTransactionsRequest dto.CreateTransactionsRequest
		if err := json.Unmarshal(msg.Data, &createTransactionsRequest); err != nil {
			logger.Error(ctx, "error %v", err)
		}

		if err := c.usecase.Transactions.Transactions(ctx, createTransactionsRequest); err != nil {
			logger.Error(ctx, "error %v", err)
		}
	})
}
