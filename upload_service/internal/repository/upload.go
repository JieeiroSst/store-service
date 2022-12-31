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
	Update(ctx context.Context, upload model.UpdateMedia) error
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

	return nil
}

func (r *UploadRepo) Update(ctx context.Context, upload model.UpdateMedia) error {

	return nil
}

func (r *UploadRepo) GetAll(ctx context.Context) ([]model.Media, error) {
	return nil, nil
}

func (r *UploadRepo) GetById(ctx context.Context, id string) (*model.Media, error) {
	var result *model.Media
	filter := bson.D{{Key: "id", Value: id}}
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Info(result)
	return result, nil
}

func (r *UploadRepo) Delete(ctx context.Context, id string) error {
	return nil
}
