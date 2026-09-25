package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
)

type memRepo struct {
	mu     sync.Mutex
	videos map[string]model.Video
	source map[string]string
	ix     []model.Interaction
}

func newMemRepo() *memRepo {
	return &memRepo{videos: map[string]model.Video{}, source: map[string]string{}}
}

func (r *memRepo) Upsert(_ context.Context, source string, vs []model.Video) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, v := range vs {
		r.videos[v.ID], r.source[v.ID] = v, source
	}
	return nil
}
func (r *memRepo) List(context.Context) ([]model.Video, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []model.Video
	for _, v := range r.videos {
		out = append(out, v)
	}
	return out, nil
}
func (r *memRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.videos, id)
	return nil
}
func (r *memRepo) DeleteMissing(_ context.Context, source string, keep []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := map[string]bool{}
	for _, id := range keep {
		k[id] = true
	}
	for id := range r.videos {
		if r.source[id] == source && !k[id] {
			delete(r.videos, id)
		}
	}
	return nil
}
func (r *memRepo) Add(_ context.Context, in model.Interaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ix = append(r.ix, in)
	return nil
}
func (r *memRepo) Since(_ context.Context, since time.Time) ([]model.Interaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []model.Interaction
	for _, in := range r.ix {
		if !in.At.Before(since) {
			out = append(out, in)
		}
	}
	return out, nil
}
func (r *memRepo) ByUser(_ context.Context, user string, _ int) ([]model.Interaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []model.Interaction
	for _, in := range r.ix {
		if in.UserID == user {
			out = append(out, in)
		}
	}
	return out, nil
}
func (r *memRepo) DeleteByVideo(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	kept := r.ix[:0]
	for _, in := range r.ix {
		if in.VideoID != id {
			kept = append(kept, in)
		}
	}
	r.ix = kept
	return nil
}
func (r *memRepo) DeleteByUserVideo(_ context.Context, user, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	kept := r.ix[:0]
	for _, in := range r.ix {
		if !(in.UserID == user && in.VideoID == id) {
			kept = append(kept, in)
		}
	}
	r.ix = kept
	return nil
}
func (r *memRepo) Ping(context.Context) error { return nil }

type snap struct {
	scope   string
	ranked  []engine.Scored
	expires time.Time
}

// snapshots is shared between services in tests, standing in for Postgres.
type snapshots struct {
	mu sync.Mutex
	m  map[string]snap
}

func newSnapshots() *snapshots { return &snapshots{m: map[string]snap{}} }

func (r *snapshots) Save(_ context.Context, id, scope string, ranked []engine.Scored, exp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[id] = snap{scope, ranked, exp}
	return nil
}
func (r *snapshots) Load(_ context.Context, id, scope string, now time.Time) ([]engine.Scored, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sn, ok := r.m[id]
	if !ok || sn.scope != scope || !sn.expires.After(now) {
		return nil, false, nil
	}
	return sn.ranked, true, nil
}
func (r *snapshots) DeleteExpired(_ context.Context, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, sn := range r.m {
		if !sn.expires.After(now) {
			delete(r.m, id)
		}
	}
	return nil
}

type fakeSource struct {
	videos []model.Video
	err    error
}

func (f fakeSource) Fetch(context.Context) ([]model.Video, error) { return f.videos, f.err }

var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func ready(id, title string) model.Video {
	return model.Video{ID: id, Title: title, Status: model.StatusReady, CreatedAt: t0}
}

func newService(repo *memRepo, src port.CatalogSource) *Service {
	return newServiceWith(repo, newSnapshots(), src)
}

func newServiceWith(repo *memRepo, snaps *snapshots, src port.CatalogSource) *Service {
	cfg := &config.Config{}
	cfg.Train.Window = 90 * 24 * time.Hour
	cfg.Snapshot.TTL = 15 * time.Minute
	cfg.Engine = engine.DefaultConfig()
	s := NewService(repo, repo, snaps, src, repo, cfg)
	s.now = func() time.Time { return t0 }
	return s
}

func TestSyncMirrorsCatalogAndPrunes(t *testing.T) {
	repo := newMemRepo()
	_ = repo.Upsert(context.Background(), sourceManual, []model.Video{ready("manual", "Manual video")})
	src := fakeSource{videos: []model.Video{ready("a", "Space battle"), ready("b", "Galaxy war")}}
	s := newService(repo, src)
	ctx := context.Background()

	if err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	// Video "a" disappears from video-service: it must leave the catalog, the
	// manually registered one must stay.
	s.source = fakeSource{videos: []model.Video{ready("b", "Galaxy war")}}
	if err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	page, _ := s.Trending(ctx, port.PageQuery{})
	got := map[string]bool{}
	for _, r := range page.Items {
		got[r.Video.ID] = true
	}
	if got["a"] || !got["b"] || !got["manual"] {
		t.Fatalf("catalog after prune = %v", got)
	}
}

func TestSyncFailureKeepsCatalog(t *testing.T) {
	repo := newMemRepo()
	s := newService(repo, fakeSource{videos: []model.Video{ready("a", "Space battle")}})
	if err := s.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.source = fakeSource{err: errors.New("video-service down")}
	if err := s.Sync(context.Background()); err == nil {
		t.Fatal("expected the source error to be reported")
	}
	if page, _ := s.Trending(context.Background(), port.PageQuery{}); page.Total != 1 {
		t.Fatalf("catalog was lost on failed sync: %d items", page.Total)
	}
}

func TestPersonalizedFlow(t *testing.T) {
	repo := newMemRepo()
	s := newService(repo, fakeSource{videos: []model.Video{
		ready("a", "Space battle"), ready("b", "Galaxy war"), ready("c", "Cooking pasta"),
	}})
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(s.Sync(ctx))
	must(s.RecordEvent(ctx, model.Interaction{UserID: "other", VideoID: "a", Type: model.Like}))
	must(s.RecordEvent(ctx, model.Interaction{UserID: "other", VideoID: "b", Type: model.Like}))
	must(s.RecordEvent(ctx, model.Interaction{UserID: "me", VideoID: "a", Type: model.Like}))
	must(s.Train(ctx))

	page, err := s.ForUser(ctx, "me", port.PageQuery{})
	must(err)
	recs := page.Items
	if len(recs) == 0 || recs[0].Video.ID != "b" || recs[0].Because == nil || recs[0].Because.ID != "a" {
		t.Fatalf("recs = %+v", recs)
	}
	for _, r := range recs {
		if r.Video.ID == "a" {
			t.Fatal("already-liked video recommended")
		}
	}
}

func TestValidationAndNotFound(t *testing.T) {
	s := newService(newMemRepo(), fakeSource{})
	ctx := context.Background()
	if err := s.RecordEvent(ctx, model.Interaction{UserID: "u", VideoID: "v", Type: "nope"}); !errors.Is(err, port.ErrInvalid) {
		t.Fatalf("RecordEvent err = %v", err)
	}
	if _, err := s.ForUser(ctx, "", port.PageQuery{}); !errors.Is(err, port.ErrInvalid) {
		t.Fatalf("ForUser err = %v", err)
	}
	if _, err := s.Similar(ctx, "missing", port.PageQuery{}); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("Similar err = %v", err)
	}
	if err := s.UpsertVideo(ctx, model.Video{ID: "x"}); !errors.Is(err, port.ErrInvalid) {
		t.Fatalf("UpsertVideo err = %v", err)
	}
}

func TestDeleteVideoDropsItImmediately(t *testing.T) {
	repo := newMemRepo()
	s := newService(repo, fakeSource{})
	ctx := context.Background()
	if err := s.UpsertVideo(ctx, model.Video{ID: "x", Title: "Manual"}); err != nil {
		t.Fatal(err)
	}
	if page, _ := s.Trending(ctx, port.PageQuery{}); page.Total != 1 {
		t.Fatalf("upsert not visible: %d", page.Total)
	}
	if err := s.DeleteVideo(ctx, "x"); err != nil {
		t.Fatal(err)
	}
	if page, _ := s.Trending(ctx, port.PageQuery{}); page.Total != 0 {
		t.Fatalf("deleted video still served: %d", page.Total)
	}
}

func TestPagination(t *testing.T) {
	repo := newMemRepo()
	var vs []model.Video
	for i := 0; i < 45; i++ {
		vs = append(vs, ready(fmt.Sprintf("v%02d", i), fmt.Sprintf("Video %d", i)))
	}
	s := newService(repo, fakeSource{videos: vs})
	ctx := context.Background()
	if err := s.Sync(ctx); err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	for p := 1; p <= 3; p++ {
		page, err := s.ForUser(ctx, "newcomer", port.PageQuery{Page: p, PageSize: 20})
		if err != nil {
			t.Fatal(err)
		}
		wantLen := map[int]int{1: 20, 2: 20, 3: 5}[p]
		if page.Total != 45 || page.Page != p || page.PageSize != 20 || len(page.Items) != wantLen {
			t.Fatalf("page %d: total=%d page=%d size=%d items=%d", p, page.Total, page.Page, page.PageSize, len(page.Items))
		}
		for _, r := range page.Items {
			if seen[r.Video.ID] {
				t.Fatalf("video %s repeated across pages", r.Video.ID)
			}
			seen[r.Video.ID] = true
		}
	}
	if len(seen) != 45 {
		t.Fatalf("paged through %d videos, want 45", len(seen))
	}

	past, _ := s.Trending(ctx, port.PageQuery{Page: 99})
	if len(past.Items) != 0 || past.Total != 45 {
		t.Fatalf("past the end: %d items, total %d", len(past.Items), past.Total)
	}
	huge, _ := s.Trending(ctx, port.PageQuery{PageSize: 1000})
	if huge.PageSize != maxPageSize || len(huge.Items) != maxPageSize-5 {
		t.Fatalf("page_size not capped: %d / %d items", huge.PageSize, len(huge.Items))
	}
}

func idsOf(p *port.Page) []string {
	var out []string
	for _, r := range p.Items {
		out = append(out, r.Video.ID)
	}
	return out
}

func snapshotFixture() []model.Video {
	return []model.Video{
		ready("a", "Space battle"), ready("b", "Galaxy war"), ready("c", "Cooking pasta"),
		ready("d", "Baking bread"), ready("e", "Alien planet"), ready("f", "Ocean life"),
	}
}

func seed(ctx context.Context, s *Service) {
	_ = s.RecordEvent(ctx, model.Interaction{UserID: "me", VideoID: "a", Type: model.Like})
	_ = s.RecordEvent(ctx, model.Interaction{UserID: "o", VideoID: "a", Type: model.Like})
	_ = s.RecordEvent(ctx, model.Interaction{UserID: "o", VideoID: "b", Type: model.Like})
}

// Two services sharing storage stand in for two replicas: page 2 must come
// from the ranking page 1 was cut from even though the other replica has since
// retrained on new events and the user has liked more videos.
func TestSnapshotPinsRankingAcrossReplicasAndRetrains(t *testing.T) {
	repo, snaps := newMemRepo(), newSnapshots()
	src := fakeSource{videos: snapshotFixture()}
	a, b := newServiceWith(repo, snaps, src), newServiceWith(repo, snaps, src)
	ctx := context.Background()
	if err := a.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	seed(ctx, a)
	_ = a.Train(ctx)
	_ = b.Train(ctx)

	first, err := a.ForUser(ctx, "me", port.PageQuery{Page: 1, PageSize: 2})
	if err != nil || first.Snapshot == "" {
		t.Fatalf("first page: %v %+v", err, first)
	}
	want, _ := a.ForUser(ctx, "me", port.PageQuery{Page: 2, PageSize: 2, Snapshot: first.Snapshot})

	for _, v := range []string{"c", "d", "e"} {
		_ = b.RecordEvent(ctx, model.Interaction{UserID: "me", VideoID: v, Type: model.Like})
	}
	_ = b.RecordEvent(ctx, model.Interaction{UserID: "o", VideoID: "f", Type: model.Like})
	_ = b.Train(ctx)

	got, err := b.ForUser(ctx, "me", port.PageQuery{Page: 2, PageSize: 2, Snapshot: first.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if got.Snapshot != first.Snapshot || fmt.Sprint(idsOf(got)) != fmt.Sprint(idsOf(want)) {
		t.Fatalf("pinned page 2 changed: %v -> %v", idsOf(want), idsOf(got))
	}

	fresh, _ := b.ForUser(ctx, "me", port.PageQuery{Page: 1, PageSize: 2})
	if fresh.Snapshot == first.Snapshot {
		t.Fatal("fresh read reused the old snapshot")
	}
}

func TestSnapshotExpiredOrForeignFallsBackWithNewToken(t *testing.T) {
	repo := newMemRepo()
	s := newService(repo, fakeSource{videos: snapshotFixture()})
	clock := t0
	s.now = func() time.Time { return clock }
	ctx := context.Background()
	_ = s.Sync(ctx)

	first, _ := s.Trending(ctx, port.PageQuery{PageSize: 2})

	// Same token, different scope: must not leak another ranking.
	other, err := s.ForUser(ctx, "me", port.PageQuery{PageSize: 2, Snapshot: first.Snapshot})
	if err != nil || other.Snapshot == first.Snapshot {
		t.Fatalf("foreign-scope token accepted: %v %s", err, other.Snapshot)
	}

	clock = clock.Add(16 * time.Minute)
	next, err := s.Trending(ctx, port.PageQuery{PageSize: 2, Snapshot: first.Snapshot})
	if err != nil || next.Snapshot == first.Snapshot || next.Snapshot == "" {
		t.Fatalf("expired token: %v %q (was %q)", err, next.Snapshot, first.Snapshot)
	}
	if err := s.Cleanup(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestSinglePageResultHasNoSnapshot(t *testing.T) {
	s := newService(newMemRepo(), fakeSource{videos: snapshotFixture()})
	_ = s.Sync(context.Background())
	page, _ := s.Trending(context.Background(), port.PageQuery{PageSize: 50})
	if page.Snapshot != "" || page.Total != 6 {
		t.Fatalf("snapshot %q total %d", page.Snapshot, page.Total)
	}
}

func TestBadSnapshotIsInvalid(t *testing.T) {
	s := newService(newMemRepo(), fakeSource{})
	for _, tok := range []string{"!!!", "abc", "ZZZZ"} {
		if _, err := s.Trending(context.Background(), port.PageQuery{Snapshot: tok}); !errors.Is(err, port.ErrInvalid) {
			t.Fatalf("token %q: err = %v", tok, err)
		}
	}
}
