package repository

import (
	"context"

	"github.com/JIeeiroSst/upload-service/model"
	"github.com/JIeeiroSst/upload-service/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Uploads interface {
	Create(ctx context.Context, upload model.CreateMedia) error
	Update(ctx context.Context, id string, upload model.UpdateMedia) error
	GetAll(ctx context.Context) ([]model.Media, error)
	GetById(ctx context.Context, id string) (*model.Media, error)
	Delete(ctx context.Context, id string) error
}

type UploadRepo struct {
	client *mongo.Client
}

func NewUploadRepo(client *mongo.Client) *UploadRepo {
	return &UploadRepo{
		client: client,
	}
}

func (r *UploadRepo) Create(ctx context.Context, upload model.CreateMedia) error {
	collection := r.client.Database("").Collection("items")
	_, err := collection.InsertOne(ctx, upload)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *UploadRepo) Update(ctx context.Context, id string, upload model.UpdateMedia) error {
	collection := r.client.Database("").Collection("items")
	filter := bson.D{{Key: "_id", Value: id}}
	_, err := collection.UpdateOne(context.TODO(), filter, upload)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (r *UploadRepo) GetAll(ctx context.Context) ([]model.Media, error) {
	collection := r.client.Database("").Collection("items")
	var result []model.Media
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		if err := cur.Decode(&result); err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}
	if err := cur.Err(); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return result, nil
}

func (r *UploadRepo) GetById(ctx context.Context, id string) (*model.Media, error) {
	var result *model.Media
	collection := r.client.Database("").Collection("items")
	filter := bson.D{{Key: "_id", Value: id}}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Info(result)
	return result, nil
}

func (r *UploadRepo) Delete(ctx context.Context, id string) error {
	filter := bson.D{{Key: "_id", Value: id}}
	collection := r.client.Database("").Collection("items")
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
