package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/JIeeiroSst/upload-service/pkg/log"
)

type MongoDB struct {
	mongo *mongo.Client
}

// mongodb://localhost:27017
func ConnectMongoDB(dns string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dns))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Error(err.Error())
		}
	}()

	return &MongoDB{
		mongo: client,
	}, nil
}
