package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/referral/service/pkg/refcode"
	"go.uber.org/zap"

	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
	"github.com/referral/service/pkg/logger"
	"github.com/referral/service/pkg/middleware"
)

var Module = fx.Options(
	fx.Provide(NewReferralService),
)

type referralService struct {
	linkRepo        ports.ReferralLinkRepository
	eventRepo       ports.ReferralEventRepository
	rewardRepo      ports.RewardRepository
	statsRepo       ports.UserStatsRepository
	programRepo     ports.RewardProgramRepository
	attributionRepo ports.ReferralAttributionRepository
	cache           ports.Cache
	cfg             *config.Config
	log             *zap.Logger
}

type Params struct {
	fx.In

	LinkRepo        ports.ReferralLinkRepository
	EventRepo       ports.ReferralEventRepository
	RewardRepo      ports.RewardRepository
	StatsRepo       ports.UserStatsRepository
	ProgramRepo     ports.RewardProgramRepository
	AttributionRepo ports.ReferralAttributionRepository
	Cache           ports.Cache
	Config          *config.Config
	Logger          *zap.Logger
}

func NewReferralService(p Params) ports.ReferralService {
	return &referralService{
		linkRepo:        p.LinkRepo,
		eventRepo:       p.EventRepo,
		rewardRepo:      p.RewardRepo,
		statsRepo:       p.StatsRepo,
		programRepo:     p.ProgramRepo,
		attributionRepo: p.AttributionRepo,
		cache:           p.Cache,
		cfg:             p.Config,
		log:             p.Logger.Named("referral-service"),
	}
}

func (s *referralService) logWithTrace(ctx context.Context) *zap.Logger {
	if traceID := middleware.TraceIDFromContext(ctx); traceID != "" {
		return s.log.With(middleware.TraceID(traceID))
	}
	return s.log
}

func (s *referralService) GenerateLink(ctx context.Context, req ports.GenerateLinkRequest) (*ports.GenerateLinkResponse, error) {
	log := s.logWithTrace(ctx)

	const maxRetries = 3
	now := time.Now().UTC()

	startOfDay := now.Truncate(24 * time.Hour)
	dayTTL := startOfDay.Add(24 * time.Hour).Sub(now)
	dateStr := now.Format("2006-01-02")
	cacheKey := fmt.Sprintf("referral:daily_link_count:%s:%s", req.OwnerUserID, dateStr)

	count, err := s.getDailyLinkCount(ctx, cacheKey, req.OwnerUserID, startOfDay.UnixMilli(), dayTTL)
	if err != nil {
		log.Warn("generate link: failed to check daily rate limit",
			logger.OwnerUserID(req.OwnerUserID),
			zap.Error(err),
		)
	} else if count >= int64(s.cfg.Referral.MaxPerDay) {
		log.Warn("generate link: daily limit reached",
			logger.OwnerUserID(req.OwnerUserID),
			zap.Int64("count", count),
			zap.Int("max_per_day", s.cfg.Referral.MaxPerDay),
		)
		return nil, fmt.Errorf("generate link: %w", domain.ErrRateLimitExceeded)
	}

	expiresAt := now.AddDate(0, 0, s.cfg.Referral.TTLDays)

	var link *domain.ReferralLink
	var refCode, deepLink string
	for attempt := 1; attempt <= maxRetries; attempt++ {
		code, err := refcode.Generate()
		if err != nil {
			return nil, fmt.Errorf("generate link: generate ref_code: %w", err)
		}
		refCode = code
		deepLink = s.buildDeepLink(refCode, req.Platform)

		link = &domain.ReferralLink{
			RefCode:     refCode,
			CreatedAt:   now.UnixMilli(),
			OwnerUserID: req.OwnerUserID,
			Channel:     req.Channel,
			Status:      domain.StatusActive,
			ExpiresAt:   expiresAt.UnixMilli(),
			DeepLink:    deepLink,
			Platform:    req.Platform,
		}

		err = s.linkRepo.Save(ctx, link)
		if err == nil {
			break
		}
		if errors.Is(err, domain.ErrDuplicateRefCode) {
			log.Warn("generate link: ref_code collision, retrying",
				logger.OwnerUserID(req.OwnerUserID),
				zap.String("ref_code", refCode),
				zap.Int("attempt", attempt),
			)
			if attempt == maxRetries {
				return nil, fmt.Errorf("generate link: ref_code collision after %d attempts", maxRetries)
			}
			continue
		}
		log.Error("generate link: save failed",
			logger.OwnerUserID(req.OwnerUserID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("generate link: save: %w", err)
	}

	if _, err := s.cache.Incr(ctx, cacheKey, dayTTL); err != nil {
		log.Warn("generate link: failed to increment daily count cache",
			logger.OwnerUserID(req.OwnerUserID),
			zap.Error(err),
		)
	}

	if err := s.eventRepo.Save(ctx, &domain.ReferralEvent{
		RefCode:     refCode,
		EventID:     uuid.New().String(),
		EventType:   domain.EventLinkCreated,
		OccurredAt:  now.UnixMilli(),
		OwnerUserID: req.OwnerUserID,
		Platform:    req.Platform,
	}); err != nil {
		log.Warn("generate link: failed to save link_created event",
			logger.RefCode(refCode),
			zap.Error(err),
		)
	}

	log.Info("referral link generated",
		logger.RefCode(refCode),
		logger.OwnerUserID(req.OwnerUserID),
		logger.Channel(string(req.Channel)),
		logger.Platform(req.Platform),
		logger.DeepLink(deepLink),
		zap.Int64("expires_at_ms", expiresAt.UnixMilli()),
	)

	return &ports.GenerateLinkResponse{
		RefCode:   refCode,
		DeepLink:  deepLink,
		ExpiresAt: expiresAt.UnixMilli(),
	}, nil
}

func (s *referralService) GetLink(ctx context.Context, refCode string) (*domain.ReferralLink, error) {
	link, err := s.linkRepo.FindByRefCode(ctx, refCode)
	if err != nil {
		s.logWithTrace(ctx).Error("get link failed",
			logger.RefCode(refCode),
			zap.Error(err),
		)
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
		s.logWithTrace(ctx).Error("list user links failed",
			logger.OwnerUserID(ownerUserID),
			zap.Error(err),
		)
		return nil, "", fmt.Errorf("list links for user %s: %w", ownerUserID, err)
	}
	return links, nextCursor, nil
}

func (s *referralService) TrackEvent(ctx context.Context, req ports.TrackEventRequest) error {
	log := s.logWithTrace(ctx)

	link, err := s.linkRepo.FindByRefCode(ctx, req.RefCode)
	if err != nil {
		log.Error("track event: find link failed",
			logger.RefCode(req.RefCode),
			zap.Error(err),
		)
		return fmt.Errorf("track event: link not found %s: %w", req.RefCode, err)
	}

	if !link.IsActive() {
		if link.IsExpired() {
			return fmt.Errorf("track event: %w: link %s is expired", domain.ErrLinkNotActive, req.RefCode)
		}
		return fmt.Errorf("track event: %w: link %s is %s", domain.ErrLinkNotActive, req.RefCode, link.Status)
	}

	// For click events: check for a prior click from the same device/IP before saving
	// to avoid inflating total_invited on repeated clicks.
	alreadyCounted := false
	if req.EventType == domain.EventLinkClicked {
		fingerprint := req.DeviceID
		if fingerprint == "" {
			fingerprint = req.IPAddress
		}
		if fingerprint != "" {
			priorEvents, err := s.eventRepo.FindByRefCode(ctx, req.RefCode)
			if err == nil {
				for _, e := range priorEvents {
					if e.EventType != domain.EventLinkClicked {
						continue
					}
					prev := e.DeviceID
					if prev == "" {
						prev = e.IPAddress
					}
					if prev == fingerprint {
						alreadyCounted = true
						break
					}
				}
			}
		}
	}

	channel := req.Channel
	if channel == "" {
		channel = link.Channel
	}

	event := &domain.ReferralEvent{
		RefCode:            req.RefCode,
		EventID:            uuid.New().String(),
		EventType:          req.EventType,
		OccurredAt:         time.Now().UnixMilli(),
		Platform:           req.Platform,
		Channel:            channel,
		NewUserID:          req.NewUserID,
		OwnerUserID:        link.OwnerUserID,
		IPAddress:          req.IPAddress,
		DeviceID:           req.DeviceID,
		UserAgent:          req.UserAgent,
		FirebaseInstanceID: req.FirebaseInstanceID,
	}

	if err := s.eventRepo.Save(ctx, event); err != nil {
		log.Error("track event: save failed",
			logger.RefCode(req.RefCode),
			logger.EventType(string(req.EventType)),
			zap.Error(err),
		)
		return fmt.Errorf("track event: save: %w", err)
	}

	if req.EventType == domain.EventLinkClicked && !alreadyCounted {
		if err := s.statsRepo.IncrementCounters(ctx, link.OwnerUserID, 1, 0, 0, 0); err != nil {
			log.Warn("track event: failed to increment invited counter",
				logger.OwnerUserID(link.OwnerUserID),
				zap.Error(err),
			)
		}
	}

	log.Info("referral event tracked",
		logger.RefCode(req.RefCode),
		logger.EventType(string(req.EventType)),
		logger.Platform(req.Platform),
		logger.Channel(string(channel)),
		logger.NewUserID(req.NewUserID),
	)

	return nil
}

func (s *referralService) ConfirmInstall(ctx context.Context, req ports.ConfirmInstallRequest) (*ports.ConfirmInstallResponse, error) {
	log := s.logWithTrace(ctx)

	if req.RefCode == "" {
		return &ports.ConfirmInstallResponse{Attributed: false}, nil
	}

	link, err := s.linkRepo.FindByRefCode(ctx, req.RefCode)
	if err != nil {
		log.Error("confirm install: find link failed",
			logger.RefCode(req.RefCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("confirm install: link not found: %w", err)
	}

	if req.NewUserID != "" && link.OwnerUserID == req.NewUserID {
		return nil, fmt.Errorf("confirm install: %w", domain.ErrSelfReferral)
	}

	if req.NewUserID != "" {
		priorEvents, err := s.eventRepo.FindByNewUserID(ctx, req.NewUserID)
		if err != nil {
			log.Error("confirm install: check prior attribution failed",
				logger.NewUserID(req.NewUserID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("confirm install: check prior attribution: %w", err)
		}
		for _, e := range priorEvents {
			if e.EventType == domain.EventAppInstalled && e.OwnerUserID == link.OwnerUserID {
				log.Warn("confirm install: new_user_id already attributed to this owner",
					logger.RefCode(req.RefCode),
					logger.NewUserID(req.NewUserID),
					logger.OwnerUserID(link.OwnerUserID),
				)
				return nil, fmt.Errorf("confirm install: %w", domain.ErrAlreadyReferred)
			}
		}
	}

	if !link.IsActive() {
		reason := string(link.Status)
		if link.IsExpired() {
			reason = "expired"
		}
		log.Warn("confirm install: link not active",
			logger.RefCode(req.RefCode),
			zap.String("reason", reason),
		)
		return &ports.ConfirmInstallResponse{Attributed: false}, nil
	}

	now := time.Now().UTC()

	if err := s.linkRepo.UpdateStatus(ctx, req.RefCode, domain.StatusUsed); err != nil {
		if errors.Is(err, domain.ErrLinkNotActive) {
			// Concurrent ConfirmInstall already claimed this link.
			return &ports.ConfirmInstallResponse{Attributed: false}, nil
		}
		log.Error("confirm install: update link status failed",
			logger.RefCode(req.RefCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("confirm install: update status: %w", err)
	}

	if err := s.eventRepo.Save(ctx, &domain.ReferralEvent{
		RefCode:            req.RefCode,
		EventID:            uuid.New().String(),
		EventType:          domain.EventAppInstalled,
		OccurredAt:         now.UnixMilli(),
		Platform:           req.Platform,
		NewUserID:          req.NewUserID,
		OwnerUserID:        link.OwnerUserID,
		DeviceID:           req.DeviceID,
		FirebaseInstanceID: req.FirebaseInstanceID,
	}); err != nil {
		log.Warn("confirm install: failed to save app_installed event",
			logger.RefCode(req.RefCode),
			zap.Error(err),
		)
	}

	rewardValue := s.resolveRewardValue(ctx, link.OwnerUserID)

	rewardedDelta := int64(0)
	rewardAmtDelta := 0.0
	if rewardValue > 0 {
		reward := &domain.ReferralReward{
			OwnerUserID: link.OwnerUserID,
			RefCode:     req.RefCode,
			NewUserID:   req.NewUserID,
			RewardType:  domain.RewardCredit,
			RewardValue: rewardValue,
			Status:      domain.RewardPending,
			CreatedAt:   now.UnixMilli(),
			UpdatedAt:   now.UnixMilli(),
		}
		if err := s.rewardRepo.Save(ctx, reward); err != nil {
			log.Error("confirm install: failed to create reward record",
				logger.RefCode(req.RefCode),
				logger.OwnerUserID(link.OwnerUserID),
				logger.RewardValue(rewardValue),
				zap.Error(err),
			)
		} else {
			rewardedDelta = 1
			rewardAmtDelta = rewardValue
			if err := s.eventRepo.Save(ctx, &domain.ReferralEvent{
				RefCode:     req.RefCode,
				EventID:     uuid.New().String(),
				EventType:   domain.EventRewardGiven,
				OccurredAt:  now.UnixMilli(),
				Platform:    req.Platform,
				NewUserID:   req.NewUserID,
				OwnerUserID: link.OwnerUserID,
			}); err != nil {
				log.Warn("confirm install: failed to save reward_given event",
					logger.RefCode(req.RefCode),
					zap.Error(err),
				)
			}
		}
	}

	if err := s.attributionRepo.Save(ctx, &domain.ReferralAttribution{
		OwnerUserID:        link.OwnerUserID,
		NewUserID:          req.NewUserID,
		RefCode:            req.RefCode,
		Platform:           req.Platform,
		DeviceID:           req.DeviceID,
		AttributedAt:       now.UnixMilli(),
		FirebaseInstanceID: req.FirebaseInstanceID,
	}); err != nil {
		log.Error("confirm install: failed to save attribution",
			logger.RefCode(req.RefCode),
			logger.OwnerUserID(link.OwnerUserID),
			logger.NewUserID(req.NewUserID),
			zap.Error(err),
		)
	}

	if err := s.statsRepo.IncrementCounters(ctx, link.OwnerUserID, 0, 1, rewardedDelta, rewardAmtDelta); err != nil {
		log.Warn("confirm install: failed to increment stats",
			logger.OwnerUserID(link.OwnerUserID),
			zap.Error(err),
		)
	}

	log.Info("install confirmed — attribution successful",
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
		s.logWithTrace(ctx).Error("get user stats failed",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get stats for %s: %w", userID, err)
	}
	return stats, nil
}

func (s *referralService) ActivateReferral(ctx context.Context, req ports.ActivateReferralRequest) (*ports.ActivateReferralResponse, error) {
	resp, err := s.ConfirmInstall(ctx, ports.ConfirmInstallRequest{
		RefCode:            req.RefCode,
		NewUserID:          req.UserID,
		Platform:           req.Platform,
		DeviceID:           req.DeviceID,
		FirebaseInstanceID: req.FirebaseInstanceID,
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
	log := s.logWithTrace(ctx)

	link, err := s.linkRepo.FindByRefCode(ctx, refCode)
	if err != nil {
		log.Error("get referral status: find link failed",
			logger.RefCode(refCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get status %s: %w", refCode, err)
	}

	linkStatus := string(link.Status)
	if link.IsExpired() {
		linkStatus = "expired"
	}

	resp := &ports.ReferralStatusResponse{
		RefCode:          link.RefCode,
		Status:           linkStatus,
		OwnerUserID:      link.OwnerUserID,
		InvitationStatus: domain.InvitationPending,
	}
	if link.IsExpired() && link.Status != domain.StatusUsed {
		resp.InvitationStatus = domain.InvitationExpired
	}

	events, err := s.eventRepo.FindByRefCode(ctx, refCode)
	if err != nil {
		log.Warn("get referral status: failed to fetch events",
			logger.RefCode(refCode),
			zap.Error(err),
		)
		return resp, nil
	}

	var hasClicked, hasInstalled bool
	var invitedAt int64
	for _, evt := range events {
		switch evt.EventType {
		case domain.EventLinkClicked:
			hasClicked = true
			// events are ordered occurred_at DESC, so keep the earliest click seen so far.
			if invitedAt == 0 || evt.OccurredAt < invitedAt {
				invitedAt = evt.OccurredAt
			}
		case domain.EventAppInstalled:
			hasInstalled = true
			t := evt.OccurredAt
			resp.ActivatedAt = &t
			resp.Platform = evt.Platform
			resp.NewUserID = evt.NewUserID
		}
	}

	if hasClicked {
		resp.InvitedAt = &invitedAt
	}
	if hasClicked && hasInstalled && *resp.ActivatedAt > invitedAt {
		delta := *resp.ActivatedAt - invitedAt
		resp.TimeToInstallMs = &delta
	}

	switch {
	case hasInstalled:
		resp.InvitationStatus = domain.InvitationInstalled
	case hasClicked:
		resp.InvitationStatus = domain.InvitationClicked
	}

	if hasInstalled {
		reward, err := s.rewardRepo.FindByOwnerAndRefCode(ctx, link.OwnerUserID, refCode)
		if err != nil {
			log.Warn("get referral status: failed to fetch reward",
				logger.RefCode(refCode),
				logger.OwnerUserID(link.OwnerUserID),
				zap.Error(err),
			)
		} else {
			resp.RewardStatus = reward.Status
			resp.RewardValue = reward.RewardValue
			// Only mark rewarded when the system has actually paid the reward.
			if reward.Status == domain.RewardPaid {
				resp.InvitationStatus = domain.InvitationRewarded
			}
		}
	}

	return resp, nil
}

func (s *referralService) getDailyLinkCount(ctx context.Context, cacheKey, ownerUserID string, startOfDayMs int64, ttl time.Duration) (int64, error) {
	cached, ok, err := s.cache.GetInt64(ctx, cacheKey)
	if err != nil {
		s.log.Warn("getDailyLinkCount: cache get failed, falling back to DB",
			logger.OwnerUserID(ownerUserID),
			zap.Error(err),
		)
	} else if ok {
		return cached, nil
	}

	count, err := s.linkRepo.CountByOwnerUserIDSince(ctx, ownerUserID, startOfDayMs)
	if err != nil {
		return 0, err
	}

	if setErr := s.cache.SetInt64(ctx, cacheKey, count, ttl); setErr != nil {
		s.log.Warn("getDailyLinkCount: cache set failed",
			logger.OwnerUserID(ownerUserID),
			zap.Error(setErr),
		)
	}

	return count, nil
}

func (s *referralService) buildDeepLink(refCode, _ string) string {
	return fmt.Sprintf("%s/r/%s", s.cfg.DeepLink.PublicURL, refCode)
}

func (s *referralService) resolveRewardValue(ctx context.Context, ownerUserID string) float64 {
	log := s.logWithTrace(ctx)

	program, err := s.programRepo.FindActive(ctx)
	if err != nil {
		log.Warn("resolveRewardValue: failed to fetch active program",
			zap.String("owner_user_id", ownerUserID),
			zap.Error(err),
		)
		return 0
	}
	if program == nil {
		return 0
	}

	stats, err := s.statsRepo.Get(ctx, ownerUserID)
	if err != nil {
		log.Warn("resolveRewardValue: failed to get stats, skipping reward",
			zap.String("owner_user_id", ownerUserID),
			zap.Error(err),
		)
		return 0
	}

	// installNumber is 1-based: first install = 1
	installNumber := stats.TotalInstalled + 1
	return program.RewardForInstall(installNumber)
}

func (s *referralService) CreateRewardProgram(ctx context.Context, req ports.CreateRewardProgramRequest) (*domain.RewardProgram, error) {
	log := s.logWithTrace(ctx)

	now := time.Now().UnixMilli()
	tiers := make([]domain.RewardTier, len(req.Tiers))
	for i, t := range req.Tiers {
		tiers[i] = domain.RewardTier{
			MinCount:    t.MinCount,
			MaxCount:    t.MaxCount,
			RewardValue: t.RewardValue,
		}
	}

	status := domain.ProgramStatusInactive
	if req.Activate {
		status = domain.ProgramStatusActive
	}

	program := &domain.RewardProgram{
		ProgramID: uuid.New().String(),
		Name:      req.Name,
		Status:    status,
		Tiers:     tiers,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if req.Activate {
		existing, err := s.programRepo.FindActive(ctx)
		if err != nil {
			log.Warn("create reward program: failed to find existing active program",
				zap.Error(err),
			)
		} else if existing != nil {
			if err := s.programRepo.UpdateStatus(ctx, existing.ProgramID, domain.ProgramStatusInactive); err != nil {
				log.Warn("create reward program: failed to deactivate old program",
					zap.String("old_program_id", existing.ProgramID),
					zap.Error(err),
				)
			}
		}
	}

	if err := s.programRepo.Save(ctx, program); err != nil {
		log.Error("create reward program: save failed",
			zap.String("program_id", program.ProgramID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("create reward program: %w", err)
	}

	log.Info("reward program created",
		zap.String("program_id", program.ProgramID),
		zap.String("name", program.Name),
		zap.String("status", string(program.Status)),
		zap.Int("tiers", len(tiers)),
	)

	return program, nil
}

func (s *referralService) GetActiveRewardProgram(ctx context.Context) (*domain.RewardProgram, error) {
	program, err := s.programRepo.FindActive(ctx)
	if err != nil {
		s.logWithTrace(ctx).Error("get active reward program failed", zap.Error(err))
		return nil, fmt.Errorf("get active reward program: %w", err)
	}
	return program, nil
}

func (s *referralService) SetRewardProgramStatus(ctx context.Context, programID string, active bool) error {
	log := s.logWithTrace(ctx)

	if active {
		existing, err := s.programRepo.FindActive(ctx)
		if err != nil {
			log.Warn("set program status: failed to find current active program",
				zap.String("program_id", programID),
				zap.Error(err),
			)
		} else if existing != nil && existing.ProgramID != programID {
			if err := s.programRepo.UpdateStatus(ctx, existing.ProgramID, domain.ProgramStatusInactive); err != nil {
				log.Warn("set program status: failed to deactivate old program",
					zap.String("old_program_id", existing.ProgramID),
					zap.Error(err),
				)
			}
		}

		if err := s.programRepo.UpdateStatus(ctx, programID, domain.ProgramStatusActive); err != nil {
			log.Error("set program status: activate failed",
				zap.String("program_id", programID),
				zap.Error(err),
			)
			return err
		}
		return nil
	}

	if err := s.programRepo.UpdateStatus(ctx, programID, domain.ProgramStatusInactive); err != nil {
		log.Error("set program status: deactivate failed",
			zap.String("program_id", programID),
			zap.Error(err),
		)
		return err
	}
	return nil
}
