package consumer

import (
	"context"

	"github.com/JIeeiroSst/order-service/internal/usecase"
)

type consumer struct {
	Usecase usecase.Usecase
}

type ConsumerInterface interface {
	Start(ctx context.Context)
}

func NewConsumer(Usecase usecase.Usecase) ConsumerInterface {
	return &consumer{
		Usecase: Usecase,
	}
}

func (c *consumer) Start(ctx context.Context) {

}
