package application

import (
	"context"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"go.uber.org/fx"
)

type CheckinService struct {
	campaigns port.CampaignRepository
	plays     port.PlayRepository
	limiter   port.RateLimiter
	guard     *Guard
	clock     port.Clock
	perMinute int
}

type CheckinDeps struct {
	fx.In

	Config    *config.Config
	Campaigns port.CampaignRepository
	Plays     port.PlayRepository
	Limiter   port.RateLimiter `optional:"true"`
	Guard     *Guard           `optional:"true"`
	Clock     port.Clock
}

func NewCheckinService(d CheckinDeps) port.CheckinUsecase {
	return &CheckinService{campaigns: d.Campaigns, plays: d.Plays, limiter: d.Limiter, guard: d.Guard, clock: d.Clock, perMinute: d.Config.RateLimit.PlaysPerMinute}
}

func checkinStatus(c *model.Campaign, day string, last *model.Checkin) port.CheckinStatus {
	st := port.CheckinStatus{
		CampaignID: c.ID,
		Day:        day,
		Streak:     model.CurrentStreak(last, day),
		Rewards:    c.CheckinRewards,
		EndsAt:     c.EndAt,
	}
	st.CheckedInToday = last != nil && last.Day == day
	next := model.NextStreak(last, day)
	if st.CheckedInToday {
		next = last.Streak + 1
	}
	st.CycleDay = (next-1)%len(c.CheckinRewards) + 1
	st.NextReward = c.CheckinReward(next)
	return st
}

func (s *CheckinService) Status(ctx context.Context, actor model.Actor) (*port.CheckinStatus, error) {
	if err := requirePlayer(actor); err != nil {
		return nil, err
	}
	c, err := running(ctx, s.campaigns, s.clock, model.GameCheckin)
	if err != nil {
		return nil, err
	}
	last, err := s.plays.LastCheckin(ctx, c.ID, actor.UserID)
	if err != nil {
		return nil, err
	}
	st := checkinStatus(c, c.Day(s.clock.Now()), last)
	return &st, nil
}

func (s *CheckinService) CheckIn(ctx context.Context, actor model.Actor, client model.ClientInfo) (*port.CheckinResult, error) {
	if err := requirePlayer(actor); err != nil {
		return nil, err
	}
	c, err := running(ctx, s.campaigns, s.clock, model.GameCheckin)
	if err != nil {
		return nil, err
	}
	if err := throttle(ctx, s.limiter, s.perMinute, "checkin:"+actor.UserID); err != nil {
		return nil, err
	}
	client = client.Normalize()
	hold, err := s.guard.Screen(ctx, actor, client)
	if err != nil {
		return nil, err
	}
	var result *port.CheckinResult
	err = s.plays.WithPlayer(ctx, c.ID, actor.UserID, func(l port.PlayerLedger) error {
		now := s.clock.Now()
		if !c.RunningAt(now) {
			return model.ErrNoActiveCampaign
		}
		day := c.Day(now)
		last, err := l.LastCheckin(ctx)
		if err != nil {
			return err
		}
		if last != nil && last.Day == day {
			return model.ErrAlreadyCheckedIn
		}
		streak := model.NextStreak(last, day)
		points := c.CheckinReward(streak)
		reason := hold
		if reason == "" && c.NeedsReview(points) {
			reason = model.HoldLargeReward
		}
		ci := &model.Checkin{
			CampaignID: c.ID,
			UserID:     actor.UserID,
			Day:        day,
			Streak:     streak,
			Points:     points,
			Client:     client,
			HoldReason: reason,
			CreatedAt:  now,
		}
		if err := l.SaveCheckin(ctx, ci); err != nil {
			return err
		}
		result = &port.CheckinResult{Checkin: *ci, Status: checkinStatus(c, day, ci)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *CheckinService) History(ctx context.Context, actor model.Actor, limit, offset int) (*port.CheckinPage, error) {
	if err := requirePlayer(actor); err != nil {
		return nil, err
	}
	limit, offset = page(limit, offset)
	items, total, err := s.plays.Checkins(ctx, port.CheckinFilter{UserID: actor.UserID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return &port.CheckinPage{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}
