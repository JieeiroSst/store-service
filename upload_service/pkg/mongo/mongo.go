package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	mongo *mongo.Client
}

// mongodb://localhost:27017
func Connect(dns string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dns))
	if err != nil {
		return nil, err
	}

	return &MongoDB{
		mongo: client,
	}, nil
}
