package application

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
)

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { return c.t }

type scripted struct{ rolls []int }

func (s *scripted) Intn(n int) int {
	if len(s.rolls) == 0 {
		return n - 1
	}
	v := s.rolls[0] % n
	s.rolls = s.rolls[1:]
	return v
}

type memCampaigns struct {
	mu   sync.Mutex
	next int64
	byID map[int64]*model.Campaign
}

func newMemCampaigns() *memCampaigns { return &memCampaigns{byID: map[int64]*model.Campaign{}} }

func clone(c *model.Campaign) *model.Campaign {
	cp := *c
	cp.Prizes = append([]model.Prize(nil), c.Prizes...)
	cp.SpinRules = append([]model.SpinRule(nil), c.SpinRules...)
	cp.CheckinRewards = append([]int64(nil), c.CheckinRewards...)
	return &cp
}

func (m *memCampaigns) Create(_ context.Context, c *model.Campaign) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	c.ID = m.next
	for i := range c.Prizes {
		c.Prizes[i].ID = m.next*100 + int64(i)
		c.Prizes[i].CampaignID = c.ID
	}
	m.byID[c.ID] = clone(c)
	return nil
}

func (m *memCampaigns) Update(_ context.Context, c *model.Campaign) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byID[c.ID] = clone(c)
	return nil
}

func (m *memCampaigns) Get(_ context.Context, id int64) (*model.Campaign, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.byID[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return clone(c), nil
}

func (m *memCampaigns) List(context.Context, port.CampaignFilter) ([]model.Campaign, int64, error) {
	return nil, 0, nil
}

func (m *memCampaigns) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.byID, id)
	return nil
}

func (m *memCampaigns) SetStatus(_ context.Context, id int64, s model.CampaignStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.byID[id]
	if s == model.StatusActive {
		for _, o := range m.byID {
			if o.ID != id && o.GameType == c.GameType && o.Status == model.StatusActive && o.StartAt.Before(c.EndAt) && c.StartAt.Before(o.EndAt) {
				return model.ErrCampaignOverlap
			}
		}
	}
	c.Status = s
	return nil
}

func (m *memCampaigns) Running(_ context.Context, at time.Time) ([]model.Campaign, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Campaign
	for _, c := range m.byID {
		if c.RunningAt(at) {
			out = append(out, *clone(c))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].GameType != out[j].GameType {
			return out[i].GameType < out[j].GameType
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (m *memCampaigns) claim(prizeID, points int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.byID {
		for i := range c.Prizes {
			if c.Prizes[i].ID == prizeID {
				if !c.Prizes[i].InStock() {
					return false
				}
				if points > 0 && c.BudgetPoints > 0 && c.PointsAwarded+points > c.BudgetPoints {
					return false
				}
				c.Prizes[i].Awarded++
				c.PointsAwarded += points
				return true
			}
		}
	}
	return false
}

type memPlays struct {
	mu        sync.Mutex
	campaigns *memCampaigns
	plays     []model.Play
	grants    []model.CreditGrant
	checkins  []model.Checkin
	outbox    []model.Reward
}

type memLedger struct {
	m          *memPlays
	campaignID int64
	userID     string
}

func (m *memPlays) WithPlayer(_ context.Context, campaignID int64, userID string, fn func(port.PlayerLedger) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(&memLedger{m: m, campaignID: campaignID, userID: userID})
}

func (m *memPlays) usage(campaignID int64, userID, day string) model.Usage {
	var u model.Usage
	used := 0
	for _, p := range m.plays {
		if p.CampaignID != campaignID || p.UserID != userID {
			continue
		}
		if p.Source == model.SourceCredit {
			used++
		}
		if p.Day == day {
			if p.Source == model.SourceDaily {
				u.DailyToday++
			} else {
				u.CreditToday++
			}
		}
	}
	for _, g := range m.grants {
		if g.CampaignID == campaignID && g.UserID == userID {
			u.CreditBalance += g.Spins
		}
	}
	u.CreditBalance -= used
	return u
}

func (m *memPlays) last(campaignID int64, userID string) *model.Checkin {
	var last *model.Checkin
	for i := range m.checkins {
		c := m.checkins[i]
		if c.CampaignID == campaignID && c.UserID == userID && (last == nil || c.Day > last.Day) {
			last = &c
		}
	}
	return last
}

func (m *memPlays) Usage(_ context.Context, campaignID int64, userID, day string) (model.Usage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usage(campaignID, userID, day), nil
}

func (m *memPlays) LastCheckin(_ context.Context, campaignID int64, userID string) (*model.Checkin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.last(campaignID, userID), nil
}

func (m *memPlays) List(context.Context, port.PlayFilter) ([]model.Play, int64, error) {
	return nil, 0, nil
}

func (m *memPlays) Checkins(context.Context, port.CheckinFilter) ([]model.Checkin, int64, error) {
	return nil, 0, nil
}

func (m *memPlays) RecentWins(_ context.Context, campaignID int64, limit int) ([]model.Play, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Play
	for i := len(m.plays) - 1; i >= 0 && len(out) < limit; i-- {
		if p := m.plays[i]; p.CampaignID == campaignID && p.Won() {
			out = append(out, p)
		}
	}
	return out, nil
}

func (m *memPlays) Pending(_ context.Context, before time.Time, limit int) ([]model.Reward, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Reward
	for _, r := range m.outbox {
		if r.PublishedAt == nil && r.Status == model.RewardPending && r.OccurredAt.Before(before) && len(out) < limit {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *memPlays) MarkPublished(_ context.Context, id int64, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.outbox {
		if m.outbox[i].ID == id {
			m.outbox[i].PublishedAt = &at
		}
	}
	return nil
}

func (l *memLedger) PlayByKey(_ context.Context, key string) (*model.Play, error) {
	for _, p := range l.m.plays {
		if p.UserID == l.userID && p.IdempotencyKey == key {
			return &p, nil
		}
	}
	return nil, nil
}

func (l *memLedger) Usage(_ context.Context, day string) (model.Usage, error) {
	return l.m.usage(l.campaignID, l.userID, day), nil
}

func (l *memLedger) ClaimPrize(_ context.Context, prizeID, points int64) (bool, error) {
	return l.m.campaigns.claim(prizeID, points), nil
}

func (l *memLedger) PointsWonOn(_ context.Context, day string) (int64, error) {
	var won int64
	for _, p := range l.m.plays {
		if p.CampaignID == l.campaignID && p.UserID == l.userID && p.Day == day {
			won += p.Points
		}
	}
	return won, nil
}

func (l *memLedger) addOutbox(r model.Reward) {
	r.ID = int64(len(l.m.outbox) + 1)
	r.Status = model.RewardPending
	if r.HoldReason != "" {
		r.Status = model.RewardHeld
	}
	l.m.outbox = append(l.m.outbox, r)
}

func (l *memLedger) SavePlay(_ context.Context, p *model.Play) error {
	p.ID = int64(len(l.m.plays) + 1)
	l.m.plays = append(l.m.plays, *p)
	if p.Won() {
		l.addOutbox(model.PlayReward(*p))
	}
	return nil
}

func (l *memLedger) Grant(_ context.Context, event, ref string) (*model.CreditGrant, error) {
	for _, g := range l.m.grants {
		if g.CampaignID == l.campaignID && g.UserID == l.userID && g.Event == event && g.ReferenceID == ref {
			return &g, nil
		}
	}
	return nil, nil
}

func (l *memLedger) GrantsOn(_ context.Context, event, day string) (int, error) {
	n := 0
	for _, g := range l.m.grants {
		if g.CampaignID == l.campaignID && g.UserID == l.userID && g.Event == event && g.Day == day {
			n++
		}
	}
	return n, nil
}

func (l *memLedger) SaveGrant(_ context.Context, g *model.CreditGrant) error {
	g.ID = int64(len(l.m.grants) + 1)
	l.m.grants = append(l.m.grants, *g)
	return nil
}

func (l *memLedger) LastCheckin(context.Context) (*model.Checkin, error) {
	return l.m.last(l.campaignID, l.userID), nil
}

func (l *memLedger) SaveCheckin(_ context.Context, c *model.Checkin) error {
	c.ID = int64(len(l.m.checkins) + 1)
	l.m.checkins = append(l.m.checkins, *c)
	l.addOutbox(model.CheckinReward(*c))
	return nil
}

type memPublisher struct {
	events []model.Reward
	fail   bool
}

func (p *memPublisher) Enabled() bool { return true }

func (p *memPublisher) PublishReward(_ context.Context, r model.Reward) error {
	if p.fail {
		return model.ErrUnavailable
	}
	p.events = append(p.events, r)
	return nil
}

type memLimiter struct {
	counts map[string]int
}

func (l *memLimiter) Allow(_ context.Context, key string, limit int, _ time.Duration) (bool, error) {
	l.counts[key]++
	return l.counts[key] <= limit, nil
}

type memWinners struct {
	items []model.Winner
}

func (w *memWinners) Push(_ context.Context, win model.Winner) error {
	w.items = append([]model.Winner{win}, w.items...)
	return nil
}

func (w *memWinners) Recent(_ context.Context, campaignID int64, limit int) ([]model.Winner, error) {
	var out []model.Winner
	for _, it := range w.items {
		if it.CampaignID == campaignID && len(out) < limit {
			out = append(out, it)
		}
	}
	return out, nil
}

type memFraud struct {
	blocks map[string]model.UserBlock
	plays  *memPlays
}

func (f *memFraud) BlockOf(_ context.Context, userID string) (*model.UserBlock, error) {
	if b, ok := f.blocks[userID]; ok {
		return &b, nil
	}
	return nil, nil
}

func (f *memFraud) Block(_ context.Context, b *model.UserBlock) (int64, error) {
	f.blocks[b.UserID] = *b
	f.plays.mu.Lock()
	defer f.plays.mu.Unlock()
	var held int64
	for i := range f.plays.outbox {
		r := &f.plays.outbox[i]
		if r.UserID == b.UserID && r.Status == model.RewardPending && r.PublishedAt == nil {
			r.Status, r.HoldReason = model.RewardHeld, model.HoldBlockedUser
			held++
		}
	}
	return held, nil
}

func (f *memFraud) Unblock(_ context.Context, userID string) error {
	if _, ok := f.blocks[userID]; !ok {
		return model.ErrNotFound
	}
	delete(f.blocks, userID)
	return nil
}

func (f *memFraud) Blocks(context.Context, int, int) ([]model.UserBlock, int64, error) {
	return nil, int64(len(f.blocks)), nil
}

func (f *memFraud) Rewards(_ context.Context, rf port.RewardFilter) ([]model.ReviewReward, int64, error) {
	var out []model.ReviewReward
	for _, r := range f.plays.outbox {
		if r.Status == rf.Status && (rf.UserID == "" || r.UserID == rf.UserID) {
			out = append(out, model.ReviewReward{ID: r.ID, UserID: r.UserID, Points: r.Points, Status: r.Status, HoldReason: r.HoldReason})
		}
	}
	return out, int64(len(out)), nil
}

func (f *memFraud) Review(_ context.Context, id int64, from []model.RewardStatus, to model.RewardStatus, by string) (*model.ReviewReward, error) {
	for i := range f.plays.outbox {
		r := &f.plays.outbox[i]
		if r.ID != id {
			continue
		}
		ok := false
		for _, s := range from {
			ok = ok || r.Status == s
		}
		if !ok || r.PublishedAt != nil {
			return nil, model.ErrConflict
		}
		r.Status = to
		return &model.ReviewReward{ID: r.ID, Status: to, ReviewedBy: by}, nil
	}
	return nil, model.ErrNotFound
}

type memSignals struct {
	seen map[string]map[string]bool
	fail bool
}

func (s *memSignals) Observe(_ context.Context, kind model.SignalKind, value, userID string) (int, error) {
	if s.fail {
		return 0, model.ErrUnavailable
	}
	k := string(kind) + ":" + value
	if s.seen[k] == nil {
		s.seen[k] = map[string]bool{}
	}
	s.seen[k][userID] = true
	return len(s.seen[k]), nil
}
