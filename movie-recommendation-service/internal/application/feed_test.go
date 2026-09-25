package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
)

func ev(user, video string, typ model.InteractionType, value float64, at time.Time) model.Interaction {
	return model.Interaction{UserID: user, VideoID: video, Type: typ, Value: value, At: at}
}

func feedService(t *testing.T) (*Service, *memRepo) {
	t.Helper()
	repo := newMemRepo()
	vs := []model.Video{
		ready("a", "Space battle"), ready("b", "Galaxy war"), ready("c", "Cooking pasta"),
		ready("d", "Baking bread"), ready("e", "Alien planet"),
	}
	for i := range vs {
		vs[i].CreatedAt = t0.Add(-time.Duration(i) * 24 * time.Hour) // a is newest
	}
	s := newService(repo, fakeSource{videos: vs})
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s, repo
}

func TestContinueWatchingShowsOnlyInProgressNewestFirst(t *testing.T) {
	s, _ := feedService(t)
	ctx := context.Background()
	for _, e := range []model.Interaction{
		ev("u", "a", model.Watch, 0.4, t0.Add(1*time.Minute)),
		ev("u", "b", model.Watch, 0.95, t0.Add(2*time.Minute)), // finished
		ev("u", "c", model.Watch, 0.01, t0.Add(3*time.Minute)), // accidental click
		ev("u", "d", model.Watch, 0.2, t0.Add(4*time.Minute)),
		ev("u", "e", model.Like, 0, t0.Add(5*time.Minute)), // never watched
		ev("u", "a", model.Watch, 0.6, t0.Add(6*time.Minute)),
	} {
		if err := s.RecordEvent(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.ContinueWatching(ctx, "u", port.PageQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(idsOf(page)) != "[a d]" || page.Total != 2 {
		t.Fatalf("continue watching = %v", idsOf(page))
	}
	if p := *page.Items[0].Progress; p != 0.6 {
		t.Fatalf("progress = %v, want the latest position 0.6", p)
	}
	if page.Items[0].Reason != model.ReasonContinue || page.Items[0].At == nil {
		t.Fatalf("item = %+v", page.Items[0])
	}
}

func TestFinishingAVideoRemovesItFromContinueWatching(t *testing.T) {
	s, _ := feedService(t)
	ctx := context.Background()
	_ = s.RecordEvent(ctx, ev("u", "a", model.Watch, 0.5, t0.Add(time.Minute)))
	_ = s.RecordEvent(ctx, ev("u", "a", model.Watch, 1, t0.Add(2*time.Minute)))
	page, _ := s.ContinueWatching(ctx, "u", port.PageQuery{})
	if page.Total != 0 {
		t.Fatalf("finished video still listed: %v", idsOf(page))
	}
}

func TestHistoryIsDistinctPagedAndRemovable(t *testing.T) {
	s, repo := feedService(t)
	ctx := context.Background()
	_ = s.RecordEvent(ctx, ev("u", "a", model.View, 0, t0.Add(1*time.Minute)))
	_ = s.RecordEvent(ctx, ev("u", "b", model.Like, 0, t0.Add(2*time.Minute)))
	_ = s.RecordEvent(ctx, ev("u", "a", model.Like, 0, t0.Add(3*time.Minute)))
	_ = s.RecordEvent(ctx, ev("other", "c", model.Like, 0, t0.Add(4*time.Minute)))

	page, err := s.History(ctx, "u", port.PageQuery{PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(idsOf(page)) != "[a]" || page.Total != 2 {
		t.Fatalf("history page 1 = %v total %d", idsOf(page), page.Total)
	}
	page2, _ := s.History(ctx, "u", port.PageQuery{Page: 2, PageSize: 1})
	if fmt.Sprint(idsOf(page2)) != "[b]" {
		t.Fatalf("history page 2 = %v", idsOf(page2))
	}

	if err := s.RemoveFromHistory(ctx, "u", "a"); err != nil {
		t.Fatal(err)
	}
	after, _ := s.History(ctx, "u", port.PageQuery{})
	if fmt.Sprint(idsOf(after)) != "[b]" {
		t.Fatalf("after remove = %v", idsOf(after))
	}
	if len(repo.ix) != 2 { // b's like and other's like remain
		t.Fatalf("removal touched other rows: %d left", len(repo.ix))
	}
	if err := s.RemoveFromHistory(ctx, "", "a"); err == nil {
		t.Fatal("expected invalid request")
	}
}

func TestNewReleasesNewestFirst(t *testing.T) {
	s, _ := feedService(t)
	page, err := s.NewReleases(context.Background(), port.PageQuery{PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(idsOf(page)) != "[a b]" || page.Total != 5 || page.Snapshot == "" {
		t.Fatalf("new = %v total %d snapshot %q", idsOf(page), page.Total, page.Snapshot)
	}
}

func rowIDs(h *model.Home) []string {
	var out []string
	for _, sec := range h.Sections {
		out = append(out, sec.ID)
	}
	return out
}

func TestHomeForNewUserIsNewReleasesAndTrending(t *testing.T) {
	s, _ := feedService(t)
	home, err := s.Home(context.Background(), "nobody")
	if err != nil {
		t.Fatal(err)
	}
	// Trending is fully covered by new releases here, and rows never repeat a video.
	if fmt.Sprint(rowIDs(home)) != "[new_releases]" || len(home.Sections[0].Items) != 5 {
		t.Fatalf("rows = %v", rowIDs(home))
	}
}

// Enough content-similar videos that "recommended for you" (10) cannot use
// them all, leaving some for the "because you watched" row.
func TestHomeRowsForActiveUserHaveNoDuplicates(t *testing.T) {
	repo := newMemRepo()
	vs := []model.Video{ready("a", "Space battle"), ready("b", "Space war"), ready("c", "Cooking pasta")}
	for i := 0; i < 30; i++ {
		vs = append(vs, ready(fmt.Sprintf("f%02d", i), fmt.Sprintf("Space filler %d", i)))
	}
	for i := range vs {
		vs[i].CreatedAt = t0.Add(-time.Duration(i) * time.Hour)
	}
	s := newService(repo, fakeSource{videos: vs})
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(s.Sync(ctx))
	must(s.RecordEvent(ctx, ev("o", "a", model.Like, 0, t0.Add(time.Minute))))
	must(s.RecordEvent(ctx, ev("o", "b", model.Like, 0, t0.Add(time.Minute))))
	must(s.RecordEvent(ctx, ev("me", "a", model.Like, 0, t0.Add(2*time.Minute))))
	must(s.RecordEvent(ctx, ev("me", "c", model.Watch, 0.5, t0.Add(3*time.Minute))))
	must(s.Train(ctx))

	home, err := s.Home(ctx, "me")
	if err != nil {
		t.Fatal(err)
	}
	rows := rowIDs(home)
	if fmt.Sprint(rows) != "[continue_watching for_you because_a new_releases trending]" {
		t.Fatalf("rows = %v", rows)
	}
	if home.Sections[2].Title != "Because you watched Space battle" {
		t.Fatalf("title = %q", home.Sections[2].Title)
	}
	seen := map[string]string{}
	for _, sec := range home.Sections {
		if len(sec.Items) > homeRowSize {
			t.Fatalf("%s has %d items", sec.ID, len(sec.Items))
		}
		for _, it := range sec.Items {
			if prev, dup := seen[it.Video.ID]; dup {
				t.Fatalf("%s appears in %s and %s", it.Video.ID, prev, sec.ID)
			}
			seen[it.Video.ID] = sec.ID
			if sec.ID != "continue_watching" && (it.Video.ID == "a" || it.Video.ID == "c") {
				t.Fatalf("already-seen %s recommended in %s", it.Video.ID, sec.ID)
			}
		}
	}
}
