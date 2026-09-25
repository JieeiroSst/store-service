package redis

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func newStore(t *testing.T) (*SnapshotStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return NewSnapshotStore(rdb), mr
}

func TestSaveLoadRoundTripAndScope(t *testing.T) {
	s, _ := newStore(t)
	ctx := t.Context()
	want := []engine.Scored{{VideoID: "a", Score: 0.5, Reason: "for_you"}, {VideoID: "b", Score: 0.2, Reason: "because_you_watched", BecauseID: "a"}}
	if err := s.Save(ctx, "id1", "user:u", want, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	got, ok, err := s.Load(ctx, "id1", "user:u", time.Now())
	if err != nil || !ok || len(got) != 2 || got[1] != want[1] {
		t.Fatalf("Load = %+v ok=%v err=%v", got, ok, err)
	}
	if _, ok, _ := s.Load(ctx, "id1", "user:someone-else", time.Now()); ok {
		t.Fatal("snapshot served for a different scope")
	}
	if _, ok, _ := s.Load(ctx, "missing", "user:u", time.Now()); ok {
		t.Fatal("unknown id reported found")
	}
}

func TestSnapshotExpires(t *testing.T) {
	s, mr := newStore(t)
	ctx := t.Context()
	if err := s.Save(ctx, "id1", "trending", []engine.Scored{{VideoID: "a"}}, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	mr.FastForward(2 * time.Minute)
	if _, ok, _ := s.Load(ctx, "id1", "trending", time.Now()); ok {
		t.Fatal("expired snapshot still served")
	}
}
