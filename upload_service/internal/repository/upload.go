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
	collection *mongo.Collection
}

func NewUploadRepo(Collection *mongo.Collection) *UploadRepo {
	return &UploadRepo{
		collection: Collection,
	}
}

func (r *UploadRepo) Create(ctx context.Context, upload model.CreateMedia) error {
	_, err := r.collection.InsertOne(ctx, upload)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *UploadRepo) Update(ctx context.Context, id string, upload model.UpdateMedia) error {
	filter := bson.D{{Key: "_id", Value: id}}
	_, err := r.collection.UpdateOne(context.TODO(), filter, upload)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (r *UploadRepo) GetAll(ctx context.Context) ([]model.Media, error) {
	var result []model.Media
	cur, err := r.collection.Find(ctx, bson.D{})
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
	filter := bson.D{{Key: "_id", Value: id}}
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Info(result)
	return result, nil
}

func (r *UploadRepo) Delete(ctx context.Context, id string) error {
	filter := bson.D{{Key: "_id", Value: id}}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
