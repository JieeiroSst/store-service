package usecase

import (
	"context"

	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/media-service/internal/proxy/cloudflare"
	"github.com/JIeeiroSst/media-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Videos interface {
	UploadVideo(ctx context.Context, req dto.UploadVideoRequest) error
}

type VideoUsecase struct {
	repo       *repository.Repository
	cache      expire.CacheHelper
	cloudflare cloudflare.CloudflareProxy
}

func NewVideoUsecase(repo *repository.Repository,
	cache expire.CacheHelper,
	cloudflare cloudflare.CloudflareProxy) *VideoUsecase {
	return &VideoUsecase{
		repo:       repo,
		cache:      cache,
		cloudflare: cloudflare,
	}
}
