package service

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type fakeReviews struct {
	outbound.ReviewRepository
	err error
}

func (f fakeReviews) Create(_ context.Context, r domain.Review) (domain.Review, error) {
	if f.err != nil {
		return domain.Review{}, f.err
	}
	r.ID = 42
	return r, nil
}

type fakeLoyalty struct {
	earned map[string]int64
	tiers  map[int64]int
	calls  int
}

func (f *fakeLoyalty) Earn(_ context.Context, member int64, key string, amount int64) error {
	if f.earned == nil {
		f.earned = map[string]int64{}
	}
	f.earned[key] = amount
	_ = member
	return nil
}
func (f *fakeLoyalty) Status(_ context.Context, id int64) (outbound.Member, error) {
	f.calls++
	return outbound.Member{Tier: f.tiers[id]}, nil
}

func TestReviewCreate(t *testing.T) {
	loy := &fakeLoyalty{}
	homes := fakeHomestayRepo{h: domain.Homestay{ID: 1, HostID: 77}}
	svc := NewReviewService(fakeReviews{}, homes, loy, nil)

	rv, err := svc.Create(context.Background(), guestP, inbound.ReviewCommand{HomestayID: 1, BookingID: 5, Rating: 5})
	if err != nil || rv.ID != 42 {
		t.Fatalf("create: %+v %v", rv, err)
	}
	if loy.earned["rent-house-review-42"] != 200 {
		t.Fatalf("a 5-star review must credit the host: %v", loy.earned)
	}

	loy.earned = nil
	svc.Create(context.Background(), guestP, inbound.ReviewCommand{HomestayID: 1, BookingID: 6, Rating: 3})
	if len(loy.earned) != 0 {
		t.Fatal("a 3-star review earns nothing")
	}
}

func TestReviewValidationAndEligibility(t *testing.T) {
	homes := fakeHomestayRepo{h: domain.Homestay{ID: 1, HostID: 77}}
	svc := NewReviewService(fakeReviews{}, homes, nil, nil)
	for name, c := range map[string]inbound.ReviewCommand{
		"rating 0":     {HomestayID: 1, BookingID: 1, Rating: 0},
		"rating 6":     {HomestayID: 1, BookingID: 1, Rating: 6},
		"no stay":      {HomestayID: 1, Rating: 5},
		"both stays":   {HomestayID: 1, BookingID: 1, LeaseID: 1, Rating: 5},
		"long comment": {HomestayID: 1, BookingID: 1, Rating: 5, Comment: string(make([]byte, 2001))},
	} {
		if _, err := svc.Create(context.Background(), guestP, c); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	// The repository decides who really stayed.
	svc = NewReviewService(fakeReviews{err: domain.ErrForbidden}, homes, nil, nil)
	if _, err := svc.Create(context.Background(), guestP, inbound.ReviewCommand{HomestayID: 1, BookingID: 1, Rating: 5}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestRecommendedRanksByRatingAndHostTier(t *testing.T) {
	list := []domain.Homestay{
		{ID: 9, HostID: 10, Rating: 4.5, ReviewCount: 20},  // bronze host, newer id
		{ID: 2, HostID: 20, Rating: 4.5, ReviewCount: 20},  // platinum host: same rating, boosted despite the older id
		{ID: 3, HostID: 10, Rating: 5, ReviewCount: 1},     // one glowing review must not lead
		{ID: 4, HostID: 30, Rating: 3.0, ReviewCount: 100}, // platinum but clearly worse
	}
	loy := &fakeLoyalty{tiers: map[int64]int{10: 1, 20: 4, 30: 4}}
	tiers := NewHostTiers(loy, 0)
	ranked := rank(list, tiers.Tiers(context.Background(), []int64{10, 20, 30, 10}))
	want := []int64{2, 9, 3, 4}
	for i, r := range ranked {
		if r.h.ID != want[i] {
			t.Fatalf("order = %v, want %v", scoredIDs(ranked), want)
		}
	}
}

func TestHostTiersCachesAndSurvivesFailure(t *testing.T) {
	loy := &fakeLoyalty{tiers: map[int64]int{10: 3}}
	tiers := NewHostTiers(loy, 1<<40)
	tiers.Tiers(context.Background(), []int64{10})
	tiers.Tiers(context.Background(), []int64{10})
	if loy.calls != 1 {
		t.Fatalf("expected the second lookup to hit the cache, calls = %d", loy.calls)
	}
	// Disabled loyalty: everyone is bronze, no panic.
	var none *HostTiers
	if got := none.Tiers(context.Background(), []int64{1, 2}); got[1] != 1 || got[2] != 1 {
		t.Fatalf("nil tiers: %v", got)
	}
}

func scoredIDs(rs []scored) []int64 {
	out := make([]int64, len(rs))
	for i, r := range rs {
		out[i] = r.h.ID
	}
	return out
}

func ids(hs []domain.Homestay) []int64 {
	out := make([]int64, len(hs))
	for i, h := range hs {
		out[i] = h.ID
	}
	return out
}
