package usecase

import (
	"context"

	"github.com/JIeeiroSst/calculate-service/dto"
	"github.com/JIeeiroSst/calculate-service/internal/repository"
	"github.com/JIeeiroSst/utils/logger"
)

type Campaigns interface {
	RegisterCampaignTypeConfig(ctx context.Context, req dto.CampaignTypeConfig) error
}

type CampaignUsecase struct {
	Repo *repository.Repository
}

func NewCampaignUsecase(repo *repository.Repository) *CampaignUsecase {
	return &CampaignUsecase{
		Repo: repo,
	}
}

func (u *CampaignUsecase) RegisterCampaignTypeConfig(ctx context.Context, req dto.CampaignTypeConfig) error {
	model := req.BuildCreate()
	if err := u.Repo.CampaignConfigs.CreateCampaignTypeConfig(ctx, &model); err != nil {
		logger.Error(ctx, "CreateCampaignTypeConfig error %v", err)
		return err
	}
	return nil
}
