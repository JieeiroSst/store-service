package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
)

type fixture struct {
	clock     *fixedClock
	rng       *scripted
	campaigns *memCampaigns
	plays     *memPlays
	limiter   *memLimiter
	winners   *memWinners
	fraudRepo *memFraud
	signals   *memSignals
	fraud     port.FraudUsecase
	admin     port.CampaignUsecase
	games     port.GameUsecase
	checkins  port.CheckinUsecase
}

var (
	adminActor = model.Actor{UserID: "1", Admin: true}
	service    = model.Actor{Service: true}
	player     = model.Actor{UserID: "42"}
)

func intp(n int) *int { return &n }

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		clock:     &fixedClock{t: time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC)},
		rng:       &scripted{},
		campaigns: newMemCampaigns(),
		limiter:   &memLimiter{counts: map[string]int{}},
		winners:   &memWinners{},
	}
	f.plays = &memPlays{campaigns: f.campaigns}
	f.fraudRepo = &memFraud{blocks: map[string]model.UserBlock{}, plays: f.plays}
	f.signals = &memSignals{seen: map[string]map[string]bool{}}
	cfg := &config.Config{
		DefaultTimezone: "Asia/Ho_Chi_Minh",
		RateLimit:       config.RateLimitConfig{PlaysPerMinute: 100},
		Fraud:           config.FraudConfig{MaxUsersPerDevice: 2, MaxUsersPerIP: 2},
	}
	guard := NewGuard(GuardDeps{Config: cfg, Fraud: f.fraudRepo, Signals: f.signals})
	f.admin = NewCampaignService(cfg, f.campaigns, f.clock)
	f.games = NewGameService(GameDeps{Config: cfg, Campaigns: f.campaigns, Plays: f.plays, Limiter: f.limiter, Winners: f.winners, Guard: guard, Clock: f.clock, Random: f.rng})
	f.checkins = NewCheckinService(CheckinDeps{Config: cfg, Campaigns: f.campaigns, Plays: f.plays, Limiter: f.limiter, Guard: guard, Clock: f.clock})
	f.fraud = NewFraudService(f.fraudRepo, f.plays)
	return f
}

func (f *fixture) input(game model.GameType) port.CampaignInput {
	stock := int64(1)
	in := port.CampaignInput{
		GameType:   string(game),
		Name:       "October " + string(game),
		StartAt:    f.clock.t.Add(-time.Hour),
		EndAt:      f.clock.t.Add(10 * 24 * time.Hour),
		WinRate:    50,
		DailySpins: intp(1),
		SpinRules:  []model.SpinRule{{Event: "order_completed", Spins: 2, MaxPerDay: 1}},
		Settings:   json.RawMessage(`{"theme":"tet"}`),
		Prizes: []port.PrizeInput{
			{Name: "100 points", Type: model.PrizePoints, Points: 100, Weight: 1},
			{Name: "1000 points", Type: model.PrizePoints, Points: 1000, Weight: 1, Quantity: &stock},
			{Name: "Good luck next time", Type: model.PrizeNone},
		},
	}
	if game == model.GameCheckin {
		in.SpinRules, in.Prizes, in.WinRate = nil, nil, 0
		in.CheckinRewards = []int64{10, 20, 30}
	}
	return in
}

func (f *fixture) running(t *testing.T, game model.GameType) *model.Campaign {
	t.Helper()
	ctx := context.Background()
	c, err := f.admin.Create(ctx, adminActor, f.input(game))
	if err != nil {
		t.Fatal(err)
	}
	if c.Timezone != "Asia/Ho_Chi_Minh" || c.Status != model.StatusDraft || c.GameType != game {
		t.Fatalf("defaults not applied: %+v", c)
	}
	if _, err := f.admin.Activate(ctx, adminActor, c.ID); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCampaignAdmin(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if _, err := f.admin.Create(ctx, player, f.input(model.GameWheel)); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("players must not create campaigns: %v", err)
	}
	in := f.input(model.GameWheel)
	in.GameType = "poker"
	if _, err := f.admin.Create(ctx, adminActor, in); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("unknown game: %v", err)
	}
	wheel := f.running(t, model.GameWheel)
	f.running(t, model.GameScratchCard)
	f.running(t, model.GameCheckin)
	other, err := f.admin.Create(ctx, adminActor, f.input(model.GameWheel))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.admin.Activate(ctx, adminActor, other.ID); !errors.Is(err, model.ErrCampaignOverlap) {
		t.Fatalf("a second wheel must not run at the same time: %v", err)
	}
	if _, err := f.admin.Update(ctx, adminActor, wheel.ID, f.input(model.GameWheel)); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("active campaigns must be paused before editing: %v", err)
	}
	if _, err := f.admin.Pause(ctx, adminActor, wheel.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.admin.Update(ctx, adminActor, wheel.ID, f.input(model.GameShake)); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("game type is immutable: %v", err)
	}
	lobby, err := f.admin.Lobby(ctx)
	if err != nil || len(lobby) != 2 || lobby[0].GameType != model.GameScratchCard || lobby[1].GameType != model.GameCheckin {
		t.Fatalf("lobby lists running games in catalog order: %+v %v", lobby, err)
	}
	if _, err := f.admin.Current(ctx, "wheel"); !errors.Is(err, model.ErrNoActiveCampaign) {
		t.Fatalf("paused wheel is not current: %v", err)
	}
}

func TestPlayQuotaAndCredits(t *testing.T) {
	f := newFixture(t)
	c := f.running(t, model.GameWheel)
	scratch := f.running(t, model.GameScratchCard)
	ctx := context.Background()

	f.rng.rolls = []int{0, 0}
	res, err := f.games.Play(ctx, player, port.PlayInput{Game: "wheel"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Won || res.Play.Points != 100 || res.Play.Source != model.SourceDaily || res.Status.Remaining != 0 || res.Play.GameType != model.GameWheel {
		t.Fatalf("first play: %+v", res)
	}
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrNoSpinsLeft) {
		t.Fatalf("second play must be refused: %v", err)
	}
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "scratch_card"}); err != nil {
		t.Fatalf("each game has its own daily quota: %v", err)
	}

	g, err := f.games.GrantPlays(ctx, service, port.GrantInput{UserID: "42", Event: "ORDER_COMPLETED", ReferenceID: "order-1"})
	if err != nil || len(g.Grants) != 2 {
		t.Fatalf("one event grants plays in every running game with the rule: %+v %v", g, err)
	}
	again, err := f.games.GrantPlays(ctx, service, port.GrantInput{UserID: "42", Event: "order_completed", ReferenceID: "order-1"})
	if err != nil || len(again.Grants) != 2 || again.Grants[0].ID != g.Grants[0].ID {
		t.Fatalf("same reference is idempotent: %+v %v", again, err)
	}
	limited, err := f.games.GrantPlays(ctx, service, port.GrantInput{UserID: "42", Event: "order_completed", ReferenceID: "order-2"})
	if err != nil || len(limited.Grants) != 0 || len(limited.Skips) != 2 || limited.Skips[0].Reason != "credit_limit_reached" {
		t.Fatalf("max_per_day caps grants: %+v %v", limited, err)
	}
	if _, err := f.games.GrantPlays(ctx, player, port.GrantInput{UserID: "42", Event: "order_completed", ReferenceID: "x"}); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("players must not grant themselves plays: %v", err)
	}
	if _, err := f.games.GrantPlays(ctx, service, port.GrantInput{UserID: "42", Event: "unknown", ReferenceID: "x"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("unknown event: %v", err)
	}

	f.rng.rolls = []int{0, 1}
	res, err = f.games.Play(ctx, player, port.PlayInput{Game: "wheel", IdempotencyKey: "k1"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Play.Points != 1000 || res.Play.Source != model.SourceCredit || res.Status.Remaining != 1 {
		t.Fatalf("credit play: %+v", res)
	}
	replay, err := f.games.Play(ctx, player, port.PlayInput{Game: "wheel", IdempotencyKey: "k1"})
	if err != nil || !replay.Replayed || replay.Play.ID != res.Play.ID {
		t.Fatalf("idempotent replay: %+v %v", replay, err)
	}

	f.rng.rolls = []int{0, 1}
	res, err = f.games.Play(ctx, player, port.PlayInput{Game: "wheel"})
	if err != nil || res.Play.Points != 100 {
		t.Fatalf("sold out prize must not be drawn again: %+v %v", res, err)
	}
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrNoSpinsLeft) {
		t.Fatalf("credits exhausted: %v", err)
	}

	f.clock.t = f.clock.t.Add(24 * time.Hour)
	f.rng.rolls = []int{9999}
	res, err = f.games.Play(ctx, player, port.PlayInput{Game: "wheel"})
	if err != nil || res.Won || res.Play.Day != "2026-10-21" {
		t.Fatalf("next day losing play: %+v %v", res, err)
	}
	stored, _ := f.campaigns.Get(ctx, c.ID)
	if stored.Prizes[1].Awarded != 1 || stored.Prizes[2].Awarded != 1 {
		t.Fatalf("awarded counters: %+v", stored.Prizes)
	}

	winners, err := f.games.Winners(ctx, "wheel", 10)
	if err != nil || len(winners) != 3 || winners[0].Points != 100 || winners[1].Points != 1000 || winners[0].User != "4***" {
		t.Fatalf("winner feed newest first, masked: %+v %v", winners, err)
	}
	f.winners.items = nil
	fallback, err := f.games.Winners(ctx, "wheel", 10)
	if err != nil || len(fallback) != 3 {
		t.Fatalf("winners fall back to the database: %+v %v", fallback, err)
	}
	_ = scratch
}

func TestChoiceGames(t *testing.T) {
	f := newFixture(t)
	f.running(t, model.GameGiftBox)
	ctx := context.Background()
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "gift_box"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("gift box needs a box: %v", err)
	}
	res, err := f.games.Play(ctx, player, port.PlayInput{Game: "gift_box", Choice: 3})
	if err != nil || res.Play.Choice != 3 || res.Play.GameType != model.GameGiftBox {
		t.Fatalf("gift box play: %+v %v", res, err)
	}
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "checkin"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("checkin is not played through the instant-win endpoint: %v", err)
	}
}

func TestPlayGuards(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrNoActiveCampaign) {
		t.Fatalf("no campaign: %v", err)
	}
	f.running(t, model.GameWheel)
	if _, err := f.games.Play(ctx, service, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("service tokens cannot play: %v", err)
	}
	if _, err := f.games.Play(ctx, model.Actor{}, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrUnauthenticated) {
		t.Fatalf("anonymous: %v", err)
	}
	f.limiter.counts["play:42"] = 100
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrRateLimited) {
		t.Fatalf("rate limited: %v", err)
	}
	f.limiter.counts["play:42"] = 0
	f.clock.t = f.clock.t.Add(11 * 24 * time.Hour)
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrNoActiveCampaign) {
		t.Fatalf("ended campaign: %v", err)
	}
}

func TestCheckinStreak(t *testing.T) {
	f := newFixture(t)
	f.running(t, model.GameCheckin)
	ctx := context.Background()

	st, err := f.checkins.Status(ctx, player)
	if err != nil || st.CheckedInToday || st.Streak != 0 || st.NextReward != 10 || st.CycleDay != 1 {
		t.Fatalf("fresh status: %+v %v", st, err)
	}
	day := func(want int64, streak int) {
		t.Helper()
		res, err := f.checkins.CheckIn(ctx, player, model.ClientInfo{})
		if err != nil {
			t.Fatal(err)
		}
		if res.Checkin.Points != want || res.Checkin.Streak != streak || !res.Status.CheckedInToday {
			t.Fatalf("check-in: %+v", res)
		}
		f.clock.t = f.clock.t.Add(24 * time.Hour)
	}
	day(10, 1)
	day(20, 2)
	f.clock.t = f.clock.t.Add(-24 * time.Hour)
	if _, err := f.checkins.CheckIn(ctx, player, model.ClientInfo{}); !errors.Is(err, model.ErrAlreadyCheckedIn) {
		t.Fatalf("one check-in per day: %v", err)
	}
	f.clock.t = f.clock.t.Add(24 * time.Hour)
	day(30, 3)
	day(10, 4)
	f.clock.t = f.clock.t.Add(24 * time.Hour)
	st, _ = f.checkins.Status(ctx, player)
	if st.Streak != 0 || st.NextReward != 10 {
		t.Fatalf("a missed day resets the streak: %+v", st)
	}
	day(10, 1)

	pub := &memPublisher{}
	relay := NewRewardRelay(&config.Config{Relay: config.RelayConfig{Grace: time.Second, Batch: 50}}, f.plays, pub, f.clock)
	n, err := relay.Flush(ctx)
	if err != nil || n != 5 || pub.events[0].EventID != "lucky-checkin-1" || pub.events[0].Source != "lucky-checkin" {
		t.Fatalf("check-ins pay out through the outbox: %d %+v %v", n, pub.events, err)
	}
}

func TestRelay(t *testing.T) {
	f := newFixture(t)
	f.running(t, model.GameShake)
	ctx := context.Background()
	f.rng.rolls = []int{0, 0}
	if _, err := f.games.Play(ctx, player, port.PlayInput{Game: "shake"}); err != nil {
		t.Fatal(err)
	}
	pub := &memPublisher{fail: true}
	relay := NewRewardRelay(&config.Config{Relay: config.RelayConfig{Grace: time.Second, Batch: 10}}, f.plays, pub, f.clock)
	if n, _ := relay.Flush(ctx); n != 0 {
		t.Fatalf("rewards inside the grace period must wait")
	}
	f.clock.t = f.clock.t.Add(time.Minute)
	if _, err := relay.Flush(ctx); err == nil {
		t.Fatal("publisher failure must surface")
	}
	pub.fail = false
	n, err := relay.Flush(ctx)
	if err != nil || n != 1 {
		t.Fatalf("flush: %d %v", n, err)
	}
	ev := pub.events[0]
	if ev.EventID != "lucky-play-1" || ev.UserID != "42" || ev.Points != 100 || ev.Source != "lucky-shake" || ev.GameType != model.GameShake {
		t.Fatalf("event: %+v", ev)
	}
	if n, _ := relay.Flush(ctx); n != 0 {
		t.Fatalf("published rewards must not be sent twice")
	}
}
