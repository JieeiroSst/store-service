package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
	"github.com/referral/service/pkg/logger"
)

var Module = fx.Options(
	fx.Provide(NewReferralService),
)

type referralService struct {
	linkRepo   ports.ReferralLinkRepository
	eventRepo  ports.ReferralEventRepository
	rewardRepo ports.RewardRepository
	statsRepo  ports.UserStatsRepository
	cfg        *config.Config
	log        *zap.Logger
}

type Params struct {
	fx.In

	LinkRepo   ports.ReferralLinkRepository
	EventRepo  ports.ReferralEventRepository
	RewardRepo ports.RewardRepository
	StatsRepo  ports.UserStatsRepository
	Config     *config.Config
	Logger     *zap.Logger
}

func NewReferralService(p Params) ports.ReferralService {
	return &referralService{
		linkRepo:   p.LinkRepo,
		eventRepo:  p.EventRepo,
		rewardRepo: p.RewardRepo,
		statsRepo:  p.StatsRepo,
		cfg:        p.Config,
		log:        p.Logger.Named("referral-service"),
	}
}

func (s *referralService) GenerateLink(ctx context.Context, req ports.GenerateLinkRequest) (*ports.GenerateLinkResponse, error) {
	refCode := uuid.New().String()
	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 0, s.cfg.Referral.TTLDays)
	deepLink := s.buildDeepLink(refCode, req.Platform)

	link := &domain.ReferralLink{
		RefCode:     refCode,
		CreatedAt:   now,
		OwnerUserID: req.OwnerUserID,
		Channel:     req.Channel,
		Status:      domain.StatusActive,
		ExpiresAt:   expiresAt,
		TTL:         expiresAt.Unix(),
		DeepLink:    deepLink,
		Platform:    req.Platform,
	}

	if err := s.linkRepo.Save(ctx, link); err != nil {
		return nil, fmt.Errorf("generate link: save: %w", err)
	}

	evt := &domain.ReferralEvent{
		RefCode:     refCode,
		EventID:     uuid.New().String(),
		EventType:   domain.EventLinkCopied,
		OccurredAt:  now,
		OwnerUserID: req.OwnerUserID,
		Platform:    req.Platform,
	}
	if err := s.eventRepo.Save(ctx, evt); err != nil {
		s.log.Warn("failed to save link_copied event",
			zap.Error(err),
			logger.RefCode(refCode),
		)
	}

	_ = s.statsRepo.IncrementCounters(ctx, req.OwnerUserID, 1, 0, 0, 0)

	s.log.Info("referral link generated",
		logger.RefCode(refCode),
		logger.OwnerUserID(req.OwnerUserID),
		logger.Channel(string(req.Channel)),
		logger.Platform(req.Platform),
		logger.DeepLink(deepLink),
		zap.Time("expires_at", expiresAt),
	)

	return &ports.GenerateLinkResponse{
		RefCode:   refCode,
		DeepLink:  deepLink,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *referralService) GetLink(ctx context.Context, refCode string) (*domain.ReferralLink, error) {
	link, err := s.linkRepo.FindByRefCode(ctx, refCode)
	if err != nil {
		return nil, fmt.Errorf("get link %s: %w", refCode, err)
	}
	return link, nil
}

func (s *referralService) ListUserLinks(ctx context.Context, ownerUserID string, limit int, cursor string) ([]*domain.ReferralLink, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	links, nextCursor, err := s.linkRepo.FindByOwnerUserID(ctx, ownerUserID, limit, cursor)
	if err != nil {
		return nil, "", fmt.Errorf("list links for user %s: %w", ownerUserID, err)
	}
	return links, nextCursor, nil
}

func (s *referralService) TrackEvent(ctx context.Context, req ports.TrackEventRequest) error {
	link, err := s.linkRepo.FindByRefCode(ctx, req.RefCode)
	if err != nil {
		return fmt.Errorf("track event: link not found %s: %w", req.RefCode, err)
	}

	if !link.IsActive() {
		if link.IsExpired() {
			return fmt.Errorf("track event: %w: link %s is expired", domain.ErrLinkNotActive, req.RefCode)
		}
		return fmt.Errorf("track event: %w: link %s is %s", domain.ErrLinkNotActive, req.RefCode, link.Status)
	}

	event := &domain.ReferralEvent{
		RefCode:     req.RefCode,
		EventID:     uuid.New().String(),
		EventType:   req.EventType,
		OccurredAt:  time.Now().UTC(),
		Platform:    req.Platform,
		NewUserID:   req.NewUserID,
		OwnerUserID: link.OwnerUserID,
		IPAddress:   req.IPAddress,
		DeviceID:    req.DeviceID,
		UserAgent:   req.UserAgent,
	}

	if err := s.eventRepo.Save(ctx, event); err != nil {
		return fmt.Errorf("track event: save: %w", err)
	}

	s.log.Info("referral event tracked",
		logger.RefCode(req.RefCode),
		logger.EventType(string(req.EventType)),
		logger.Platform(req.Platform),
		logger.NewUserID(req.NewUserID),
	)

	return nil
}

func (s *referralService) ConfirmInstall(ctx context.Context, req ports.ConfirmInstallRequest) (*ports.ConfirmInstallResponse, error) {
	link, err := s.linkRepo.FindByRefCode(ctx, req.RefCode)
	if err != nil {
		return nil, fmt.Errorf("confirm install: link not found: %w", err)
	}

	if req.NewUserID != "" && link.OwnerUserID == req.NewUserID {
		return nil, fmt.Errorf("confirm install: %w", domain.ErrSelfReferral)
	}

	if !link.IsActive() {
		reason := string(link.Status)
		if link.IsExpired() {
			reason = "expired"
		}
		s.log.Warn("confirm install: link not active",
			logger.RefCode(req.RefCode),
			zap.String("reason", reason),
		)
		return &ports.ConfirmInstallResponse{Attributed: false}, nil
	}

	now := time.Now().UTC()

	if err := s.linkRepo.UpdateStatus(ctx, req.RefCode, link.CreatedAt.Format(time.RFC3339), domain.StatusUsed); err != nil {
		if errors.Is(err, domain.ErrLinkNotActive) {
			// Concurrent ConfirmInstall already claimed this link
			return &ports.ConfirmInstallResponse{Attributed: false}, nil
		}
		return nil, fmt.Errorf("confirm install: update status: %w", err)
	}

	_ = s.eventRepo.Save(ctx, &domain.ReferralEvent{
		RefCode:     req.RefCode,
		EventID:     uuid.New().String(),
		EventType:   domain.EventAppInstalled,
		OccurredAt:  now,
		Platform:    req.Platform,
		NewUserID:   req.NewUserID,
		OwnerUserID: link.OwnerUserID,
		DeviceID:    req.DeviceID,
	})

	const rewardValue = 50000.0
	reward := &domain.ReferralReward{
		OwnerUserID: link.OwnerUserID,
		RefCode:     req.RefCode,
		NewUserID:   req.NewUserID,
		RewardType:  domain.RewardCredit,
		RewardValue: rewardValue,
		Status:      domain.RewardPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	rewardSaved := true
	if err := s.rewardRepo.Save(ctx, reward); err != nil {
		rewardSaved = false
		s.log.Error("failed to create reward record",
			zap.Error(err),
			logger.RefCode(req.RefCode),
			logger.OwnerUserID(link.OwnerUserID),
		)
	}

	rewardedDelta := int64(0)
	rewardAmtDelta := 0.0
	if rewardSaved {
		rewardedDelta = 1
		rewardAmtDelta = rewardValue
	}
	_ = s.statsRepo.IncrementCounters(ctx, link.OwnerUserID, 0, 1, rewardedDelta, rewardAmtDelta)

	s.log.Info("install confirmed — attribution successful",
		logger.RefCode(req.RefCode),
		logger.OwnerUserID(link.OwnerUserID),
		logger.NewUserID(req.NewUserID),
		logger.Platform(req.Platform),
		logger.RewardType(string(domain.RewardCredit)),
		logger.RewardValue(rewardValue),
	)

	return &ports.ConfirmInstallResponse{
		Attributed:  true,
		OwnerUserID: link.OwnerUserID,
		RewardType:  domain.RewardCredit,
	}, nil
}

func (s *referralService) GetUserStats(ctx context.Context, userID string) (*domain.UserReferralStats, error) {
	stats, err := s.statsRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get stats for %s: %w", userID, err)
	}
	return stats, nil
}

func (s *referralService) ActivateReferral(ctx context.Context, req ports.ActivateReferralRequest) (*ports.ActivateReferralResponse, error) {
	resp, err := s.ConfirmInstall(ctx, ports.ConfirmInstallRequest{
		RefCode:   req.RefCode,
		NewUserID: req.UserID,
		Platform:  req.Platform,
		DeviceID:  req.DeviceID,
	})
	if err != nil {
		return nil, err
	}
	return &ports.ActivateReferralResponse{
		Attributed:  resp.Attributed,
		OwnerUserID: resp.OwnerUserID,
		RewardType:  resp.RewardType,
	}, nil
}

func (s *referralService) GetReferralStatus(ctx context.Context, refCode string) (*ports.ReferralStatusResponse, error) {
	link, err := s.linkRepo.FindByRefCode(ctx, refCode)
	if err != nil {
		return nil, fmt.Errorf("get status %s: %w", refCode, err)
	}

	status := string(link.Status)
	if link.IsExpired() {
		status = "expired"
	}

	resp := &ports.ReferralStatusResponse{
		RefCode:     link.RefCode,
		Status:      status,
		OwnerUserID: link.OwnerUserID,
	}

	events, err := s.eventRepo.FindByRefCode(ctx, refCode)
	if err == nil {
		for _, evt := range events {
			if evt.EventType == domain.EventAppInstalled {
				t := evt.OccurredAt
				resp.ActivatedAt = &t
				resp.Platform = evt.Platform
				resp.NewUserID = evt.NewUserID
				break
			}
		}
	}

	return resp, nil
}

func (s *referralService) buildDeepLink(refCode, _ string) string {
	return fmt.Sprintf("%s/r/%s", s.cfg.DeepLink.BaseURL, refCode)
}
