package model

import (
	"errors"
	"testing"
	"time"
)

func qty(n int64) *int64 { return &n }

func validCampaign() Campaign {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	return Campaign{
		GameType:   GameWheel,
		Name:       " Wheel ",
		StartAt:    start,
		EndAt:      start.Add(30 * 24 * time.Hour),
		Timezone:   "Asia/Ho_Chi_Minh",
		Status:     StatusDraft,
		WinRate:    40,
		DailySpins: 1,
		SpinRules:  []SpinRule{{Event: " Order_Completed ", Spins: 1, MaxPerDay: 2}},
		Prizes: []Prize{
			{ID: 1, Name: "100 points", Type: PrizePoints, Points: 100, Weight: 3},
			{ID: 2, Name: "500 points", Type: PrizePoints, Points: 500, Weight: 1, Quantity: qty(10)},
			{ID: 3, Name: "Better luck", Type: "NONE", Points: 9, Weight: 9},
		},
	}
}

func TestNormalizeAndValidate(t *testing.T) {
	c := validCampaign()
	c.Normalize()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.Name != "Wheel" || c.SpinRules[0].Event != "order_completed" {
		t.Fatalf("normalize: %+v", c)
	}
	blank := c.Blank()
	if blank == nil || blank.ID != 3 || blank.Points != 0 || blank.Weight != 0 {
		t.Fatalf("blank segment must be stripped of points: %+v", blank)
	}
	for i, p := range c.Prizes {
		if p.Position != i+1 {
			t.Fatalf("positions must be assigned in order: %+v", c.Prizes)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]func(c *Campaign){
		"no blank":      func(c *Campaign) { c.Prizes[2].Type = PrizePoints; c.Prizes[2].Points = 1; c.Prizes[2].Weight = 1 },
		"two blanks":    func(c *Campaign) { c.Prizes[0].Type = PrizeNone },
		"window":        func(c *Campaign) { c.EndAt = c.StartAt },
		"win rate":      func(c *Campaign) { c.WinRate = 100.5 },
		"timezone":      func(c *Campaign) { c.Timezone = "Mars/Base" },
		"zero weight":   func(c *Campaign) { c.Prizes[0].Weight = 0 },
		"zero points":   func(c *Campaign) { c.Prizes[0].Points = 0 },
		"dup rule":      func(c *Campaign) { c.SpinRules = append(c.SpinRules, SpinRule{Event: "order_completed", Spins: 1}) },
		"stock < given": func(c *Campaign) { c.Prizes[1].Awarded = 11 },
		"dup position":  func(c *Campaign) { c.Prizes[0].Position = 1; c.Prizes[1].Position = 1; c.Prizes[2].Position = 2 },
		"bad type":      func(c *Campaign) { c.Prizes[0].Type = "voucher" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := validCampaign()
			mutate(&c)
			c.Normalize()
			if err := c.Validate(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("expected invalid, got %v", err)
			}
		})
	}
}

func seq(values ...int) func(int) int {
	return func(n int) int {
		v := values[0]
		values = values[1:]
		if v >= n {
			panic("roll out of range")
		}
		return v
	}
}

func TestDraw(t *testing.T) {
	c := validCampaign()
	c.Normalize()
	if p := c.Draw(seq(4000)); p.Type != PrizeNone {
		t.Fatalf("roll at win rate boundary must lose, got %+v", p)
	}
	if p := c.Draw(seq(3999, 0)); p.ID != 1 {
		t.Fatalf("first weight bucket, got %+v", p)
	}
	if p := c.Draw(seq(0, 3)); p.ID != 2 {
		t.Fatalf("last weight bucket, got %+v", p)
	}
	c.Prizes[1].Awarded = 10
	if p := c.Draw(seq(0, 2)); p.ID != 1 {
		t.Fatalf("sold out prize must leave the pool, got %+v", p)
	}
	c.Prizes[0].Quantity = qty(0)
	if p := c.Draw(seq(0)); p.Type != PrizeNone {
		t.Fatalf("empty pool must fall back to blank, got %+v", p)
	}
	c = validCampaign()
	c.Normalize()
	c.WinRate = 0
	if p := c.Draw(seq(0)); p.Type != PrizeNone {
		t.Fatalf("0%% win rate never wins, got %+v", p)
	}
	c.WinRate = 100
	if p := c.Draw(seq(9999, 0)); p.Type != PrizePoints {
		t.Fatalf("100%% win rate always wins, got %+v", p)
	}
}

func TestDrawDistribution(t *testing.T) {
	c := validCampaign()
	c.Normalize()
	state := uint64(42)
	roll := func(n int) int {
		state = state*6364136223846793005 + 1442695040888963407
		return int((state >> 33) % uint64(n))
	}
	wins := 0
	const runs = 200_000
	for range runs {
		if c.Draw(roll).Type == PrizePoints {
			wins++
		}
	}
	rate := float64(wins) / runs * 100
	if rate < 39 || rate > 41 {
		t.Fatalf("win rate drifted: %.2f%%", rate)
	}
}

func TestAllowance(t *testing.T) {
	c := validCampaign()
	cases := []struct {
		name      string
		max       int
		u         Usage
		source    PlaySource
		ok        bool
		remaining int
	}{
		{"first spin of the day", 0, Usage{}, SourceDaily, true, 1},
		{"daily used, no credits", 0, Usage{DailyToday: 1}, "", false, 0},
		{"daily used, credits left", 0, Usage{DailyToday: 1, CreditBalance: 2}, SourceCredit, true, 2},
		{"daily cap reached", 2, Usage{DailyToday: 1, CreditToday: 1, CreditBalance: 5}, "", false, 0},
		{"daily cap limits remaining", 3, Usage{DailyToday: 1, CreditBalance: 5}, SourceCredit, true, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c.MaxSpinsPerDay = tc.max
			src, ok := c.Allowance(tc.u)
			if src != tc.source || ok != tc.ok {
				t.Fatalf("allowance = %q %v", src, ok)
			}
			if r := c.Remaining(tc.u); r != tc.remaining {
				t.Fatalf("remaining = %d", r)
			}
		})
	}
}

func TestDayUsesCampaignTimezone(t *testing.T) {
	c := validCampaign()
	at := time.Date(2026, 10, 20, 18, 30, 0, 0, time.UTC)
	if d := c.Day(at); d != "2026-10-21" {
		t.Fatalf("18:30 UTC is already the next day in Vietnam, got %s", d)
	}
}

func TestGames(t *testing.T) {
	g, err := ParseGame(" Gift_Box ")
	if err != nil || g != GameGiftBox || !g.Instant() || !g.NeedsChoice() {
		t.Fatalf("parse: %q %v", g, err)
	}
	if _, err := ParseGame("poker"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown game: %v", err)
	}
	if GameCheckin.Instant() {
		t.Fatal("checkin is not an instant-win game")
	}
	c := validCampaign()
	if c.ValidateChoice(0) != nil || c.ValidateChoice(2) == nil {
		t.Fatal("wheel takes no choice")
	}
	c.GameType = GameFlipCard
	if c.ValidateChoice(0) == nil || c.ValidateChoice(25) == nil || c.ValidateChoice(3) != nil {
		t.Fatal("flip card needs a choice between 1 and 24")
	}
}

func checkinCampaign() Campaign {
	c := validCampaign()
	c.GameType = GameCheckin
	c.Prizes, c.SpinRules = nil, nil
	c.CheckinRewards = []int64{10, 20, 50}
	return c
}

func TestCheckinValidation(t *testing.T) {
	c := checkinCampaign()
	c.Normalize()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(c *Campaign){
		"no rewards":       func(c *Campaign) { c.CheckinRewards = nil },
		"zero reward":      func(c *Campaign) { c.CheckinRewards = []int64{10, 0} },
		"prizes":           func(c *Campaign) { c.Prizes = validCampaign().Prizes },
		"rules":            func(c *Campaign) { c.SpinRules = []SpinRule{{Event: "x", Spins: 1}} },
		"too long a cycle": func(c *Campaign) { c.CheckinRewards = make([]int64, 32) },
		"settings array":   func(c *Campaign) { c.Settings = []byte(`[1]`) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := checkinCampaign()
			mutate(&c)
			c.Normalize()
			if err := c.Validate(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("expected invalid, got %v", err)
			}
		})
	}
	w := validCampaign()
	w.CheckinRewards = []int64{1}
	w.Normalize()
	if err := w.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("instant games have no streak rewards: %v", err)
	}
}

func TestStreak(t *testing.T) {
	c := checkinCampaign()
	if NextStreak(nil, "2026-10-20") != 1 || CurrentStreak(nil, "2026-10-20") != 0 {
		t.Fatal("first check-in starts the streak")
	}
	last := &Checkin{Day: "2026-10-19", Streak: 3}
	if NextStreak(last, "2026-10-20") != 4 || CurrentStreak(last, "2026-10-20") != 3 {
		t.Fatal("consecutive day continues the streak")
	}
	if NextStreak(last, "2026-10-21") != 1 || CurrentStreak(last, "2026-10-21") != 0 {
		t.Fatal("a missed day resets the streak")
	}
	if NextStreak(&Checkin{Day: "2026-02-28", Streak: 1}, "2026-03-01") != 2 {
		t.Fatal("month boundary")
	}
	for streak, want := range map[int]int64{1: 10, 3: 50, 4: 10, 6: 50} {
		if got := c.CheckinReward(streak); got != want {
			t.Fatalf("reward for day %d = %d, want %d", streak, got, want)
		}
	}
}

func TestMaskUser(t *testing.T) {
	for in, want := range map[string]string{"12345": "12***5", "7": "7***", "42": "4***", "": "***"} {
		if got := MaskUser(in); got != want {
			t.Fatalf("MaskUser(%q) = %q, want %q", in, got, want)
		}
	}
}
