package usecase

import (
	"context"

	"github.com/JIeeiroSst/message-service/internal/repository"
	kafkaPkg "github.com/JIeeiroSst/message-service/pkg/kafka"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type PubSub interface {
	Producer(ctx context.Context, topic string, data []byte)
	Consume(ctx context.Context, topic string) (*kafka.Message, error)
}

type PubSubUsecase struct {
	queueKakfa kafkaPkg.QueueKakfa
	trackRepo  repository.Tracks
}

func NewPubSubUsecase(queueKakfa kafkaPkg.QueueKakfa,
	trackRepo repository.Tracks) *PubSubUsecase {
	return &PubSubUsecase{
		queueKakfa: queueKakfa,
		trackRepo:  trackRepo,
	}
}

func (u *PubSubUsecase) Producer(ctx context.Context, topic string, data []byte) {

}
func (u *PubSubUsecase) Consume(ctx context.Context, topic string) (*kafka.Message, error) {
	return nil, nil
}
