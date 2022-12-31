package usecase

import (
	"context"

	"github.com/JIeeiroSst/upload-service/internal/repository"
	"github.com/JIeeiroSst/upload-service/model"
	"github.com/JIeeiroSst/upload-service/pkg/snowflake"
)

type Uploads interface {
	Create(ctx context.Context, upload model.CreateMedia) error
	Update(ctx context.Context, id string, upload model.UpdateMedia) error
	GetAll(ctx context.Context) ([]model.Media, error)
	GetById(ctx context.Context, id string) (*model.Media, error)
	Delete(ctx context.Context, id string) error
}

type UploadUsecase struct {
	uploadRepo repository.Uploads
	snowflake  snowflake.SnowflakeData
}

func NewUploadUsecase(uploadRepo repository.Uploads,
	snowflake snowflake.SnowflakeData) *UploadUsecase {
	return &UploadUsecase{
		uploadRepo: uploadRepo,
		snowflake:  snowflake,
	}
}

func (u *UploadUsecase) Create(ctx context.Context, upload model.CreateMedia) error {
	if err := u.uploadRepo.Create(ctx, upload); err != nil {
		return err
	}
	return nil
}

func (u *UploadUsecase) Update(ctx context.Context, id string, upload model.UpdateMedia) error {
	if err := u.uploadRepo.Update(ctx, id, upload); err != nil {
		return err
	}
	return nil
}

func (u *UploadUsecase) GetAll(ctx context.Context) ([]model.Media, error) {
	uploads, err := u.uploadRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return uploads, nil
}

func (u *UploadUsecase) GetById(ctx context.Context, id string) (*model.Media, error) {
	upload, err := u.uploadRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return upload, nil
}

func (u *UploadUsecase) Delete(ctx context.Context, id string) error {
	if err := u.uploadRepo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}
