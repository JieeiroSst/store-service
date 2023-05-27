package repository

import (
	"context"

	"github.com/JIeeiroSst/chat-service/common"
	"github.com/JIeeiroSst/chat-service/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type Messages interface {
	SaveMessage(ctx context.Context, message model.Messages) error
	GetMessageById(ctx context.Context, id int) (*model.Messages, error)
	CreateReport(ctx context.Context, report model.Reports) error
	GetReportByUser(ctx context.Context, userId int) ([]model.Reports, error)
	DeleteMessage(ctx context.Context, messageId, userId int) error
}

type MessageRepo struct {
	db *mongo.Client
}

func NewMessageRepo(db *mongo.Client) *MessageRepo {
	return &MessageRepo{
		db: db,
	}
}

func (r *MessageRepo) SaveMessage(ctx context.Context, message model.Messages) error {
	_, err := r.db.Database(common.Database).Collection(common.Collection).InsertOne(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

func (r *MessageRepo) GetMessageById(ctx context.Context, id int) (*model.Messages, error) {
	var message model.Messages
	err := r.db.Database(common.Database).Collection(common.Collection).FindOne(ctx, model.Messages{ID: id}).Decode(&message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *MessageRepo) CreateReport(ctx context.Context, report model.Reports) error {
	_, err := r.db.Database(common.Database).Collection(common.Collection).InsertOne(ctx, report)
	if err != nil {
		return err
	}
	return nil
}

func (r *MessageRepo) GetReportByUser(ctx context.Context, userId int) ([]model.Reports, error) {
	var reports []model.Reports
	cursor, err := r.db.Database(common.Database).Collection(common.Collection).Find(ctx, model.Reports{UserId: userId})
	if err != nil {
		return nil, err

	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var report model.Reports
		cursor.Decode(&report)
		reports = append(reports, report)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *MessageRepo) DeleteMessage(ctx context.Context, messageId, userId int) error {
	message := model.DeletedMessages{
		MessagesId: messageId,
		UserId:     userId,
	}
	_, err := r.db.Database(common.Database).Collection(common.Collection).InsertOne(ctx, message)
	if err != nil {
		return err
	}
	return nil
}
