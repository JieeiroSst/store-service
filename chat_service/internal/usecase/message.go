package usecase

import (
	"github.com/JIeeiroSst/chat-service/internal/repository"
	"github.com/JIeeiroSst/chat-service/pkg/cache"
)

type Messages interface {
}

type Messagesecase struct {
	MessageRepo repository.MessageRepo
	CacheHelper cache.CacheHelper
}

func NewMessageUsecase(MessageRepo repository.MessageRepo, CacheHelper cache.CacheHelper) *Messagesecase {
	return &Messagesecase{
		MessageRepo: MessageRepo,
		CacheHelper: CacheHelper,
	}
}

