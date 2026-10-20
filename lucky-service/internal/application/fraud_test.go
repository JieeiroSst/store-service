package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
)

func (f *fixture) play(t *testing.T, user string, client model.ClientInfo) (*port.PlayResult, error) {
	t.Helper()
	f.rng.rolls = []int{0, 0}
	return f.games.Play(context.Background(), model.Actor{UserID: user}, port.PlayInput{Game: "wheel", Client: client})
}

func (f *fixture) flush(t *testing.T) []model.Reward {
	t.Helper()
	pub := &memPublisher{}
	relay := NewRewardRelay(&config.Config{Relay: config.RelayConfig{Batch: 100}}, f.plays, pub, &fixedClock{t: f.clock.t.Add(time.Hour)})
	if _, err := relay.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	return pub.events
}

func TestDeviceFarming(t *testing.T) {
	f := newFixture(t)
	f.running(t, model.GameWheel)
	phone := model.ClientInfo{DeviceID: "phone-1", IP: "10.0.0.1"}
	for _, u := range []string{"1001", "1002"} {
		if _, err := f.play(t, u, phone); err != nil {
			t.Fatalf("user %s on a shared family phone: %v", u, err)
		}
	}
	if _, err := f.play(t, "1003", phone); !errors.Is(err, model.ErrDeviceLimit) {
		t.Fatalf("third account on the same device must be refused: %v", err)
	}
	if _, err := f.play(t, "1003", model.ClientInfo{DeviceID: "phone-2", IP: "10.0.0.9"}); err != nil {
		t.Fatalf("the same account on its own device plays: %v", err)
	}
	f.signals.fail = true
	if _, err := f.play(t, "1004", phone); err != nil {
		t.Fatalf("a signals outage must fail open: %v", err)
	}
}

func TestRequireDevice(t *testing.T) {
	f := newFixture(t)
	f.running(t, model.GameWheel)
	cfg := &config.Config{Fraud: config.FraudConfig{RequireDevice: true}, RateLimit: config.RateLimitConfig{PlaysPerMinute: 100}}
	guard := NewGuard(GuardDeps{Config: cfg, Fraud: f.fraudRepo})
	games := NewGameService(GameDeps{Config: cfg, Campaigns: f.campaigns, Plays: f.plays, Guard: guard, Clock: f.clock, Random: f.rng})
	if _, err := games.Play(context.Background(), player, port.PlayInput{Game: "wheel"}); !errors.Is(err, model.ErrDeviceRequired) {
		t.Fatalf("missing device id: %v", err)
	}
	if _, err := games.Play(context.Background(), player, port.PlayInput{Game: "wheel", Client: model.ClientInfo{DeviceID: " d "}}); err != nil {
		t.Fatal(err)
	}
}

func TestSharedIPHoldsRewards(t *testing.T) {
	f := newFixture(t)
	f.running(t, model.GameWheel)
	for i, u := range []string{"2001", "2002", "2003"} {
		res, err := f.play(t, u, model.ClientInfo{IP: "203.0.113.7", DeviceID: "dev-" + u})
		if err != nil || !res.Won {
			t.Fatalf("play %d: %+v %v", i, res, err)
		}
	}
	events := f.flush(t)
	if len(events) != 2 {
		t.Fatalf("the reward of the third account on one IP waits for review, published %d", len(events))
	}
	held, err := f.fraud.Rewards(context.Background(), adminActor, port.ListRewardsInput{})
	if err != nil || held.Total != 1 || held.Items[0].UserID != "2003" || held.Items[0].HoldReason != model.HoldSharedIP {
		t.Fatalf("held rewards: %+v %v", held, err)
	}
	if _, err := f.fraud.Approve(context.Background(), player, held.Items[0].ID); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("players cannot approve rewards: %v", err)
	}
	if _, err := f.fraud.Approve(context.Background(), adminActor, held.Items[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.fraud.Approve(context.Background(), adminActor, held.Items[0].ID); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("approving twice: %v", err)
	}
	if events := f.flush(t); len(events) != 1 || events[0].UserID != "2003" {
		t.Fatalf("approved reward is paid: %+v", events)
	}
}

func TestBlockedUser(t *testing.T) {
	f := newFixture(t)
	f.running(t, model.GameWheel)
	f.running(t, model.GameCheckin)
	ctx := context.Background()
	if _, err := f.play(t, "42", model.ClientInfo{}); err != nil {
		t.Fatal(err)
	}
	res, err := f.fraud.Block(ctx, adminActor, port.BlockInput{UserID: "42", Reason: "multi-account farming"})
	if err != nil || res.HeldRewards != 1 || res.Block.BlockedBy != "1" {
		t.Fatalf("block holds unpaid rewards: %+v %v", res, err)
	}
	if events := f.flush(t); len(events) != 0 {
		t.Fatalf("a blocked user's rewards must not be paid: %+v", events)
	}
	if _, err := f.play(t, "42", model.ClientInfo{}); !errors.Is(err, model.ErrBlocked) {
		t.Fatalf("blocked user plays: %v", err)
	}
	if _, err := f.checkins.CheckIn(ctx, player, model.ClientInfo{}); !errors.Is(err, model.ErrBlocked) {
		t.Fatalf("blocked user checks in: %v", err)
	}
	held, _ := f.fraud.Rewards(ctx, adminActor, port.ListRewardsInput{UserID: "42"})
	if _, err := f.fraud.Reject(ctx, adminActor, held.Items[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := f.fraud.Unblock(ctx, adminActor, "42"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.checkins.CheckIn(ctx, player, model.ClientInfo{}); err != nil {
		t.Fatalf("unblocked user checks in: %v", err)
	}
	if events := f.flush(t); len(events) != 1 || events[0].Source != "lucky-checkin" {
		t.Fatalf("rejected reward stays unpaid, new check-in pays: %+v", events)
	}
	if _, err := f.fraud.Block(ctx, adminActor, port.BlockInput{UserID: "42"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("a block needs a reason: %v", err)
	}
}

func TestPointCaps(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	in := f.input(model.GameWheel)
	in.DailySpins = intp(5)
	in.MaxPointsDay = 150
	in.BudgetPoints = 300
	in.ReviewPoints = 1000
	c, err := f.admin.Create(ctx, adminActor, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.admin.Activate(ctx, adminActor, c.ID); err != nil {
		t.Fatal(err)
	}
	won := func(user string) int64 {
		t.Helper()
		res, err := f.play(t, user, model.ClientInfo{})
		if err != nil {
			t.Fatal(err)
		}
		return res.Play.Points
	}
	if won("42") != 100 || won("42") != 0 {
		t.Fatal("the second 100 would pass the 150 daily cap and must turn into a blank")
	}
	f.clock.t = f.clock.t.Add(24 * time.Hour)
	if won("42") != 100 || won("43") != 100 || won("44") != 0 {
		t.Fatal("the campaign budget of 300 must stop the fourth win")
	}
	stored, _ := f.campaigns.Get(ctx, c.ID)
	if stored.PointsAwarded != 300 {
		t.Fatalf("points awarded: %d", stored.PointsAwarded)
	}

	in = f.input(model.GameScratchCard)
	in.ReviewPoints = 1000
	sc, _ := f.admin.Create(ctx, adminActor, in)
	if _, err := f.admin.Activate(ctx, adminActor, sc.ID); err != nil {
		t.Fatal(err)
	}
	f.rng.rolls = []int{0, 1}
	res, err := f.games.Play(ctx, model.Actor{UserID: "50"}, port.PlayInput{Game: "scratch_card"})
	if err != nil || res.Play.Points != 1000 {
		t.Fatalf("jackpot: %+v %v", res, err)
	}
	held, _ := f.fraud.Rewards(ctx, adminActor, port.ListRewardsInput{UserID: "50"})
	if held.Total != 1 || held.Items[0].HoldReason != model.HoldLargeReward {
		t.Fatalf("rewards at or above review_points wait for review: %+v", held)
	}
}

func TestFraudAdminOnly(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if _, err := f.fraud.Block(ctx, player, port.BlockInput{UserID: "1", Reason: "x"}); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("block: %v", err)
	}
	if _, err := f.fraud.Plays(ctx, adminActor, port.ListPlaysInput{}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("investigation needs a filter: %v", err)
	}
	if _, err := f.fraud.Rewards(ctx, adminActor, port.ListRewardsInput{Status: "paid"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("unknown status: %v", err)
	}
}
