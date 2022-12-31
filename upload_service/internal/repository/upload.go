package repository

import (
	"context"

	"github.com/JIeeiroSst/upload-service/model"
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
	client *mongo.Client
}

func NewUploadRepo(client *mongo.Client) *UploadRepo {
	return &UploadRepo{
		client: client,
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
	return nil, nil
}

func (r *UploadRepo) Delete(ctx context.Context, id string) error {
	return nil
}
