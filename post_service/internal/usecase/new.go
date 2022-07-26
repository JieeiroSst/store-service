package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/post-service/internal/repository"
	"github.com/JIeeiroSst/post-service/model"
	"github.com/JIeeiroSst/post-service/pkg/minio"
	"github.com/JIeeiroSst/post-service/pkg/snowflake"
	"github.com/google/uuid"
)

type News interface {
	Create(cateID string, new model.New, upload minio.UploadFileArgs) error
	News() ([]model.New, error)
	NewById(id string) (*model.New, error)
	Update(id string, new model.New) error
	uploadFile(ctx context.Context, args *minio.UploadFileArgs) (*minio.UploadObjectResponse,
		error)
}

type NewsUsecase struct {
	NewRepo   repository.News
	Snowflake snowflake.SnowflakeData
	MediaRepo repository.Medias
	Minio     minio.ClientS3
}

func NewNewsUsecase(NewRepo repository.News,
	Snowflake snowflake.SnowflakeData, MediaRepo repository.Medias,
	Minio minio.ClientS3) *NewsUsecase {
	return &NewsUsecase{
		NewRepo:   NewRepo,
		Snowflake: Snowflake,
		MediaRepo: MediaRepo,
		Minio:     Minio,
	}
}

func (u *NewsUsecase) Create(cateID string, new model.New, upload minio.UploadFileArgs) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, err := u.uploadFile(ctx, &upload)
	if err != nil {
		return err
	}

	imageId := u.Snowflake.GearedID()

	media := model.Media{
		ID:          imageId,
		URL:         image.URL,
		Description: image.FileName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	newId := u.Snowflake.GearedID()

	newArg := model.New{
		Id:          newId,
		AuthorId:    new.AuthorId,
		Name:        new.Name,
		Content:     new.Content,
		Description: new.Description,
		MediaId:     new.MediaId,
	}

	argsCateNew := model.NewCategory{
		NewId:      newId,
		CategoryId: cateID,
	}

	if err := u.MediaRepo.SaveMedia(ctx, media); err != nil {
		return err
	}

	if err := u.NewRepo.Create(newArg, argsCateNew); err != nil {
		return err
	}

	return nil
}

func (u *NewsUsecase) News() ([]model.New, error) {
	news, err := u.NewRepo.News()
	if err != nil {
		return nil, err
	}
	return news, nil
}

func (u *NewsUsecase) NewById(id string) (*model.New, error) {
	new, err := u.NewRepo.NewById(id)
	if err != nil {
		return nil, err
	}
	return new, nil
}

func (u *NewsUsecase) Update(id string, new model.New) error {
	if err := u.NewRepo.Update(id, new); err != nil {
		return err
	}
	return nil
}

func (u *NewsUsecase) uploadFile(ctx context.Context, args *minio.UploadFileArgs) (*minio.UploadObjectResponse,
	error) {
	userMetaData := map[string]string{
		"x-amz-acl": "public-read",
	}
	var fileExtension string
	splitedArr := strings.Split(args.FileHeader.Filename, ".")
	if len(splitedArr) > 0 {
		fileExtension = splitedArr[len(splitedArr)-1]
	}
	uuid := uuid.New().String()
	uuidFileName := fmt.Sprintf("%v.%v", uuid, fileExtension)
	res, err := u.Minio.UploadFile(ctx, &minio.UploadFileArgs{
		UserMetaData: userMetaData,
		File:         args.File,
		FileHeader:   args.FileHeader,
		FileName:     uuidFileName,
	})
	if err != nil {
		return nil, err
	}
	return &minio.UploadObjectResponse{URL: res.URL, FileName: uuidFileName}, nil
}
