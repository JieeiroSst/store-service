package domain

import (
	"github.com/JIeeiroSst/cdn-service/internal/config"
	"github.com/JIeeiroSst/cdn-service/internal/ports"
	"go.uber.org/fx"
)

func newFileServiceOptions(cfg *config.Config) FileServiceOptions {
	return FileServiceOptions{
		MaxUploadSizeBytes: cfg.MinIO.MaxUploadSizeBytes,
		PresignPutExpiry:   cfg.MinIO.PresignExpiry(),
		PresignGetExpiry:   cfg.MinIO.PresignExpiry(),
		EdgeCacheBaseURL:   cfg.Edge.BaseURL,
	}
}

func newFileUsecase(repo ports.FileRepository, storage ports.ObjectStorage, opts FileServiceOptions) ports.FileUsecase {
	return NewFileService(repo, storage, opts)
}

var Module = fx.Options(
	fx.Provide(newFileServiceOptions),
	fx.Provide(newFileUsecase),
)
