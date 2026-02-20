package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/qr-service/internal/infrastructure/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func NewMongoDatabase(cfg *config.Config, logger *zap.Logger) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect to mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	logger.Info("connected to MongoDB", zap.String("db", cfg.MongoDB.Database))

	return client.Database(cfg.MongoDB.Database), nil
}
