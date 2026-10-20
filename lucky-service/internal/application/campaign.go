package application

import (
	"context"
	"fmt"
	"slices"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type CampaignService struct {
	campaigns       port.CampaignRepository
	clock           port.Clock
	defaultTimezone string
}

func NewCampaignService(cfg *config.Config, campaigns port.CampaignRepository, clock port.Clock) port.CampaignUsecase {
	return &CampaignService{campaigns: campaigns, clock: clock, defaultTimezone: cfg.DefaultTimezone}
}

func requireAdmin(actor model.Actor) error {
	if !actor.Admin && !actor.Service {
		return model.ErrForbidden
	}
	return nil
}

func page(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}
	return min(limit, maxLimit), max(offset, 0)
}

func (s *CampaignService) build(c *model.Campaign, in port.CampaignInput) {
	c.Name = in.Name
	c.Description = in.Description
	c.StartAt = in.StartAt.UTC()
	c.EndAt = in.EndAt.UTC()
	c.Timezone = in.Timezone
	if c.Timezone == "" {
		c.Timezone = s.defaultTimezone
	}
	c.WinRate = in.WinRate
	if in.DailySpins != nil {
		c.DailySpins = *in.DailySpins
	} else if c.ID == 0 {
		c.DailySpins = 1
	}
	c.MaxSpinsPerDay = in.MaxSpinsPerDay
	c.MaxPointsDay = in.MaxPointsDay
	c.BudgetPoints = in.BudgetPoints
	c.ReviewPoints = in.ReviewPoints
	c.SpinRules = append([]model.SpinRule{}, in.SpinRules...)
	c.CheckinRewards = append([]int64{}, in.CheckinRewards...)
	c.Settings = in.Settings
	if c.GameType == model.GameCheckin {
		c.WinRate, c.DailySpins, c.MaxSpinsPerDay, c.MaxPointsDay, c.BudgetPoints = 0, 1, 0, 0, 0
	}
}

func (s *CampaignService) Create(ctx context.Context, actor model.Actor, in port.CampaignInput) (*model.Campaign, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	game, err := model.ParseGame(in.GameType)
	if err != nil {
		return nil, err
	}
	c := &model.Campaign{GameType: game, Status: model.StatusDraft}
	s.build(c, in)
	for _, p := range in.Prizes {
		if p.ID != 0 {
			return nil, &model.ValidationError{Field: "prizes.id", Reason: "must be empty when creating a campaign"}
		}
		c.Prizes = append(c.Prizes, prizeFrom(p))
	}
	c.Normalize()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.campaigns.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func prizeFrom(p port.PrizeInput) model.Prize {
	return model.Prize{
		ID:       p.ID,
		Name:     p.Name,
		Type:     p.Type,
		Points:   p.Points,
		Weight:   p.Weight,
		Quantity: p.Quantity,
		Position: p.Position,
		ImageURL: p.ImageURL,
	}
}

func (s *CampaignService) Get(ctx context.Context, actor model.Actor, id int64) (*model.Campaign, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	return s.campaigns.Get(ctx, id)
}

func (s *CampaignService) List(ctx context.Context, actor model.Actor, in port.ListCampaignsInput) (*port.CampaignPage, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	f := port.CampaignFilter{Status: model.CampaignStatus(in.Status)}
	if f.Status != "" && !f.Status.Valid() {
		return nil, &model.ValidationError{Field: "status", Reason: "must be draft, active or paused"}
	}
	if in.GameType != "" {
		game, err := model.ParseGame(in.GameType)
		if err != nil {
			return nil, err
		}
		f.GameType = game
	}
	f.Limit, f.Offset = page(in.Limit, in.Offset)
	items, total, err := s.campaigns.List(ctx, f)
	if err != nil {
		return nil, err
	}
	return &port.CampaignPage{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

func (s *CampaignService) Update(ctx context.Context, actor model.Actor, id int64, in port.CampaignInput) (*model.Campaign, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	c, err := s.campaigns.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status == model.StatusActive {
		return nil, fmt.Errorf("%w: pause the campaign before editing it", model.ErrConflict)
	}
	if in.GameType != "" && model.GameType(in.GameType) != c.GameType {
		return nil, &model.ValidationError{Field: "game_type", Reason: "cannot be changed, create a new campaign"}
	}
	existing := make(map[int64]model.Prize, len(c.Prizes))
	for _, p := range c.Prizes {
		existing[p.ID] = p
	}
	s.build(c, in)
	c.Prizes = c.Prizes[:0]
	for _, pi := range in.Prizes {
		p := prizeFrom(pi)
		if p.ID != 0 {
			old, ok := existing[p.ID]
			if !ok {
				return nil, &model.ValidationError{Field: "prizes.id", Reason: fmt.Sprintf("prize %d does not belong to this campaign", p.ID)}
			}
			p.Awarded = old.Awarded
		}
		c.Prizes = append(c.Prizes, p)
	}
	c.Normalize()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.campaigns.Update(ctx, c); err != nil {
		return nil, err
	}
	return s.campaigns.Get(ctx, id)
}

func (s *CampaignService) Delete(ctx context.Context, actor model.Actor, id int64) error {
	if err := requireAdmin(actor); err != nil {
		return err
	}
	c, err := s.campaigns.Get(ctx, id)
	if err != nil {
		return err
	}
	if c.Status != model.StatusDraft {
		return fmt.Errorf("%w: only draft campaigns can be deleted", model.ErrConflict)
	}
	return s.campaigns.Delete(ctx, id)
}

func (s *CampaignService) Activate(ctx context.Context, actor model.Actor, id int64) (*model.Campaign, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	c, err := s.campaigns.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status == model.StatusActive {
		return c, nil
	}
	if !c.EndAt.After(s.clock.Now()) {
		return nil, &model.ValidationError{Field: "end_at", Reason: "campaign has already ended"}
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.campaigns.SetStatus(ctx, id, model.StatusActive); err != nil {
		return nil, err
	}
	c.Status = model.StatusActive
	return c, nil
}

func (s *CampaignService) Pause(ctx context.Context, actor model.Actor, id int64) (*model.Campaign, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	c, err := s.campaigns.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	switch c.Status {
	case model.StatusPaused:
		return c, nil
	case model.StatusDraft:
		return nil, fmt.Errorf("%w: a draft campaign cannot be paused", model.ErrConflict)
	}
	if err := s.campaigns.SetStatus(ctx, id, model.StatusPaused); err != nil {
		return nil, err
	}
	c.Status = model.StatusPaused
	return c, nil
}

func (s *CampaignService) Current(ctx context.Context, game string) (*model.Campaign, error) {
	g, err := model.ParseGame(game)
	if err != nil {
		return nil, err
	}
	return running(ctx, s.campaigns, s.clock, g)
}

func (s *CampaignService) Lobby(ctx context.Context) ([]model.Campaign, error) {
	now := s.clock.Now()
	all, err := s.campaigns.Running(ctx, now)
	if err != nil {
		return nil, err
	}
	out := all[:0:0]
	for _, c := range all {
		if c.RunningAt(now) {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(a, b model.Campaign) int {
		return slices.Index(model.Games, a.GameType) - slices.Index(model.Games, b.GameType)
	})
	return out, nil
}

func running(ctx context.Context, campaigns port.CampaignRepository, clock port.Clock, game model.GameType) (*model.Campaign, error) {
	now := clock.Now()
	all, err := campaigns.Running(ctx, now)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].GameType == game && all[i].RunningAt(now) {
			return &all[i], nil
		}
	}
	return nil, model.ErrNoActiveCampaign
}
