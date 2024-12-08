package consumer

import "github.com/JIeeiroSst/payer-service/internal/usecase"

type Consumer struct {
	usecase *usecase.Usecase
}

func NewConsumer(usecase *usecase.Usecase) *Consumer {
	return &Consumer{
		usecase: usecase,
	}
}
