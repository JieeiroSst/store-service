package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.elastic.co/apm"
)

type client struct {
	minioClient *minio.Client
	config      *Config
}

type Config struct {
	Endpoint        string
	SecretAccessKey string
	AccessKey       string
	BucketName      string
}

type ClientS3 interface {
	RemoveObject(ctx context.Context, fileName string) error
	UploadFile(ctx context.Context, args *UploadFileArgs) (*UploadObjectResponse, error)
}

func NewStorage(cfg *Config) ClientS3 {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil
	}

	return &client{
		minioClient: minioClient,
		config:      cfg,
	}
}

func (c *client) RemoveObject(ctx context.Context, fileName string) error {
	trx := apm.TransactionFromContext(ctx)
	minioServiceSpan := trx.StartSpan(fmt.Sprintf("Remove file in Minio server"), "External service.Minio", nil)
	defer minioServiceSpan.End()
	if err := c.minioClient.RemoveObject(ctx, c.config.BucketName, fileName, minio.RemoveObjectOptions{}); err != nil {
		return nil
	}
	return nil
}

func (c *client) makeFileURL(fileName string) string {
	return fmt.Sprintf("https://%v/%v/%v", c.config.Endpoint, c.config.BucketName, fileName)
}

func (c *client) UploadFile(ctx context.Context, args *UploadFileArgs) (*UploadObjectResponse, error) {
	var contentType string
	trx := apm.TransactionFromContext(ctx)
	minioServiceSpan := trx.StartSpan(fmt.Sprintf("Upload file to Minio server"), "External service.Minio", nil)
	defer minioServiceSpan.End()

	fileSize := args.FileHeader.Size

	contentTypes := args.FileHeader.Header["Content-Type"]
	if len(contentTypes) > 0 && contentTypes[0] != "" {
		for _, t := range contentTypes {
			contentType = t
		}
	}

	buffer, err := args.FileHeader.Open()
	if err != nil {
		return nil, err
	}

	// Upload file to CMC Cloud (Minio)
	_, err = c.minioClient.PutObject(ctx, c.config.BucketName, args.FileName, buffer, fileSize, minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: args.UserMetaData,
	})
	if err != nil {
		return nil, err
	}
	return &UploadObjectResponse{
		URL: c.makeFileURL(args.FileName),
	}, nil
}
