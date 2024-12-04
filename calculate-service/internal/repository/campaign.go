package repository

import (
	"context"

	"github.com/JIeeiroSst/calculate-service/common"
	"github.com/JIeeiroSst/calculate-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/pagination"
	"gorm.io/gorm"
)

type CampaignConfigs interface {
	CreateCampaignConfig(ctx context.Context, config *model.CampaignConfig) error
	FindByID(ctx context.Context, id string) (*model.CampaignConfig, error)
	FindPagination(ctx context.Context, p pagination.Pagination) (*pagination.Pagination, error)
	UpdateCampaignConfig(ctx context.Context, campaignConfig model.CampaignConfig) error
	FindByActive(ctx context.Context) (*model.CampaignConfig, error)
	SaveCampaignTypeConfig(ctx context.Context, typeConfig *model.CampaignTypeConfig) error
}

type CampaignConfigRepo struct {
	db *gorm.DB
}

func NewCampaignConfigRepo(db *gorm.DB) *CampaignConfigRepo {
	return &CampaignConfigRepo{
		db: db,
	}
}

func (r *CampaignConfigRepo) CreateCampaignConfig(ctx context.Context, config *model.CampaignConfig) error {
	if config != nil {
		if err := r.db.Create(&config).Error; err != nil {
			logger.Info(ctx, "CampaignConfigRepo config error %v", err)
			return err
		}
	}
	return nil
}

func (r *CampaignConfigRepo) FindByID(ctx context.Context, id string) (*model.CampaignConfig, error) {
	var result model.CampaignConfig
	if err := r.db.Where("id = ?", id).Find(&result).Error; err != nil {
		logger.Info(ctx, "FindByID err %v", err)
		return nil, err
	}

	return &result, nil
}

func (r *CampaignConfigRepo) FindPagination(ctx context.Context, param pagination.Pagination) (*pagination.Pagination, error) {
	var campaignConfig []model.CampaignConfig

	r.db.Scopes(pagination.Paginate(campaignConfig, &param, r.db)).Find(&campaignConfig)
	param.Rows = campaignConfig

	return &param, nil
}

func (r *CampaignConfigRepo) UpdateCampaignConfig(ctx context.Context, campaignConfig model.CampaignConfig) error {
	err := r.db.Model(model.CampaignConfig{}).Where("id = ? ", campaignConfig.ID).Updates(campaignConfig).Error
	if err != nil {
		logger.Info(ctx, "Update CampaignConfigRepo error %v", err)
		return err
	}
	return nil
}

func (r *CampaignConfigRepo) FindByActive(ctx context.Context) (*model.CampaignConfig, error) {
	var result model.CampaignConfig

	if err := r.db.Where("status = ?", common.Active.Value()).Find(&result).Error; err != nil {
		logger.Info(ctx, "FindByActive error %v", err)
		return nil, err
	}

	return &result, nil
}

func (r *CampaignConfigRepo) SaveCampaignTypeConfig(ctx context.Context, req *model.CampaignTypeConfig) error {
	var campaignTypeConfig model.CampaignTypeConfig
	if req == nil {
		return nil
	}
	if err := r.db.Where("id = ?", req.ID).Find(&campaignTypeConfig).Error; err != nil {
		logger.Error(ctx, "FindCampaignTypeConfig error %v", err)
	}

	if campaignTypeConfig.ID == "" {
		if err := r.db.Create(&req).Error; err != nil {
			logger.Info(ctx, "Update CampaignConfigRepo typeConfig error %v", err)
			return err
		}

		return nil
	}

	if err := r.db.Save(&req).Error; err != nil {
		logger.Info(ctx, "Update CampaignConfigRepo typeConfig error %v", err)
		return err
	}

	return nil
}
