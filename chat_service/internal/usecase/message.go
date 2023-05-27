package usecase

import (
	"context"

	"github.com/JIeeiroSst/chat-service/dto"
	"github.com/JIeeiroSst/chat-service/internal/repository"
	"github.com/JIeeiroSst/chat-service/pkg/cache"
	"github.com/JIeeiroSst/chat-service/pkg/snowflake"
)

type Messages interface {
	SaveMessage(ctx context.Context, message dto.Messages) error
	GetMessageById(ctx context.Context, id int) (*dto.Messages, error)
	CreateReport(ctx context.Context, report dto.Reports) error
	GetReportByUser(ctx context.Context, userId int) ([]dto.Reports, error)
	DeleteMessage(ctx context.Context, messageId, userId int) error
}

type Messagesecase struct {
	MessageRepo repository.MessageRepo
	CacheHelper cache.CacheHelper
	Snowflake   snowflake.SnowflakeData
}

func NewMessageUsecase(MessageRepo repository.MessageRepo, cacheHelper cache.CacheHelper,
	snowflake snowflake.SnowflakeData) *Messagesecase {
	return &Messagesecase{
		MessageRepo: MessageRepo,
		CacheHelper: cacheHelper,
		Snowflake:   snowflake,
	}
}

func (u *Messagesecase) SaveMessage(ctx context.Context, message dto.Messages) error {
	return nil
}

func (u *Messagesecase) GetMessageById(ctx context.Context, id int) (*dto.Messages, error) {
	return nil, nil
}

func (u *Messagesecase) CreateReport(ctx context.Context, report dto.Reports) error {
	return nil
}

func (u *Messagesecase) GetReportByUser(ctx context.Context, userId int) ([]dto.Reports, error) {
	return nil, nil
}

func (u *Messagesecase) DeleteMessage(ctx context.Context, messageId, userId int) error {
	return nil
}
