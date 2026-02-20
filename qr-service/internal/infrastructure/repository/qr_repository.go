package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/qr-service/internal/domain/entity"
	"github.com/JIeeiroSst/qr-service/internal/domain/port"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const qrCodesCollection = "qr_codes"

type mongoQRCodeRepository struct {
	col    *mongo.Collection
	logger *zap.Logger
}

func NewMongoQRCodeRepository(db *mongo.Database, logger *zap.Logger) port.QRCodeRepository {
	col := db.Collection(qrCodesCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "short_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "created_by", Value: 1}}},
	}

	if _, err := col.Indexes().CreateMany(ctx, indexes); err != nil {
		logger.Warn("failed to create qr_codes indexes", zap.Error(err))
	}

	return &mongoQRCodeRepository{col: col, logger: logger}
}

func (r *mongoQRCodeRepository) Create(ctx context.Context, qr *entity.QRCode) (*entity.QRCode, error) {
	_, err := r.col.InsertOne(ctx, qr)
	if err != nil {
		return nil, fmt.Errorf("insert qr code: %w", err)
	}
	return qr, nil
}

func (r *mongoQRCodeRepository) GetByID(ctx context.Context, id string) (*entity.QRCode, error) {
	var qr entity.QRCode
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&qr)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("qr code not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

func (r *mongoQRCodeRepository) GetByShortCode(ctx context.Context, shortCode string) (*entity.QRCode, error) {
	var qr entity.QRCode
	err := r.col.FindOne(ctx, bson.M{"short_code": shortCode}).Decode(&qr)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("qr code not found for short_code: %s", shortCode)
	}
	if err != nil {
		return nil, err
	}
	return &qr, nil
}

func (r *mongoQRCodeRepository) List(ctx context.Context, filter port.QRCodeFilter) ([]*entity.QRCode, int64, error) {
	query := bson.M{}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.Type != "" {
		query["type"] = filter.Type
	}
	if filter.CreatedBy != "" {
		query["created_by"] = filter.CreatedBy
	}
	if filter.Search != "" {
		query["$or"] = bson.A{
			bson.M{"title": bson.M{"$regex": filter.Search, "$options": "i"}},
			bson.M{"content": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	total, err := r.col.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skip := (filter.Page - 1) * filter.Limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(filter.Limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.col.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []*entity.QRCode
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *mongoQRCodeRepository) Update(ctx context.Context, id string, qr *entity.QRCode) (*entity.QRCode, error) {
	qr.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": id}, qr)
	if err != nil {
		return nil, fmt.Errorf("update qr code: %w", err)
	}
	return qr, nil
}

func (r *mongoQRCodeRepository) UpdateContent(ctx context.Context, id string, content string, redirectURL string) (*entity.QRCode, error) {
	update := bson.M{
		"$set": bson.M{
			"content":      content,
			"redirect_url": redirectURL,
			"updated_at":   time.Now(),
		},
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, fmt.Errorf("update content: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *mongoQRCodeRepository) Delete(ctx context.Context, id string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *mongoQRCodeRepository) IncrementScanCount(ctx context.Context, id string) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"scan_count": 1}},
	)
	return err
}
