package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"go.uber.org/fx"
)

const (
	maxIdempotencyKey = 128
	maxWinners        = 50
)

type GameService struct {
	campaigns port.CampaignRepository
	plays     port.PlayRepository
	limiter   port.RateLimiter
	winners   port.WinnerFeed
	guard     *Guard
	clock     port.Clock
	random    port.Random
	perMinute int
}

type GameDeps struct {
	fx.In

	Config    *config.Config
	Campaigns port.CampaignRepository
	Plays     port.PlayRepository
	Limiter   port.RateLimiter `optional:"true"`
	Winners   port.WinnerFeed  `optional:"true"`
	Guard     *Guard           `optional:"true"`
	Clock     port.Clock
	Random    port.Random
}

func NewGameService(d GameDeps) port.GameUsecase {
	return &GameService{
		campaigns: d.Campaigns,
		plays:     d.Plays,
		limiter:   d.Limiter,
		winners:   d.Winners,
		guard:     d.Guard,
		clock:     d.Clock,
		random:    d.Random,
		perMinute: d.Config.RateLimit.PlaysPerMinute,
	}
}

func requirePlayer(actor model.Actor) error {
	if actor.Service {
		return fmt.Errorf("%w: only end users can play", model.ErrForbidden)
	}
	if actor.UserID == "" {
		return model.ErrUnauthenticated
	}
	return nil
}

func throttle(ctx context.Context, limiter port.RateLimiter, perMinute int, key string) error {
	if limiter == nil || perMinute <= 0 {
		return nil
	}
	ok, err := limiter.Allow(ctx, key, perMinute, time.Minute)
	if err != nil {
		slog.Warn("rate limiter unavailable, allowing play", "err", err)
		return nil
	}
	if !ok {
		return model.ErrRateLimited
	}
	return nil
}

func instant(ctx context.Context, campaigns port.CampaignRepository, clock port.Clock, game string) (*model.Campaign, error) {
	g, err := model.ParseGame(game)
	if err != nil {
		return nil, err
	}
	if !g.Instant() {
		return nil, &model.ValidationError{Field: "game_type", Reason: "use the checkin endpoints for daily check-in"}
	}
	return running(ctx, campaigns, clock, g)
}

func status(c *model.Campaign, day string, u model.Usage) port.PlayerStatus {
	return port.PlayerStatus{
		CampaignID: c.ID,
		GameType:   c.GameType,
		Day:        day,
		DailySpins: c.DailySpins,
		Remaining:  c.Remaining(u),
		Usage:      u,
		EndsAt:     c.EndAt,
	}
}

func (s *GameService) Status(ctx context.Context, actor model.Actor, game string) (*port.PlayerStatus, error) {
	if err := requirePlayer(actor); err != nil {
		return nil, err
	}
	c, err := instant(ctx, s.campaigns, s.clock, game)
	if err != nil {
		return nil, err
	}
	day := c.Day(s.clock.Now())
	u, err := s.plays.Usage(ctx, c.ID, actor.UserID, day)
	if err != nil {
		return nil, err
	}
	st := status(c, day, u)
	return &st, nil
}

func (s *GameService) Play(ctx context.Context, actor model.Actor, in port.PlayInput) (*port.PlayResult, error) {
	if err := requirePlayer(actor); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(in.IdempotencyKey)
	if len(key) > maxIdempotencyKey {
		return nil, &model.ValidationError{Field: "Idempotency-Key", Reason: "must be at most 128 characters"}
	}
	c, err := instant(ctx, s.campaigns, s.clock, in.Game)
	if err != nil {
		return nil, err
	}
	if err := c.ValidateChoice(in.Choice); err != nil {
		return nil, err
	}
	if err := throttle(ctx, s.limiter, s.perMinute, "play:"+actor.UserID); err != nil {
		return nil, err
	}
	client := in.Client.Normalize()
	hold, err := s.guard.Screen(ctx, actor, client)
	if err != nil {
		return nil, err
	}
	var result *port.PlayResult
	err = s.plays.WithPlayer(ctx, c.ID, actor.UserID, func(l port.PlayerLedger) error {
		now := s.clock.Now()
		if !c.RunningAt(now) {
			return model.ErrNoActiveCampaign
		}
		day := c.Day(now)
		if key != "" {
			prev, err := l.PlayByKey(ctx, key)
			if err != nil {
				return err
			}
			if prev != nil {
				u, err := l.Usage(ctx, day)
				if err != nil {
					return err
				}
				result = &port.PlayResult{Play: *prev, Won: prev.Won(), Replayed: true, Status: status(c, day, u)}
				return nil
			}
		}
		u, err := l.Usage(ctx, day)
		if err != nil {
			return err
		}
		source, ok := c.Allowance(u)
		if !ok {
			return model.ErrNoSpinsLeft
		}
		prize, err := s.claim(ctx, l, c, day)
		if err != nil {
			return err
		}
		reason := hold
		if prize.Type == model.PrizePoints && reason == "" && c.NeedsReview(prize.Points) {
			reason = model.HoldLargeReward
		}
		p := &model.Play{
			CampaignID:     c.ID,
			GameType:       c.GameType,
			UserID:         actor.UserID,
			PrizeID:        prize.ID,
			PrizeName:      prize.Name,
			PrizeType:      prize.Type,
			Points:         prize.Points,
			Position:       prize.Position,
			Choice:         in.Choice,
			Source:         source,
			Day:            day,
			Client:         client,
			IdempotencyKey: key,
			HoldReason:     reason,
			CreatedAt:      now,
		}
		if err := l.SavePlay(ctx, p); err != nil {
			return err
		}
		if source == model.SourceDaily {
			u.DailyToday++
		} else {
			u.CreditToday++
			u.CreditBalance--
		}
		result = &port.PlayResult{Play: *p, Won: p.Won(), Status: status(c, day, u)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if result.Won && !result.Replayed && s.winners != nil {
		if err := s.winners.Push(ctx, model.PlayWinner(result.Play)); err != nil {
			slog.Warn("winner feed unavailable", "err", err)
		}
	}
	return result, nil
}

func (s *GameService) claim(ctx context.Context, l port.PlayerLedger, c *model.Campaign, day string) (*model.Prize, error) {
	blank := c.Blank()
	if blank == nil {
		return nil, fmt.Errorf("campaign %d has no blank slot", c.ID)
	}
	prize := c.Draw(s.random.Intn)
	if prize.ID != blank.ID {
		ok, err := s.affordable(ctx, l, c, day, prize.Points)
		if err != nil {
			return nil, err
		}
		if ok {
			ok, err = l.ClaimPrize(ctx, prize.ID, prize.Points)
			if err != nil {
				return nil, err
			}
		}
		if ok {
			return prize, nil
		}
	}
	if _, err := l.ClaimPrize(ctx, blank.ID, 0); err != nil {
		return nil, err
	}
	return blank, nil
}

func (s *GameService) affordable(ctx context.Context, l port.PlayerLedger, c *model.Campaign, day string, points int64) (bool, error) {
	if c.MaxPointsDay == 0 {
		return true, nil
	}
	won, err := l.PointsWonOn(ctx, day)
	if err != nil {
		return false, err
	}
	return c.WithinDailyCap(won, points), nil
}

func (s *GameService) History(ctx context.Context, actor model.Actor, in port.ListPlaysInput) (*port.PlayPage, error) {
	if err := requirePlayer(actor); err != nil {
		return nil, err
	}
	in.UserID = actor.UserID
	return s.list(ctx, in)
}

func (s *GameService) CampaignPlays(ctx context.Context, actor model.Actor, in port.ListPlaysInput) (*port.PlayPage, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	if _, err := s.campaigns.Get(ctx, in.CampaignID); err != nil {
		return nil, err
	}
	return s.list(ctx, in)
}

func (s *GameService) list(ctx context.Context, in port.ListPlaysInput) (*port.PlayPage, error) {
	f := port.PlayFilter{CampaignID: in.CampaignID, UserID: in.UserID}
	if in.GameType != "" {
		g, err := model.ParseGame(in.GameType)
		if err != nil {
			return nil, err
		}
		f.GameType = g
	}
	f.Limit, f.Offset = page(in.Limit, in.Offset)
	items, total, err := s.plays.List(ctx, f)
	if err != nil {
		return nil, err
	}
	return &port.PlayPage{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

func (s *GameService) Winners(ctx context.Context, game string, limit int) ([]model.Winner, error) {
	c, err := instant(ctx, s.campaigns, s.clock, game)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > maxWinners {
		limit = 20
	}
	if s.winners != nil {
		w, err := s.winners.Recent(ctx, c.ID, limit)
		if err == nil && len(w) > 0 {
			return w, nil
		}
		if err != nil {
			slog.Warn("winner feed unavailable, reading from database", "err", err)
		}
	}
	plays, err := s.plays.RecentWins(ctx, c.ID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]model.Winner, 0, len(plays))
	for _, p := range plays {
		out = append(out, model.PlayWinner(p))
	}
	return out, nil
}

func (s *GameService) GrantPlays(ctx context.Context, actor model.Actor, in port.GrantInput) (*port.GrantResult, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.ReferenceID = strings.TrimSpace(in.ReferenceID)
	in.Event = strings.ToLower(strings.TrimSpace(in.Event))
	switch {
	case in.UserID == "" || len(in.UserID) > 64:
		return nil, &model.ValidationError{Field: "user_id", Reason: "must be 1-64 characters"}
	case in.ReferenceID == "" || len(in.ReferenceID) > 128:
		return nil, &model.ValidationError{Field: "reference_id", Reason: "must be 1-128 characters"}
	case in.Event == "":
		return nil, &model.ValidationError{Field: "event", Reason: "is required"}
	}
	now := s.clock.Now()
	all, err := s.campaigns.Running(ctx, now)
	if err != nil {
		return nil, err
	}
	res := &port.GrantResult{Grants: []model.CreditGrant{}, Skips: []port.GrantSkip{}}
	for i := range all {
		c := &all[i]
		rule, ok := c.Rule(in.Event)
		if !c.GameType.Instant() || !c.RunningAt(now) || !ok {
			continue
		}
		g, err := s.grant(ctx, c, rule, in, now)
		switch {
		case errors.Is(err, model.ErrCreditLimit):
			res.Skips = append(res.Skips, port.GrantSkip{CampaignID: c.ID, GameType: c.GameType, Reason: "credit_limit_reached"})
		case err != nil:
			return nil, err
		default:
			res.Grants = append(res.Grants, *g)
		}
	}
	if len(res.Grants) == 0 && len(res.Skips) == 0 {
		return nil, &model.ValidationError{Field: "event", Reason: "no running game grants extra plays for this event"}
	}
	return res, nil
}

func (s *GameService) grant(ctx context.Context, c *model.Campaign, rule model.SpinRule, in port.GrantInput, now time.Time) (*model.CreditGrant, error) {
	var out *model.CreditGrant
	err := s.plays.WithPlayer(ctx, c.ID, in.UserID, func(l port.PlayerLedger) error {
		prev, err := l.Grant(ctx, rule.Event, in.ReferenceID)
		if err != nil {
			return err
		}
		if prev != nil {
			out = prev
			return nil
		}
		day := c.Day(now)
		if rule.MaxPerDay > 0 {
			n, err := l.GrantsOn(ctx, rule.Event, day)
			if err != nil {
				return err
			}
			if n >= rule.MaxPerDay {
				return model.ErrCreditLimit
			}
		}
		g := &model.CreditGrant{
			CampaignID:  c.ID,
			GameType:    c.GameType,
			UserID:      in.UserID,
			Event:       rule.Event,
			ReferenceID: in.ReferenceID,
			Spins:       rule.Spins,
			Day:         day,
			CreatedAt:   now,
		}
		if err := l.SaveGrant(ctx, g); err != nil {
			return err
		}
		out = g
		return nil
	})
	return out, err
}
