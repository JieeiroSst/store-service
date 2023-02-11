package usecase

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/JIeeiroSst/upload-service/internal/repository"
	"github.com/JIeeiroSst/upload-service/model"
	uploadAPI "github.com/JIeeiroSst/upload-service/pkg/api"
	"github.com/JIeeiroSst/upload-service/pkg/cache"
	"github.com/JIeeiroSst/upload-service/pkg/snowflake"
)

type Uploads interface {
	Create(ctx context.Context, f multipart.File, h *multipart.FileHeader, ReceiverId string) error
	Update(ctx context.Context, id string, f multipart.File, h *multipart.FileHeader) error
	GetAll(ctx context.Context) ([]model.Media, error)
	GetById(ctx context.Context, id string) (*model.Media, error)
	Delete(ctx context.Context, id string) error
}

type UploadUsecase struct {
	uploadRepo repository.Uploads
	snowflake  snowflake.SnowflakeData
	uploadApi  uploadAPI.UploadApi
	cache      cache.CacheHelper
}

func NewUploadUsecase(uploadRepo repository.Uploads,
	snowflake snowflake.SnowflakeData, uploadApi uploadAPI.UploadApi,
	cache cache.CacheHelper) *UploadUsecase {
	return &UploadUsecase{
		uploadRepo: uploadRepo,
		snowflake:  snowflake,
		uploadApi:  uploadApi,
		cache:      cache,
	}
}

func (u *UploadUsecase) Create(ctx context.Context, f multipart.File, h *multipart.FileHeader, ReceiverId string) error {
	response, err := u.uploadApi.UploadFile(f, h)
	if err != nil {
		return err
	}
	upload := model.CreateMedia{
		Id:         u.snowflake.GearedID(),
		FileName:   response.Data.Title,
		URL:        response.Data.DisplayUrl,
		ReceiverId: ReceiverId,
		CreateDate: time.Now(),
	}
	if err := u.uploadRepo.Create(ctx, upload); err != nil {
		return err
	}
	return nil
}

func (u *UploadUsecase) Update(ctx context.Context, id string, f multipart.File, h *multipart.FileHeader) error {
	response, err := u.uploadApi.UploadFile(f, h)
	if err != nil {
		return err
	}
	upload := model.UpdateMedia{
		FileName:   response.Data.Title,
		URL:        response.Data.DisplayUrl,
		UpdateDate: time.Now(),
	}
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
