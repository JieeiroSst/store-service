package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/chat-service/common"
	"github.com/JIeeiroSst/chat-service/dto"
	"github.com/JIeeiroSst/chat-service/internal/repository"
	"github.com/JIeeiroSst/chat-service/model"
	"github.com/JIeeiroSst/chat-service/pkg/cache"
	"github.com/JIeeiroSst/chat-service/pkg/snowflake"
	"github.com/redis/go-redis/v9"
)

type Messages interface {
	SaveMessage(ctx context.Context, message dto.Messages) error
	GetMessageById(ctx context.Context, id int) (*dto.Messages, error)
	CreateReport(ctx context.Context, report dto.Reports) error
	GetReportByUser(ctx context.Context, userId int) ([]dto.Reports, error)
	DeleteMessage(ctx context.Context, messageId, userId int) error
}

type Messagesecase struct {
	MessageRepo repository.Messages
	CacheHelper cache.CacheHelper
	Snowflake   snowflake.SnowflakeData
}

func NewMessageUsecase(MessageRepo repository.Messages, cacheHelper cache.CacheHelper,
	snowflake snowflake.SnowflakeData) *Messagesecase {
	return &Messagesecase{
		MessageRepo: MessageRepo,
		CacheHelper: cacheHelper,
		Snowflake:   snowflake,
	}
}

func (u *Messagesecase) SaveMessage(ctx context.Context, message dto.Messages) error {
	messageType, err := model.ParseMessageType(message.MessageType)
	if err != nil {
		return err
	}
	mesageModel := model.Messages{
		ID:          u.Snowflake.GearedID(),
		SenderId:    message.SenderId,
		MessageType: messageType,
		CreatedAt:   time.Now(),
	}
	if err := u.MessageRepo.SaveMessage(ctx, mesageModel); err != nil {
		return err
	}
	return nil
}

func (u *Messagesecase) GetMessageById(ctx context.Context, id int) (*dto.Messages, error) {
	var (
		message *model.Messages
		errDB   error
	)
	valueIntrface, err := u.CacheHelper.GetInterface(context.Background(), fmt.Sprintf(common.ListMessageBySenderID, id), message)
	if err != nil {
		message, errDB = u.MessageRepo.GetMessageById(ctx, id)
		if errDB != nil {
			return nil, err
		}
		if err == redis.Nil {
			_ = u.CacheHelper.Set(context.Background(), fmt.Sprintf(common.ListMessageBySenderID, id), message, time.Second*60)
		}
	} else {
		message = valueIntrface.(*model.Messages)
	}

	return &dto.Messages{
		ID:          message.ID,
		SenderId:    message.SenderId,
		MessageType: message.MessageType.String(),
		CreatedAt:   message.CreatedAt,
		DeletedAt:   message.DeletedAt,
	}, nil
}

func (u *Messagesecase) CreateReport(ctx context.Context, report dto.Reports) error {
	status, err := model.ParseStatus(report.Status)
	if err != nil {
		return err
	}
	reportModel := model.Reports{
		ID:        u.Snowflake.GearedID(),
		UserId:    report.UserId,
		Notes:     report.Notes,
		Status:    status,
		CreatedAt: time.Now(),
	}
	if err := u.MessageRepo.CreateReport(ctx, reportModel); err != nil {
		return err
	}
	return nil
}

func (u *Messagesecase) GetReportByUser(ctx context.Context, userId int) ([]dto.Reports, error) {
	var (
		reportsModel []model.Reports
		reports      []dto.Reports
		errDB        error
	)
	valueInterface, err := u.CacheHelper.GetInterface(context.Background(), fmt.Sprintf(common.ListReportByUser, userId), reportsModel)
	if err != nil {
		reportsModel, errDB = u.MessageRepo.GetReportByUser(ctx, userId)
		if errDB != nil {
			return nil, err
		}
		if err == redis.Nil {
			_ = u.CacheHelper.Set(context.Background(), fmt.Sprintf(common.ListReportByUser, userId), reportsModel, time.Second*60)
		}
	} else {
		reportsModel = valueInterface.([]model.Reports)
	}
	for _, value := range reportsModel {
		reports = append(reports, dto.Reports{
			ID:         value.ID,
			UserId:     value.UserId,
			ReportType: value.ReportType,
			Notes:      value.Notes,
			Status:     value.Status.String(),
			CreatedAt:  value.CreatedAt,
		})
	}
	return reports, nil
}

func (u *Messagesecase) DeleteMessage(ctx context.Context, messageId, userId int) error {
	message := model.DeletedMessages{
		ID:         u.Snowflake.GearedID(),
		MessagesId: messageId,
		UserId:     userId,
		CreatedAt:  time.Now(),
		DeletedAt:  time.Now(),
	}
	if err := u.MessageRepo.DeleteMessage(ctx, message); err != nil {
		return err
	}
	return nil
}
