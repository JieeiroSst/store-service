package usecase

import (
	"context"
	"mime/multipart"

	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/media-service/internal/proxy/cloudflare"
	"github.com/JIeeiroSst/media-service/internal/repository"
	"github.com/JIeeiroSst/media-service/internal/usecase/build"
	"github.com/JIeeiroSst/media-service/model"
	"github.com/JIeeiroSst/media-service/utils"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/jinzhu/copier"
)

type Videos interface {
	UploadVideo(ctx context.Context, fileHeader *multipart.FileHeader) (string, error)
	SearchVideo(ctx context.Context, req dto.SearchVideoRequest) (*dto.SearchVideo, error)
	FindVideoByIDES(ctx context.Context, videoID int) (*dto.Video, error)
	SaveVideo(ctx context.Context, req dto.UploadVideoRequest) error
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

func (u *VideoUsecase) SaveVideo(ctx context.Context, req dto.UploadVideoRequest) error {
	var (
		video model.Video
		tag   model.Tag
	)

	if req.Video.VideoID < 0 {
		req.Video.VideoID = geared_id.GearedIntID()
	}

	if req.Tag.TagID < 0 {
		req.Tag.TagID = geared_id.GearedIntID()
	}

	req.Video.TagID = req.Tag.TagID

	if err := copier.Copy(&video, &req.Video); err != nil {
		return err
	}
	if err := copier.Copy(&tag, &req.Tag); err != nil {
		return err
	}

	if err := u.repo.Video.UploadVideo(ctx, video, tag); err != nil {
		return err
	}

	if err := u.repo.Video.InsertOrUpdateVideo(ctx, video); err != nil {
		return err
	}

	if err := u.repo.Video.InsertOrUpdateTag(ctx, tag); err != nil {
		return err
	}

	return nil
}

func (u *VideoUsecase) SearchVideo(ctx context.Context, req dto.SearchVideoRequest) (*dto.SearchVideo, error) {
	req = req.Build()
	videos, err := u.repo.Video.SearchVideo(ctx, req.Query, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	return build.BuildSearchVideo(videos), nil
}

func (u *VideoUsecase) FindVideoByIDES(ctx context.Context, videoID int) (*dto.Video, error) {
	video, err := u.repo.Video.FindVideoByIDES(ctx, videoID)
	if err != nil {
		return nil, err
	}

	return build.BuildVideo(video), nil
}

func (u *VideoUsecase) UploadVideo(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	buffer, err := utils.FileHeaderToBytesBuffer(fileHeader)
	if err != nil {
		return "", err
	}

	streamURL, err := u.cloudflare.UploadVideo(ctx, buffer)
	if err != nil {
		return "", err
	}
	return streamURL, nil
}
