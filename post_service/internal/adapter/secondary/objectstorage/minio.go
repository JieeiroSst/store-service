package objectstorage

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/post-service/config"
	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

func NewObjectStorage(cfg *config.Config) (port.ObjectStorage, error) {
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to minio: %w", err)
	}
	return &minioStorage{client: client, bucketName: cfg.Minio.BucketName, endpoint: cfg.Minio.Endpoint}, nil
}

func (s *minioStorage) UploadFile(ctx context.Context, input port.UploadFileInput) (*port.UploadFileResult, error) {
	_, err := s.client.PutObject(ctx, s.bucketName, input.FileName, input.Reader, input.Size, minio.PutObjectOptions{
		ContentType: input.ContentType,
		UserMetadata: map[string]string{
			"x-amz-acl": "public-read",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("upload to minio: %w", err)
	}
	return &port.UploadFileResult{URL: s.makeFileURL(input.FileName)}, nil
}

func (s *minioStorage) RemoveObject(ctx context.Context, fileName string) error {
	return s.client.RemoveObject(ctx, s.bucketName, fileName, minio.RemoveObjectOptions{})
}

func (s *minioStorage) makeFileURL(fileName string) string {
	return fmt.Sprintf("https://%s/%s/%s", s.endpoint, s.bucketName, fileName)
}
