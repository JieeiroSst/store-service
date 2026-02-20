package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/qr-service/internal/domain/entity"
	"github.com/JIeeiroSst/qr-service/internal/domain/port"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const scanHistoryCollection = "scan_histories"

type mongoScanHistoryRepository struct {
	col    *mongo.Collection
	logger *zap.Logger
}

func NewMongoScanHistoryRepository(db *mongo.Database, logger *zap.Logger) port.ScanHistoryRepository {
	col := db.Collection(scanHistoryCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "qr_code_id", Value: 1}}},
		{Keys: bson.D{{Key: "short_code", Value: 1}}},
		{Keys: bson.D{{Key: "scanned_at", Value: -1}}},
	}

	if _, err := col.Indexes().CreateMany(ctx, indexes); err != nil {
		logger.Warn("failed to create scan_histories indexes", zap.Error(err))
	}

	return &mongoScanHistoryRepository{col: col, logger: logger}
}

func (r *mongoScanHistoryRepository) Create(ctx context.Context, history *entity.ScanHistory) error {
	_, err := r.col.InsertOne(ctx, history)
	return err
}

func (r *mongoScanHistoryRepository) GetByQRCodeID(ctx context.Context, qrCodeID string, page, limit int64) ([]*entity.ScanHistory, int64, error) {
	filter := bson.M{"qr_code_id": qrCodeID}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "scanned_at", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []*entity.ScanHistory
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *mongoScanHistoryRepository) GetStats(ctx context.Context, qrCodeID string) (*port.ScanStats, error) {
	filter := bson.M{"qr_code_id": qrCodeID}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("count total scans: %w", err)
	}

	// Count unique IPs
	uniqueIPs, err := r.col.Distinct(ctx, "ip_address", filter)
	if err != nil {
		return nil, fmt.Errorf("count unique ips: %w", err)
	}

	// Device breakdown aggregation
	devicePipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{"_id": "$device_type", "count": bson.M{"$sum": 1}}}},
	}
	deviceBreakdown, err := r.aggregateBreakdown(ctx, devicePipeline)
	if err != nil {
		return nil, err
	}

	// OS breakdown aggregation
	osPipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{"_id": "$os", "count": bson.M{"$sum": 1}}}},
	}
	osBreakdown, err := r.aggregateBreakdown(ctx, osPipeline)
	if err != nil {
		return nil, err
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	dailyPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"qr_code_id": qrCodeID,
			"scanned_at": bson.M{"$gte": thirtyDaysAgo},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$scanned_at",
				},
			},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := r.col.Aggregate(ctx, dailyPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dailyScans []port.DailyScan
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			dailyScans = append(dailyScans, port.DailyScan{
				Date:  result.ID,
				Count: result.Count,
			})
		}
	}

	return &port.ScanStats{
		TotalScans:      total,
		UniqueIPs:       int64(len(uniqueIPs)),
		DeviceBreakdown: deviceBreakdown,
		OSBreakdown:     osBreakdown,
		DailyScans:      dailyScans,
	}, nil
}

func (r *mongoScanHistoryRepository) DeleteByQRCodeID(ctx context.Context, qrCodeID string) error {
	_, err := r.col.DeleteMany(ctx, bson.M{"qr_code_id": qrCodeID})
	return err
}

func (r *mongoScanHistoryRepository) aggregateBreakdown(ctx context.Context, pipeline mongo.Pipeline) (map[string]int64, error) {
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	breakdown := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			key := result.ID
			if key == "" {
				key = "unknown"
			}
			breakdown[key] = result.Count
		}
	}
	return breakdown, nil
}
