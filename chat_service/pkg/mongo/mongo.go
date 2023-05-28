package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Client *mongo.Client
}

func NewMongo(host string) *Mongo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dns := fmt.Sprintf("mongodb://%v", host)
	clientOptions := options.Client().ApplyURI(dns)
	client, _ := mongo.Connect(ctx, clientOptions)
	return &Mongo{
		Client: client,
	}
}
