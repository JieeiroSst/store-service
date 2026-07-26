package application

import (
	"context"

	"github.com/JIeeiroSst/admanagement-service/common"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
	"github.com/JIeeiroSst/admanagement-service/internal/domain/port"
)

type adCampaignService struct {
	repo port.AdCampaignRepository
}

func NewAdCampaignService(repo port.AdCampaignRepository) port.AdCampaignUsecase {
	return &adCampaignService{repo: repo}
}

func (s *adCampaignService) Create(ctx context.Context, campaign *model.AdCampaign) error {
	if campaign.Status == "" {
		campaign.Status = model.CampaignStatusDraft
	}
	return s.repo.Create(ctx, campaign)
}

func (s *adCampaignService) Get(ctx context.Context, id uint) (*model.AdCampaign, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *adCampaignService) Update(ctx context.Context, campaign *model.AdCampaign) error {
	return s.repo.Update(ctx, campaign)
}

func (s *adCampaignService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *adCampaignService) List(ctx context.Context) ([]model.AdCampaign, error) {
	return s.repo.List(ctx)
}

// ChangeStatus enforces the campaign status lifecycle (draft -> active ->
// paused/completed/cancelled) instead of allowing an arbitrary overwrite.
func (s *adCampaignService) ChangeStatus(ctx context.Context, id uint, status model.CampaignStatus) (*model.AdCampaign, error) {
	campaign, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !campaign.Status.CanTransitionTo(status) {
		return nil, common.ErrInvalidStatus
	}

	campaign.Status = status
	if err := s.repo.Update(ctx, campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}
