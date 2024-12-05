package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/calculate-service/dto"
	"github.com/JIeeiroSst/calculate-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/pagination"
	"github.com/redis/go-redis/v9"
)

var (
	CampaignConfigByActiveKey = "campaign_config_by_active_key"
)

type Campaigns interface {
	RegisterCampaignTypeConfig(ctx context.Context, req dto.CampaignTypeConfig) error
	CreateCampaignConfig(ctx context.Context, req dto.CreateCampaignConfigRequest) error
	UpdateCampaignConfig(ctx context.Context, req dto.UpdateCampaignConfigRequest) error
	FindCampaignConfigByID(ctx context.Context, id string) (*dto.CampaignConfig, error)
	FindCampaignConfigByActive(ctx context.Context) (*dto.CampaignConfig, error)
	FindPagination(ctx context.Context, p pagination.Pagination) (*pagination.Pagination, error)
}

type CampaignUsecase struct {
	Repo  *repository.Repository
	Redis expire.CacheHelper
}

func NewCampaignUsecase(repo *repository.Repository, redis expire.CacheHelper) *CampaignUsecase {
	return &CampaignUsecase{
		Repo:  repo,
		Redis: redis,
	}
}

func (u *CampaignUsecase) RegisterCampaignTypeConfig(ctx context.Context, req dto.CampaignTypeConfig) error {
	model := req.BuildCreate()
	if err := u.Repo.CampaignConfigs.SaveCampaignTypeConfig(ctx, &model); err != nil {
		logger.Error(ctx, "RegisterCampaignTypeConfig error %v", err)
		return err
	}
	return nil
}

func (u *CampaignUsecase) CreateCampaignConfig(ctx context.Context, req dto.CreateCampaignConfigRequest) error {
	model := req.Build()
	if err := u.Repo.CampaignConfigs.CreateCampaignConfig(ctx, &model); err != nil {
		logger.Error(ctx, "CreateCampaignConfig error %v", err)
		return err
	}
	return nil
}

func (u *CampaignUsecase) UpdateCampaignConfig(ctx context.Context, req dto.UpdateCampaignConfigRequest) error {
	model := req.Build()

	if err := u.Repo.CampaignConfigs.UpdateCampaignConfig(ctx, model); err != nil {
		logger.Error(ctx, "UpdateCampaignConfig error %v", err)
		return err
	}

	if !model.DeletedAt.IsZero() {
		key := CampaignConfigByActiveKey
		u.Redis.Removekey(ctx, key)
	}
	return nil
}

func (u *CampaignUsecase) FindCampaignConfigByID(ctx context.Context, id string) (*dto.CampaignConfig, error) {
	campaignConfig, err := u.Repo.CampaignConfigs.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	campaignConfigDto := dto.BuildCampaignConfig(*campaignConfig)

	return campaignConfigDto, nil
}

func (u *CampaignUsecase) FindCampaignConfigByActive(ctx context.Context) (*dto.CampaignConfig, error) {
	campaignConfig, err := u.getCampaignConfigByActive(ctx)
	if err != nil {
		logger.Error(ctx, "FindCampaignConfigByActive err %v", err)
		return nil, err
	}

	return campaignConfig, nil
}

func (u *CampaignUsecase) setCampaignConfigByActive(ctx context.Context, campaignConfig *dto.CampaignConfig) error {
	if campaignConfig == nil {
		return errors.New("not found")
	}

	key := CampaignConfigByActiveKey

	if err := u.Redis.SetInterface(ctx, key, campaignConfig, time.Hour*24); err != nil {
		logger.Error(ctx, "SetCampaignConfigByActive err %v", err)
		return err
	}

	return nil
}

func (u *CampaignUsecase) getCampaignConfigByActive(ctx context.Context) (*dto.CampaignConfig, error) {
	var (
		campaignConfig *dto.CampaignConfig
	)
	key := CampaignConfigByActiveKey

	data, err := u.Redis.GetInterface(ctx, key)
	if err != nil {
		campaignConfigDB, errDB := u.Repo.CampaignConfigs.FindByActive(ctx)
		if errDB != nil {
			return nil, err
		}
		campaignConfig = dto.BuildCampaignConfig(*campaignConfigDB)
		if err == redis.Nil {
			if err := u.setCampaignConfigByActive(ctx, campaignConfig); err != nil {
				logger.Error(ctx, "SetCampaignConfigByActive err %v", err)
				return nil, err
			}
		}
	} else {
		campaignConfig = data.(*dto.CampaignConfig)
	}
	return campaignConfig, nil
}

func (u *CampaignUsecase) FindPagination(ctx context.Context, p pagination.Pagination) (*pagination.Pagination, error) {
	pagination, err := u.Repo.CampaignConfigs.FindPagination(ctx, p)
	if err != nil {
		logger.Error(ctx, "FindPagination error %v", err)
		return nil, err
	}
	return pagination, nil
}
