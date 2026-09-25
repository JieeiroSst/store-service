package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
)

var now = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func vid(id, title string, tags ...string) model.Video {
	return model.Video{ID: id, Title: title, Tags: tags, Status: model.StatusReady, CreatedAt: now}
}

func like(user, video string) model.Interaction {
	return model.Interaction{UserID: user, VideoID: video, Type: model.Like, At: now}
}

func catalog() []model.Video {
	return []model.Video{
		vid("a", "Space battle", "scifi", "action"),
		vid("b", "Galaxy war", "scifi", "action"),
		vid("c", "Cooking pasta", "food"),
		vid("d", "Baking bread", "food"),
		vid("e", "Alien planet", "scifi"),
	}
}

func ids(s []Scored) []string {
	var out []string
	for _, x := range s {
		out = append(out, x.VideoID)
	}
	return out
}

func TestSimilarUsesContent(t *testing.T) {
	m := Build(catalog(), nil, now, DefaultConfig())
	got, ok := m.Similar("a", 2)
	if !ok || len(got) == 0 || got[0].VideoID != "b" {
		t.Fatalf("Similar(a) = %v, want b first", ids(got))
	}
	if _, ok := m.Similar("zzz", 5); ok {
		t.Fatal("unknown video should report not found")
	}
}

func TestSimilarUsesCoViewing(t *testing.T) {
	// c and e share nothing textually, but the same users like both.
	var ix []model.Interaction
	for _, u := range []string{"u1", "u2", "u3"} {
		ix = append(ix, like(u, "c"), like(u, "e"))
	}
	m := Build(catalog(), ix, now, DefaultConfig())
	got, _ := m.Similar("c", 1)
	if len(got) != 1 || got[0].VideoID != "e" || got[0].Reason != model.ReasonWatchedTogether {
		t.Fatalf("Similar(c) = %+v, want e watched_together", got)
	}
}

func TestForUserExcludesSeenAndPersonalizes(t *testing.T) {
	ix := []model.Interaction{like("other", "a"), like("other", "b")}
	m := Build(catalog(), ix, now, DefaultConfig())
	got := m.ForUser([]model.Interaction{like("me", "a")}, 3)
	if len(got) == 0 || got[0].VideoID != "b" {
		t.Fatalf("ForUser = %v, want b first", ids(got))
	}
	if got[0].Reason != model.ReasonBecauseWatched || got[0].BecauseID != "a" {
		t.Fatalf("reason = %s because %s", got[0].Reason, got[0].BecauseID)
	}
	for _, s := range got {
		if s.VideoID == "a" {
			t.Fatal("seen video was recommended")
		}
	}
}

func TestForUserDislikeIsNotRecommendedAndColdStartIsTrending(t *testing.T) {
	ix := []model.Interaction{like("x", "e"), like("y", "e")}
	m := Build(catalog(), ix, now, DefaultConfig())

	cold := m.ForUser(nil, 1)
	if len(cold) != 1 || cold[0].VideoID != "e" || cold[0].Reason != model.ReasonTrending {
		t.Fatalf("cold start = %+v, want trending e", cold)
	}

	dislike := model.Interaction{UserID: "me", VideoID: "e", Type: model.Dislike, At: now}
	for _, s := range m.ForUser([]model.Interaction{dislike}, 10) {
		if s.VideoID == "e" {
			t.Fatal("disliked video was recommended")
		}
	}
}

func TestTrendingDecaysOldInteractions(t *testing.T) {
	old := now.Add(-60 * 24 * time.Hour)
	ix := []model.Interaction{
		{UserID: "u1", VideoID: "c", Type: model.Like, At: old},
		{UserID: "u2", VideoID: "c", Type: model.Like, At: old},
		{UserID: "u3", VideoID: "d", Type: model.Like, At: now},
	}
	m := Build(catalog(), ix, now, DefaultConfig())
	if got := m.Trending(1, nil); got[0].VideoID != "d" {
		t.Fatalf("trending = %v, want d (recent) over c (stale)", ids(got))
	}
}

func TestBuildSkipsNotReadyAndDuplicates(t *testing.T) {
	vs := append(catalog(), model.Video{ID: "f", Title: "processing", Status: "processing"}, vid("a", "dup"))
	m := Build(vs, nil, now, DefaultConfig())
	if m.Len() != 5 {
		t.Fatalf("Len = %d, want 5", m.Len())
	}
	if _, ok := m.Video("f"); ok {
		t.Fatal("non-ready video indexed")
	}
}

func TestInteractionWeightAndValidate(t *testing.T) {
	cases := []struct {
		in model.Interaction
		w  float64
		ok bool
	}{
		{model.Interaction{UserID: "u", VideoID: "v", Type: model.Rating, Value: 5}, 4, true},
		{model.Interaction{UserID: "u", VideoID: "v", Type: model.Rating, Value: 3}, 0, true},
		{model.Interaction{UserID: "u", VideoID: "v", Type: model.Rating, Value: 6}, 0, false},
		{model.Interaction{UserID: "u", VideoID: "v", Type: model.Watch, Value: 0.5}, 1.5, true},
		{model.Interaction{UserID: "u", VideoID: "v", Type: model.Watch, Value: 2}, 0, false},
		{model.Interaction{UserID: "", VideoID: "v", Type: model.Like}, 0, false},
		{model.Interaction{UserID: "u", VideoID: "v", Type: "bogus"}, 0, false},
	}
	for _, c := range cases {
		if err := c.in.Validate(); (err == nil) != c.ok {
			t.Errorf("Validate(%+v) err=%v, want ok=%v", c.in, err, c.ok)
		}
		if c.ok && c.in.Weight() != c.w {
			t.Errorf("Weight(%+v) = %v, want %v", c.in, c.in.Weight(), c.w)
		}
	}
}

func TestNewReleasesFavoursRecentAndSkipsExcluded(t *testing.T) {
	old := vid("old", "Old video")
	old.CreatedAt = now.Add(-120 * 24 * time.Hour)
	fresh := vid("fresh", "Fresh video")
	fresh.CreatedAt = now.Add(-time.Hour)
	undated := model.Video{ID: "undated", Title: "No date", Status: model.StatusReady}
	m := Build([]model.Video{old, fresh, undated}, nil, now, DefaultConfig())

	got := m.NewReleases(10, nil)
	if fmt.Sprint(ids(got)) != "[fresh old]" || got[0].Reason != model.ReasonNewRelease {
		t.Fatalf("NewReleases = %v", ids(got))
	}
	if got := m.NewReleases(10, map[string]bool{"fresh": true}); fmt.Sprint(ids(got)) != "[old]" {
		t.Fatalf("exclude ignored: %v", ids(got))
	}
}

func TestDiversityInterleavesNearDuplicates(t *testing.T) {
	// a1 and a2 are identical; b is a different topic that still matches the
	// query video x a little. Without diversity the twins take the top two
	// slots; with it b gets the second one.
	videos := []model.Video{
		vid("x", "space battle galaxy"),
		vid("a1", "space battle galaxy"),
		vid("a2", "space battle galaxy"),
		vid("b", "space cooking"),
	}
	plain := DefaultConfig()
	plain.Diversity = 0
	mixed := DefaultConfig()
	mixed.Diversity = 0.5

	p, _ := Build(videos, nil, now, plain).Similar("x", 3)
	if fmt.Sprint(ids(p)) != "[a1 a2 b]" {
		t.Fatalf("plain = %v", ids(p))
	}
	d, _ := Build(videos, nil, now, mixed).Similar("x", 3)
	if fmt.Sprint(ids(d)) != "[a1 b a2]" {
		t.Fatalf("diversified = %v", ids(d))
	}
}
