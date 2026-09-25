package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

// searchRepo returns a fixed candidate list and remembers the filter it was asked with.
type searchRepo struct {
	outbound.HomestayRepository
	all  []domain.Homestay
	last domain.HomestayFilter
}

func (r *searchRepo) List(_ context.Context, f domain.HomestayFilter) ([]domain.Homestay, error) {
	r.last = f
	if f.Limit > 0 && f.Limit < len(r.all) {
		return append([]domain.Homestay(nil), r.all[:f.Limit]...), nil
	}
	return append([]domain.Homestay(nil), r.all...), nil
}

func catalogue() []domain.Homestay {
	// IDs 1..7, ratings chosen so the recommended order (with tiers below) is 6,4,2,7,5,3,1.
	return []domain.Homestay{
		{ID: 1, HostID: 10, Rating: 3.0, ReviewCount: 10},
		{ID: 2, HostID: 10, Rating: 4.5, ReviewCount: 30},
		{ID: 3, HostID: 10, Rating: 3.5, ReviewCount: 10},
		{ID: 4, HostID: 10, Rating: 4.8, ReviewCount: 40},
		{ID: 5, HostID: 10, Rating: 4.0, ReviewCount: 20},
		{ID: 6, HostID: 20, Rating: 4.9, ReviewCount: 50}, // platinum host on top
		{ID: 7, HostID: 10, Rating: 4.2, ReviewCount: 25},
	}
}

func TestRecommendedPagesWalkTheWholeRankingOnce(t *testing.T) {
	repo := &searchRepo{all: catalogue()}
	svc := NewHomestayService(repo, nil, NewHostTiers(&fakeLoyalty{tiers: map[int64]int{10: 1, 20: 4}}, 1<<40))
	ctx := context.Background()

	whole, err := svc.List(ctx, domain.HomestayFilter{Limit: 100})
	if err != nil || whole.Next != nil {
		t.Fatalf("whole: %v next=%v", err, whole.Next)
	}

	var got []int64
	var cur *domain.HomestayCursor
	pages := 0
	for {
		pg, err := svc.List(ctx, domain.HomestayFilter{Limit: 3, After: cur})
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, ids(pg.Items)...)
		pages++
		if pg.Next == nil {
			break
		}
		if len(pg.Items) != 3 {
			t.Fatalf("only the last page may be short: %d", len(pg.Items))
		}
		cur = pg.Next
		if pages > 10 {
			t.Fatal("cursor does not advance")
		}
	}
	if !reflect.DeepEqual(got, ids(whole.Items)) || len(got) != 7 || pages != 3 {
		t.Fatalf("paged %v in %d pages, whole %v", got, pages, ids(whole.Items))
	}
	if got[0] != 6 {
		t.Fatalf("the platinum host's top-rated homestay leads: %v", got)
	}
}

func TestRecommendedCursorSurvivesNewHomestays(t *testing.T) {
	repo := &searchRepo{all: catalogue()}
	svc := NewHomestayService(repo, nil, nil)
	first, _ := svc.List(context.Background(), domain.HomestayFilter{Limit: 2})
	// A new, top-rated homestay appears between two page requests: an offset would repeat one
	// item on page 2; a cursor continues right after where page 1 ended.
	repo.all = append(repo.all, domain.Homestay{ID: 99, HostID: 10, Rating: 5, ReviewCount: 200})
	second, err := svc.List(context.Background(), domain.HomestayFilter{Limit: 2, After: first.Next})
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range second.Items {
		for _, seen := range first.Items {
			if h.ID == seen.ID || h.ID == 99 {
				t.Fatalf("page 2 %v repeats page 1 %v or jumps back to the new item", ids(second.Items), ids(first.Items))
			}
		}
	}
}

func TestDatabaseSortsPassCursorAndFetchOneExtra(t *testing.T) {
	repo := &searchRepo{all: []domain.Homestay{
		{ID: 9, Rating: 4.5, ReviewCount: 3, Rates: []domain.Rate{{Model: domain.ModelMonth, Price: "5000000.00", Active: true}}},
		{ID: 8, Rating: 4.5, ReviewCount: 2, Rates: []domain.Rate{{Model: domain.ModelMonth, Price: "6000000.00", Active: true}}},
		{ID: 7, Rating: 4.0, ReviewCount: 9},
	}}
	svc := NewHomestayService(repo, nil, nil)
	ctx := context.Background()

	pg, err := svc.List(ctx, domain.HomestayFilter{Sort: "rating", Limit: 2})
	if err != nil || repo.last.Limit != 3 || len(pg.Items) != 2 {
		t.Fatalf("one extra row tells whether there is more: %v limit=%d items=%d", err, repo.last.Limit, len(pg.Items))
	}
	want := &domain.HomestayCursor{Sort: "rating", ID: 8, Rating: 4.5, Count: 2}
	if !reflect.DeepEqual(pg.Next, want) {
		t.Fatalf("next = %+v, want %+v", pg.Next, want)
	}
	if _, err := svc.List(ctx, domain.HomestayFilter{Sort: "rating", Limit: 2, After: pg.Next}); err != nil || !reflect.DeepEqual(repo.last.After, want) {
		t.Fatalf("cursor must reach the repository: %v %+v", err, repo.last.After)
	}

	// Price sorts remember the price of the model being sorted on ("" when the homestay has none).
	pg, _ = svc.List(ctx, domain.HomestayFilter{Sort: "price_asc", Model: domain.ModelMonth, Limit: 2})
	if pg.Next == nil || pg.Next.Price != "6000000.00" || pg.Next.ID != 8 {
		t.Fatalf("price cursor: %+v", pg.Next)
	}
	// Last page: nothing more.
	repo.all = repo.all[:2]
	if pg, _ = svc.List(ctx, domain.HomestayFilter{Sort: "newest", Limit: 2}); pg.Next != nil {
		t.Fatalf("no next page expected: %+v", pg.Next)
	}
}

func TestCursorMustMatchTheSort(t *testing.T) {
	svc := NewHomestayService(&searchRepo{all: catalogue()}, nil, nil)
	_, err := svc.List(context.Background(), domain.HomestayFilter{Sort: "newest", After: &domain.HomestayCursor{Sort: "rating", ID: 1}})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
	// No sort means recommended.
	_, err = svc.List(context.Background(), domain.HomestayFilter{After: &domain.HomestayCursor{Sort: "newest", ID: 1}})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("default sort vs newest cursor: %v", err)
	}
}
