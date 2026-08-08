package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const schedulerActor = "scheduler"

type PublisherRegistry map[model.Channel]port.ChannelPublisher

type publisherRegistryParams struct {
	fx.In

	Publishers []port.ChannelPublisher `group:"content_publishers"`
}

func NewPublisherRegistry(p publisherRegistryParams) PublisherRegistry {
	registry := make(PublisherRegistry, len(p.Publishers))
	for _, pub := range p.Publishers {
		registry[pub.Channel()] = pub
	}
	return registry
}

type contentService struct {
	repo       port.PostRepository
	publishers PublisherRegistry
}

func NewContentService(repo port.PostRepository, publishers PublisherRegistry) port.ContentUsecase {
	return &contentService{repo: repo, publishers: publishers}
}

func (s *contentService) CreatePost(ctx context.Context, input port.CreatePostInput) (*model.Post, error) {
	if input.Text == "" && len(input.Media) == 0 {
		return nil, fmt.Errorf("post must have text or media")
	}
	if input.CronExpr != "" {
		if err := s.validateCronExpr(input.CronExpr); err != nil {
			return nil, err
		}
	}

	mediaKind := input.MediaKind
	if mediaKind == "" {
		mediaKind = model.DeriveMediaKind(input.Media)
	}

	now := time.Now()
	post := &model.Post{
		ID:              uuid.NewString(),
		Title:           input.Title,
		Text:            input.Text,
		Hashtags:        input.Hashtags,
		Media:           input.Media,
		MediaKind:       mediaKind,
		Channels:        input.Channels,
		Timezone:        input.Timezone,
		CronExpr:        input.CronExpr,
		MaxRunsPerDay:   input.MaxRunsPerDay,
		MaxRunsPerMonth: input.MaxRunsPerMonth,
		Campaign:        input.Campaign,
		CreatedBy:       input.CreatedBy,
		StatusChangedBy: input.CreatedBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if input.ScheduledAt != nil {
		post.ScheduledAt = *input.ScheduledAt
	}

	if input.SaveAsDraft {
		post.Status = model.PostStatusDraft
	} else {
		if err := s.validateChannels(input.Channels); err != nil {
			return nil, err
		}
		if post.IsRecurring() {
			post.Status = model.PostStatusActive
		} else {
			if post.ScheduledAt.IsZero() {
				post.ScheduledAt = now
			}
			post.Status = model.PostStatusScheduled
		}
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *contentService) UpdatePost(ctx context.Context, id string, input port.UpdatePostInput) (*model.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.Status != model.PostStatusDraft && post.Status != model.PostStatusPendingReview {
		return nil, fmt.Errorf("post %q cannot be edited (status=%s)", id, post.Status)
	}

	if input.Title != nil {
		post.Title = *input.Title
	}
	if input.Text != nil {
		post.Text = *input.Text
	}
	if input.Hashtags != nil {
		post.Hashtags = input.Hashtags
	}
	if input.Media != nil {
		post.Media = input.Media
	}
	if input.MediaKind != nil {
		post.MediaKind = *input.MediaKind
	}
	if input.Channels != nil {
		post.Channels = input.Channels
	}
	if input.ScheduledAt != nil {
		post.ScheduledAt = *input.ScheduledAt
	}
	if input.Timezone != nil {
		post.Timezone = *input.Timezone
	}
	if input.CronExpr != nil {
		if *input.CronExpr != "" {
			if err := s.validateCronExpr(*input.CronExpr); err != nil {
				return nil, err
			}
		}
		post.CronExpr = *input.CronExpr
	}
	if input.MaxRunsPerDay != nil {
		post.MaxRunsPerDay = *input.MaxRunsPerDay
	}
	if input.MaxRunsPerMonth != nil {
		post.MaxRunsPerMonth = *input.MaxRunsPerMonth
	}
	if input.Campaign != nil {
		post.Campaign = *input.Campaign
	}

	if post.Text == "" && len(post.Media) == 0 {
		return nil, fmt.Errorf("post must have text or media")
	}
	if input.MediaKind == nil && input.Media != nil {
		post.MediaKind = model.DeriveMediaKind(post.Media)
	}
	if len(post.Channels) > 0 {
		if err := s.validateChannels(post.Channels); err != nil {
			return nil, err
		}
	}

	if post.Status == model.PostStatusPendingReview {
		post.Status = model.PostStatusDraft
	}
	post.RejectReason = ""
	post.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *contentService) GetPost(ctx context.Context, id string) (*model.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *contentService) ListPosts(ctx context.Context, limit, offset int32, status model.PostStatus, campaign string) ([]model.Post, error) {
	return s.repo.List(ctx, limit, offset, status, campaign)
}

func (s *contentService) SubmitForReview(ctx context.Context, id string, changedBy string) (*model.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.Status != model.PostStatusDraft {
		return nil, fmt.Errorf("post %q cannot be submitted for review (status=%s)", id, post.Status)
	}
	if post.Text == "" && len(post.Media) == 0 {
		return nil, fmt.Errorf("post must have text or media")
	}
	if err := s.validateChannels(post.Channels); err != nil {
		return nil, err
	}
	if post.CronExpr != "" {
		if err := s.validateCronExpr(post.CronExpr); err != nil {
			return nil, err
		}
	}
	if !post.HasSchedule() {
		return nil, fmt.Errorf("post must have scheduled_at or cron_expr set before it can be submitted for review")
	}

	post.Status = model.PostStatusPendingReview
	post.RejectReason = ""
	post.StatusChangedBy = changedBy
	post.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *contentService) ApprovePost(ctx context.Context, id string, changedBy string) (*model.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.Status != model.PostStatusPendingReview {
		return nil, fmt.Errorf("post %q cannot be approved (status=%s)", id, post.Status)
	}

	if post.IsRecurring() {
		post.Status = model.PostStatusActive
	} else {
		if post.ScheduledAt.IsZero() {
			post.ScheduledAt = time.Now()
		}
		post.Status = model.PostStatusScheduled
	}
	post.RejectReason = ""
	post.ApprovedBy = changedBy
	post.StatusChangedBy = changedBy
	post.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *contentService) RejectPost(ctx context.Context, id string, changedBy string, reason string) (*model.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.Status != model.PostStatusPendingReview {
		return nil, fmt.Errorf("post %q cannot be rejected (status=%s)", id, post.Status)
	}

	post.Status = model.PostStatusDraft
	post.RejectReason = reason
	post.StatusChangedBy = changedBy
	post.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *contentService) validateChannels(channels []model.Channel) error {
	if len(channels) == 0 {
		return fmt.Errorf("post must target at least one channel")
	}
	for _, ch := range channels {
		if _, ok := s.publishers[ch]; !ok {
			return fmt.Errorf("no publisher registered for channel %q", ch)
		}
	}
	return nil
}

func (s *contentService) validateCronExpr(expr string) error {
	if _, err := cron.ParseStandard(expr); err != nil {
		return fmt.Errorf("invalid cron_expr %q: %w", expr, err)
	}
	return nil
}

func (s *contentService) PublishNow(ctx context.Context, id string, changedBy string) (*model.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if post.IsRecurring() {
		err = s.runRecurring(ctx, post, now, true, changedBy)
	} else {
		err = s.publishOneOff(ctx, post, changedBy)
	}
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (s *contentService) CancelScheduledPost(ctx context.Context, id string, changedBy string) error {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	switch post.Status {
	case model.PostStatusDraft, model.PostStatusPendingReview, model.PostStatusScheduled, model.PostStatusActive:
	default:
		return fmt.Errorf("post %q cannot be cancelled (status=%s)", id, post.Status)
	}

	post.Status = model.PostStatusCancelled
	post.StatusChangedBy = changedBy
	post.UpdatedAt = time.Now()
	return s.repo.Update(ctx, post)
}

func (s *contentService) PublishDuePosts(ctx context.Context) error {
	lg := logger.WithContext(ctx)
	now := time.Now()

	dueOneOff, err := s.repo.DueForPublish(ctx, now)
	if err != nil {
		return fmt.Errorf("fetch due one-off posts: %w", err)
	}
	for i := range dueOneOff {
		post := &dueOneOff[i]
		if post.Status != model.PostStatusScheduled {
			continue
		}
		if err := s.publishOneOff(ctx, post, schedulerActor); err != nil {
			lg.Error("PublishDuePosts: update one-off post", zap.String("post_id", post.ID), zap.Error(err))
		}
	}

	recurring, err := s.repo.ActiveRecurring(ctx)
	if err != nil {
		return fmt.Errorf("fetch active recurring posts: %w", err)
	}
	for i := range recurring {
		post := &recurring[i]
		if post.Status != model.PostStatusActive {
			continue
		}

		schedule, err := cron.ParseStandard(post.CronExpr)
		if err != nil {
			lg.Error("PublishDuePosts: invalid cron_expr on stored post",
				zap.String("post_id", post.ID), zap.String("cron_expr", post.CronExpr), zap.Error(err))
			continue
		}

		reference := post.CreatedAt
		if post.LastRunAt != nil {
			reference = *post.LastRunAt
		}
		if schedule.Next(reference).After(now) {
			continue
		}

		if err := s.runRecurring(ctx, post, now, false, schedulerActor); err != nil {
			lg.Error("PublishDuePosts: update recurring post", zap.String("post_id", post.ID), zap.Error(err))
		}
	}
	return nil
}

func (s *contentService) publishOneOff(ctx context.Context, post *model.Post, changedBy string) error {
	lg := logger.WithContext(ctx)

	post.Status = model.PostStatusPublishing
	post.StatusChangedBy = changedBy
	post.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, post); err != nil {
		lg.Error("publishOneOff: lock post", zap.String("post_id", post.ID), zap.Error(err))
	}

	results, allOK := s.dispatch(ctx, *post, true)
	post.Results = results
	if allOK {
		post.Status = model.PostStatusPublished
	} else {
		post.Status = model.PostStatusFailed
	}
	post.UpdatedAt = time.Now()
	return s.repo.Update(ctx, post)
}

func (s *contentService) runRecurring(ctx context.Context, post *model.Post, now time.Time, bypassCap bool, changedBy string) error {
	lg := logger.WithContext(ctx)

	day := now.Format("2006-01-02")
	month := now.Format("2006-01")
	if post.LastRunDate != day {
		post.RunsToday = 0
		post.LastRunDate = day
	}
	if post.LastRunMonth != month {
		post.RunsThisMonth = 0
		post.LastRunMonth = month
	}

	if !bypassCap {
		if post.MaxRunsPerDay > 0 && post.RunsToday >= post.MaxRunsPerDay {
			s.skipCapped(post, now, "daily", post.MaxRunsPerDay)
			post.StatusChangedBy = changedBy
			return s.repo.Update(ctx, post)
		}
		if post.MaxRunsPerMonth > 0 && post.RunsThisMonth >= post.MaxRunsPerMonth {
			s.skipCapped(post, now, "monthly", post.MaxRunsPerMonth)
			post.StatusChangedBy = changedBy
			return s.repo.Update(ctx, post)
		}
	}

	activeStatus := post.Status
	post.Status = model.PostStatusPublishing
	post.UpdatedAt = now
	if err := s.repo.Update(ctx, post); err != nil {
		lg.Error("runRecurring: lock post", zap.String("post_id", post.ID), zap.Error(err))
	}

	results, _ := s.dispatch(ctx, *post, false)
	post.Results = results
	post.RunsToday++
	post.RunsThisMonth++
	post.LastRunAt = &now
	post.Status = activeStatus
	post.StatusChangedBy = changedBy
	post.UpdatedAt = now
	return s.repo.Update(ctx, post)
}

func (s *contentService) skipCapped(post *model.Post, now time.Time, window string, limit int) {
	post.Results = model.ResultList{{
		Error:       fmt.Sprintf("skipped: %s run cap (%d) reached", window, limit),
		PublishedAt: now,
	}}
	post.LastRunAt = &now
	post.UpdatedAt = now
}

func (s *contentService) dispatch(ctx context.Context, post model.Post, skipExisting bool) ([]model.ChannelPublishResult, bool) {
	lg := logger.WithContext(ctx)

	alreadyPublished := map[model.Channel]model.ChannelPublishResult{}
	if skipExisting {
		for _, r := range post.Results {
			if r.Error == "" {
				alreadyPublished[r.Channel] = r
			}
		}
	}

	results := make([]model.ChannelPublishResult, 0, len(post.Channels))
	allOK := true
	for _, ch := range post.Channels {
		if prev, ok := alreadyPublished[ch]; ok {
			results = append(results, prev)
			continue
		}

		res := model.ChannelPublishResult{Channel: ch, PublishedAt: time.Now()}

		publisher, ok := s.publishers[ch]
		switch {
		case !ok:
			res.Error = fmt.Sprintf("no publisher registered for channel %q", ch)
			allOK = false
		default:
			externalID, err := publisher.Publish(ctx, post)
			if err != nil {
				lg.Error("dispatch: channel publish failed",
					zap.String("post_id", post.ID),
					zap.String("channel", string(ch)),
					zap.Error(err),
				)
				res.Error = err.Error()
				allOK = false
			} else {
				res.ExternalID = externalID
				res.PublishedURL = buildPublishedURL(ch, externalID)
			}
		}
		results = append(results, res)
	}
	return results, allOK
}

func buildPublishedURL(ch model.Channel, externalID string) string {
	if externalID == "" {
		return ""
	}
	switch ch {
	case model.ChannelTwitter:
		return "https://twitter.com/i/web/status/" + externalID
	case model.ChannelFacebook:
		return "https://www.facebook.com/" + externalID
	default:
		return ""
	}
}
