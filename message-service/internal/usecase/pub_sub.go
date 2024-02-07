package usecase

import (
	"context"
	"sync"

	"github.com/JIeeiroSst/message-service/internal/builder"
	"github.com/JIeeiroSst/message-service/internal/repository"
	kafkaPkg "github.com/JIeeiroSst/message-service/pkg/kafka"
	"github.com/JIeeiroSst/message-service/pkg/logger"
)

type PubSub interface {
	Producer(ctx context.Context, topic string, data []byte)
	Consume(ctx context.Context, topic string) ([]byte, error)
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
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		u.queueKakfa.Producer(ctx, topic, data)
	}()

	go func() {
		defer wg.Done()
		model := builder.BuildTrackModel(data, "pub", topic)
		if err := u.trackRepo.TrackProducer(ctx, model); err != nil {
			logger.Logger().Error(err.Error())
		}
	}()

	wg.Wait()
}
func (u *PubSubUsecase) Consume(ctx context.Context, topic string) ([]byte, error) {
	message, err := u.queueKakfa.Consume(ctx, topic)
	if err != nil {
		return nil, err
	}

	model := builder.BuildTrackModel(message.Value, "sub", topic)
	if err := u.trackRepo.TrackProducer(ctx, model); err != nil {
		logger.Logger().Error(err.Error())
	}

	return message.Value, nil
}
