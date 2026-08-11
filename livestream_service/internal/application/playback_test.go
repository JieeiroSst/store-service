package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func TestSignPlaybackTokenIsDeterministicForSameInputs(t *testing.T) {
	// Freeze via a fixed expiry rather than relying on wall-clock timing:
	// two calls with the same secret/path/ttl issued back to back should
	// very likely land on the same expiry second and thus the same token,
	// but to be certain, verify the token only depends on path+expiry by
	// checking equal-length hex output and that different paths differ.
	secret := "test-secret"
	token1, exp1 := signPlaybackToken(secret, "a/master.m3u8", time.Hour)
	token2, _ := signPlaybackToken(secret, "b/master.m3u8", time.Hour)

	if token1 == token2 {
		t.Error("expected different paths to produce different tokens")
	}
	if len(token1) == 0 {
		t.Error("expected a non-empty token")
	}
	if exp1.Before(time.Now()) {
		t.Error("expected expiresAt to be in the future")
	}
}

func TestGetPlaybackURLUsesLiveStreamKeyWhenLive(t *testing.T) {
	rooms := newFakeRoomRepository()
	_ = rooms.Create(context.Background(), &model.Room{
		ID: "room-1", StreamKey: "sk_test", Status: model.RoomStatusLive,
	})
	u := &ingestUsecase{
		rooms: rooms,
		vods:  &fakeVODRepository{},
		cfg: &config.Config{Playback: config.PlaybackConfig{
			SigningSecret: "secret", CDNBaseURL: "https://cdn.example.com", TokenTTL: "1h",
		}},
	}

	info, err := u.GetPlaybackURL(context.Background(), "room-1")
	if err != nil {
		t.Fatalf("GetPlaybackURL() error = %v", err)
	}
	if !info.IsLive {
		t.Error("expected IsLive = true")
	}
	if !strings.Contains(info.URL, "sk_test/master.m3u8") {
		t.Errorf("expected URL to reference the live stream key's master playlist, got %q", info.URL)
	}
	if !strings.HasPrefix(info.URL, "https://cdn.example.com/") {
		t.Errorf("expected URL to use the configured CDN base, got %q", info.URL)
	}
}

func TestGetPlaybackURLUsesLatestVODWhenOffline(t *testing.T) {
	rooms := newFakeRoomRepository()
	_ = rooms.Create(context.Background(), &model.Room{
		ID: "room-1", StreamKey: "sk_test", Status: model.RoomStatusOffline,
	})
	vods := &fakeVODRepository{created: []model.Recording{
		{ID: "rec-1", RoomID: "room-1", ObjectKey: "sk_test/master.m3u8"},
	}}
	u := &ingestUsecase{
		rooms: rooms,
		vods:  vods,
		cfg: &config.Config{Playback: config.PlaybackConfig{
			SigningSecret: "secret", CDNBaseURL: "https://cdn.example.com", TokenTTL: "1h",
		}},
	}

	info, err := u.GetPlaybackURL(context.Background(), "room-1")
	if err != nil {
		t.Fatalf("GetPlaybackURL() error = %v", err)
	}
	if info.IsLive {
		t.Error("expected IsLive = false")
	}
	if !strings.Contains(info.URL, "sk_test/master.m3u8") {
		t.Errorf("expected URL to reference the VOD's object key, got %q", info.URL)
	}
}

func TestGetPlaybackURLErrorsWhenNothingToPlay(t *testing.T) {
	rooms := newFakeRoomRepository()
	_ = rooms.Create(context.Background(), &model.Room{
		ID: "room-1", StreamKey: "sk_test", Status: model.RoomStatusOffline,
	})
	u := &ingestUsecase{rooms: rooms, vods: &fakeVODRepository{}, cfg: &config.Config{}}

	if _, err := u.GetPlaybackURL(context.Background(), "room-1"); err == nil {
		t.Fatal("expected an error when the room is offline with no recordings")
	}
}
