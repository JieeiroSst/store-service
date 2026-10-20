package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/application"
	"github.com/JIeeiroSst/lucky-service/internal/config"
	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func testDB(t *testing.T) *gorm.DB {
	host := os.Getenv("LUCKY_TEST_DB_HOST")
	if host == "" {
		t.Skip("LUCKY_TEST_DB_HOST not set")
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, env("LUCKY_TEST_DB_PORT", "5432"), env("LUCKY_TEST_DB_USER", "postgres"),
		os.Getenv("LUCKY_TEST_DB_PASSWORD"), env("LUCKY_TEST_DB_NAME", "postgres"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true, Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("DROP TABLE IF EXISTS lucky_user_blocks, lucky_reward_outbox, lucky_checkins, lucky_spin_credits, lucky_plays, lucky_prizes, lucky_campaigns").Error; err != nil {
		t.Fatal(err)
	}
	r := NewRepository(db)
	var wg sync.WaitGroup
	errs := make([]error, 3)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = r.Migrate(context.Background())
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatalf("concurrent migrate: %v", err)
		}
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	return db
}

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) add(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

type alwaysWin struct{}

func (alwaysWin) Intn(int) int { return 0 }

type harness struct {
	campaigns *Repository
	plays     *PlayRepository
	fraudRepo *FraudRepository
	fraud     port.FraudUsecase
	clock     *clock
	admin     port.CampaignUsecase
	games     port.GameUsecase
	checkins  port.CheckinUsecase
}

var (
	admin = model.Actor{UserID: "1", Admin: true}
	svc   = model.Actor{Service: true}
)

func newHarness(t *testing.T) *harness {
	db := testDB(t)
	h := &harness{
		campaigns: NewRepository(db),
		plays:     NewPlayRepository(db),
		fraudRepo: NewFraudRepository(db),
		clock:     &clock{t: time.Now().UTC().Truncate(time.Second)},
	}
	cfg := &config.Config{DefaultTimezone: "Asia/Ho_Chi_Minh"}
	guard := application.NewGuard(application.GuardDeps{Config: cfg, Fraud: h.fraudRepo})
	h.admin = application.NewCampaignService(cfg, h.campaigns, h.clock)
	h.games = application.NewGameService(application.GameDeps{Config: cfg, Campaigns: h.campaigns, Plays: h.plays, Guard: guard, Clock: h.clock, Random: alwaysWin{}})
	h.checkins = application.NewCheckinService(application.CheckinDeps{Config: cfg, Campaigns: h.campaigns, Plays: h.plays, Guard: guard, Clock: h.clock})
	h.fraud = application.NewFraudService(h.fraudRepo, h.plays)
	return h
}

func (h *harness) input(game model.GameType, stock int64) port.CampaignInput {
	daily := 1
	in := port.CampaignInput{
		GameType:   string(game),
		Name:       "Lucky " + string(game),
		StartAt:    h.clock.Now().Add(-time.Hour),
		EndAt:      h.clock.Now().Add(7 * 24 * time.Hour),
		WinRate:    100,
		DailySpins: &daily,
		SpinRules:  []model.SpinRule{{Event: "order_completed", Spins: 1}},
		Settings:   json.RawMessage(`{"boxes":6}`),
		Prizes: []port.PrizeInput{
			{Name: "Jackpot", Type: model.PrizePoints, Points: 500, Weight: 1, Quantity: &stock},
			{Name: "Try again", Type: model.PrizeNone},
		},
	}
	if game == model.GameCheckin {
		in.SpinRules, in.Prizes = nil, nil
		in.CheckinRewards = []int64{5, 10, 15}
	}
	return in
}

func TestCampaignLifecycle(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	c, err := h.admin.Create(ctx, admin, h.input(model.GameWheel, 5))
	if err != nil {
		t.Fatal(err)
	}
	if c.ID == 0 || len(c.Prizes) != 2 || c.Prizes[0].ID == 0 || c.CreatedAt.IsZero() || c.SpinRules[0].Event != "order_completed" || string(c.Settings) != `{"boxes": 6}` {
		t.Fatalf("created: %+v settings=%s", c, c.Settings)
	}

	in := h.input(model.GameWheel, 5)
	in.Name = "Wheel v2"
	in.Prizes = []port.PrizeInput{
		{ID: c.Prizes[1].ID, Name: "Try again", Type: model.PrizeNone, Position: 1},
		{ID: c.Prizes[0].ID, Name: "Jackpot", Type: model.PrizePoints, Points: 900, Weight: 1, Position: 2},
		{Name: "Small", Type: model.PrizePoints, Points: 10, Weight: 5, Position: 3},
	}
	upd, err := h.admin.Update(ctx, admin, c.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	if upd.Name != "Wheel v2" || len(upd.Prizes) != 3 || upd.Prizes[0].Type != model.PrizeNone || upd.Prizes[1].Points != 900 || upd.Prizes[1].Quantity != nil {
		t.Fatalf("swapped positions and new prize: %+v", upd.Prizes)
	}

	if _, err := h.admin.Activate(ctx, admin, c.ID); err != nil {
		t.Fatal(err)
	}
	other, err := h.admin.Create(ctx, admin, h.input(model.GameWheel, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.admin.Activate(ctx, admin, other.ID); !errors.Is(err, model.ErrCampaignOverlap) {
		t.Fatalf("exclusion constraint must stop a second active wheel: %v", err)
	}
	scratch, err := h.admin.Create(ctx, admin, h.input(model.GameScratchCard, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.admin.Activate(ctx, admin, scratch.ID); err != nil {
		t.Fatalf("different games run side by side: %v", err)
	}
	checkin, err := h.admin.Create(ctx, admin, h.input(model.GameCheckin, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.admin.Activate(ctx, admin, checkin.ID); err != nil {
		t.Fatal(err)
	}
	later := h.input(model.GameWheel, 1)
	later.StartAt = h.clock.Now().Add(7 * 24 * time.Hour)
	later.EndAt = later.StartAt.Add(24 * time.Hour)
	next, err := h.admin.Create(ctx, admin, later)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.admin.Activate(ctx, admin, next.ID); err != nil {
		t.Fatalf("back-to-back windows do not overlap: %v", err)
	}
	lobby, err := h.admin.Lobby(ctx)
	if err != nil || len(lobby) != 3 || lobby[0].ID != c.ID || lobby[2].GameType != model.GameCheckin || lobby[2].CheckinRewards[1] != 10 {
		t.Fatalf("lobby: %+v %v", lobby, err)
	}
	page, err := h.admin.List(ctx, admin, port.ListCampaignsInput{GameType: "wheel", Status: "active"})
	if err != nil || page.Total != 2 {
		t.Fatalf("list active wheels: %+v %v", page, err)
	}
	if err := h.admin.Delete(ctx, admin, other.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.campaigns.Get(ctx, other.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("deleted: %v", err)
	}
}

func TestConcurrentPlays(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	c, err := h.admin.Create(ctx, admin, h.input(model.GameFlipCard, 3))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.admin.Activate(ctx, admin, c.ID); err != nil {
		t.Fatal(err)
	}

	const users, attempts = 6, 4
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, refused := 0, 0
	for u := range users {
		for range attempts {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := h.games.Play(ctx, model.Actor{UserID: fmt.Sprint(100 + u)}, port.PlayInput{Game: "flip_card", Choice: 2})
				mu.Lock()
				defer mu.Unlock()
				switch {
				case err == nil:
					ok++
				case errors.Is(err, model.ErrNoSpinsLeft):
					refused++
				default:
					t.Errorf("play: %v", err)
				}
			}()
		}
	}
	wg.Wait()
	if ok != users || refused != users*(attempts-1) {
		t.Fatalf("each user gets exactly one daily play: ok=%d refused=%d", ok, refused)
	}
	stored, _ := h.campaigns.Get(ctx, c.ID)
	if stored.Prizes[0].Awarded != 3 || stored.Prizes[1].Awarded != 3 {
		t.Fatalf("stock of 3 must never be oversold: %+v", stored.Prizes)
	}

	player := model.Actor{UserID: "100"}
	g, err := h.games.GrantPlays(ctx, svc, port.GrantInput{UserID: "100", Event: "order_completed", ReferenceID: "o-1"})
	if err != nil || len(g.Grants) != 1 || g.Grants[0].GameType != model.GameFlipCard {
		t.Fatalf("grant: %+v %v", g, err)
	}
	st, err := h.games.Status(ctx, player, "flip_card")
	if err != nil || st.Remaining != 1 || st.Usage.CreditBalance != 1 || st.Usage.DailyToday != 1 {
		t.Fatalf("status after grant: %+v %v", st, err)
	}
	first, err := h.games.Play(ctx, player, port.PlayInput{Game: "flip_card", Choice: 5, IdempotencyKey: "retry-1"})
	if err != nil || first.Play.Source != model.SourceCredit || first.Play.Choice != 5 {
		t.Fatalf("credit play: %+v %v", first, err)
	}
	replay, err := h.games.Play(ctx, player, port.PlayInput{Game: "flip_card", Choice: 5, IdempotencyKey: "retry-1"})
	if err != nil || !replay.Replayed || replay.Play.ID != first.Play.ID {
		t.Fatalf("idempotent replay: %+v %v", replay, err)
	}

	h.clock.add(24 * time.Hour)
	if _, err := h.games.Play(ctx, player, port.PlayInput{Game: "flip_card", Choice: 1}); err != nil {
		t.Fatalf("a new day brings a new daily play: %v", err)
	}
	hist, err := h.games.History(ctx, player, port.ListPlaysInput{GameType: "flip_card"})
	if err != nil || hist.Total != 3 || hist.Items[0].Day == hist.Items[2].Day {
		t.Fatalf("history: %+v %v", hist, err)
	}
	wins, err := h.games.Winners(ctx, "flip_card", 10)
	if err != nil || len(wins) != 3 || wins[0].User[:5] != "10***" {
		t.Fatalf("winners from the database: %+v %v", wins, err)
	}

	pending, err := h.plays.Pending(ctx, h.clock.Now().Add(time.Hour), 100)
	if err != nil || len(pending) != 3 || pending[0].Points != 500 || pending[0].Source != "lucky-flip_card" {
		t.Fatalf("three jackpots wait in the outbox: %+v %v", pending, err)
	}
	if err := h.plays.MarkPublished(ctx, pending[0].ID, h.clock.Now()); err != nil {
		t.Fatal(err)
	}
	pending, _ = h.plays.Pending(ctx, h.clock.Now().Add(time.Hour), 100)
	if len(pending) != 2 {
		t.Fatalf("published reward must leave the outbox: %d", len(pending))
	}

	if _, err := h.admin.Pause(ctx, admin, c.ID); err != nil {
		t.Fatal(err)
	}
	in := h.input(model.GameFlipCard, 3)
	in.Prizes = []port.PrizeInput{
		{ID: stored.Prizes[1].ID, Name: "Try again", Type: model.PrizeNone},
		{Name: "New", Type: model.PrizePoints, Points: 1, Weight: 1},
	}
	if _, err := h.admin.Update(ctx, admin, c.ID, in); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("removing a prize that was already won must be refused: %v", err)
	}
	in = h.input(model.GameFlipCard, 1)
	in.Prizes[0].ID, in.Prizes[1].ID = stored.Prizes[0].ID, stored.Prizes[1].ID
	if _, err := h.admin.Update(ctx, admin, c.ID, in); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("stock below awarded must be refused: %v", err)
	}
}

func TestCheckins(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	c, err := h.admin.Create(ctx, admin, h.input(model.GameCheckin, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.admin.Activate(ctx, admin, c.ID); err != nil {
		t.Fatal(err)
	}
	player := model.Actor{UserID: "55"}
	var wg sync.WaitGroup
	var mu sync.Mutex
	created, dup := 0, 0
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.checkins.CheckIn(ctx, player, model.ClientInfo{})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				created++
			case errors.Is(err, model.ErrAlreadyCheckedIn):
				dup++
			default:
				t.Errorf("check-in: %v", err)
			}
		}()
	}
	wg.Wait()
	if created != 1 || dup != 4 {
		t.Fatalf("one check-in per day under concurrency: created=%d dup=%d", created, dup)
	}
	for _, want := range []int64{10, 15, 5} {
		h.clock.add(24 * time.Hour)
		res, err := h.checkins.CheckIn(ctx, player, model.ClientInfo{})
		if err != nil || res.Checkin.Points != want {
			t.Fatalf("streak reward: %+v %v", res, err)
		}
	}
	st, err := h.checkins.Status(ctx, player)
	if err != nil || !st.CheckedInToday || st.Streak != 4 || st.NextReward != 10 || st.CycleDay != 2 {
		t.Fatalf("status: %+v %v", st, err)
	}
	h.clock.add(48 * time.Hour)
	res, err := h.checkins.CheckIn(ctx, player, model.ClientInfo{})
	if err != nil || res.Checkin.Streak != 1 || res.Checkin.Points != 5 {
		t.Fatalf("missed day resets: %+v %v", res, err)
	}
	hist, err := h.checkins.History(ctx, player, 0, 0)
	if err != nil || hist.Total != 5 || hist.Items[0].Streak != 1 {
		t.Fatalf("history: %+v %v", hist, err)
	}
	pending, err := h.plays.Pending(ctx, h.clock.Now().Add(time.Hour), 100)
	if err != nil || len(pending) != 5 || pending[0].EventID != fmt.Sprintf("lucky-checkin-%d", hist.Items[4].ID) {
		t.Fatalf("every check-in pays out: %+v %v", pending, err)
	}
}

func TestFraudControls(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	in := h.input(model.GameWheel, 2)
	daily := 10
	in.DailySpins = &daily
	in.BudgetPoints = 1000
	in.ReviewPoints = 500
	c, err := h.admin.Create(ctx, admin, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.admin.Activate(ctx, admin, c.ID); err != nil {
		t.Fatal(err)
	}
	client := model.ClientInfo{IP: "198.51.100.4", DeviceID: "device-abc", UserAgent: "LuckyApp/1.0"}
	for _, u := range []string{"300", "301", "302"} {
		res, err := h.games.Play(ctx, model.Actor{UserID: u}, port.PlayInput{Game: "wheel", Client: client})
		if err != nil {
			t.Fatal(err)
		}
		if want := map[string]int64{"300": 500, "301": 500, "302": 0}[u]; res.Play.Points != want {
			t.Fatalf("user %s won %d, want %d", u, res.Play.Points, want)
		}
	}
	stored, _ := h.campaigns.Get(ctx, c.ID)
	if stored.PointsAwarded != 1000 || stored.Prizes[0].Awarded != 2 {
		t.Fatalf("budget and stock: %+v", stored)
	}

	in = h.input(model.GameScratchCard, 5)
	in.DailySpins = &daily
	in.BudgetPoints = 500
	sc, _ := h.admin.Create(ctx, admin, in)
	if _, err := h.admin.Activate(ctx, admin, sc.ID); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := h.games.Play(ctx, model.Actor{UserID: "400"}, port.PlayInput{Game: "scratch_card"}); err != nil {
			t.Fatal(err)
		}
	}
	scStored, _ := h.campaigns.Get(ctx, sc.ID)
	if scStored.Prizes[0].Awarded != 1 || scStored.PointsAwarded != 500 {
		t.Fatalf("a budget refusal must roll the stock claim back: %+v", scStored.Prizes)
	}

	plays, err := h.fraud.Plays(ctx, admin, port.ListPlaysInput{DeviceID: "device-abc"})
	if err != nil || plays.Total != 3 || plays.Items[0].Client.UserAgent != "LuckyApp/1.0" || plays.Items[0].Client.IP != "198.51.100.4" {
		t.Fatalf("plays by device: %+v %v", plays, err)
	}

	held, err := h.fraud.Rewards(ctx, admin, port.ListRewardsInput{})
	if err != nil || held.Total != 2 || held.Items[0].HoldReason != model.HoldLargeReward {
		t.Fatalf("rewards >= review_points are held: %+v %v", held, err)
	}
	pending, _ := h.plays.Pending(ctx, h.clock.Now().Add(time.Hour), 100)
	if len(pending) != 1 || pending[0].CampaignID != sc.ID {
		t.Fatalf("held rewards must not reach the relay: %+v", pending)
	}
	approved, err := h.fraud.Approve(ctx, admin, held.Items[0].ID)
	if err != nil || approved.Status != model.RewardPending || approved.ReviewedBy != "1" {
		t.Fatalf("approve: %+v %v", approved, err)
	}
	pending, _ = h.plays.Pending(ctx, h.clock.Now().Add(time.Hour), 100)
	if len(pending) != 2 || pending[0].ID != held.Items[0].ID || pending[0].HoldReason != model.HoldLargeReward {
		t.Fatalf("approved reward goes to the relay: %+v", pending)
	}
	if err := h.plays.MarkPublished(ctx, pending[0].ID, h.clock.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := h.fraud.Reject(ctx, admin, pending[0].ID); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("a paid reward cannot be rejected: %v", err)
	}
	if _, err := h.fraud.Reject(ctx, admin, held.Items[1].ID); err != nil {
		t.Fatal(err)
	}
	if refunded, _ := h.campaigns.Get(ctx, c.ID); refunded.PointsAwarded != 500 {
		t.Fatalf("a rejected reward returns its points to the budget: %d", refunded.PointsAwarded)
	}

	blocked, err := h.fraud.Block(ctx, admin, port.BlockInput{UserID: "400", Reason: "bot pattern"})
	if err != nil || blocked.HeldRewards != 1 {
		t.Fatalf("blocking holds the unpaid scratch card reward: %+v %v", blocked, err)
	}
	if _, err := h.games.Play(ctx, model.Actor{UserID: "400"}, port.PlayInput{Game: "scratch_card"}); !errors.Is(err, model.ErrBlocked) {
		t.Fatalf("blocked user plays: %v", err)
	}
	again, err := h.fraud.Block(ctx, admin, port.BlockInput{UserID: "400", Reason: "confirmed"})
	if err != nil || again.Block.Reason != "confirmed" {
		t.Fatalf("re-blocking updates the reason: %+v %v", again, err)
	}
	list, err := h.fraud.Blocks(ctx, admin, 0, 0)
	if err != nil || list.Total != 1 {
		t.Fatalf("blocks: %+v %v", list, err)
	}
	if err := h.fraud.Unblock(ctx, admin, "400"); err != nil {
		t.Fatal(err)
	}
	if err := h.fraud.Unblock(ctx, admin, "400"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("unblock twice: %v", err)
	}
}
