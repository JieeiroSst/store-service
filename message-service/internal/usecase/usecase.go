package usecase

import (
	"github.com/JIeeiroSst/message-service/internal/repository"
	"github.com/JIeeiroSst/message-service/pkg/kafka"
)

type Usecase struct {
	PubSub
}

type Dependency struct {
	Repos      *repository.Repositories
	QueueKakfa kafka.QueueKakfa
}

func NewUsecase(deps Dependency) *Usecase {
	pubSubUsecase := NewPubSubUsecase(deps.QueueKakfa,deps.Repos.Tracks)

	return &Usecase{
		PubSub: pubSubUsecase,
	}
}
