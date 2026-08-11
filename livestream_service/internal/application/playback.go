package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

// signPlaybackToken produces an HMAC-SHA256 token over path+expiry - the
// same shape CDN "token authentication" features (Cloudflare, BunnyCDN,
// etc.) expect, though the exact algorithm/param names vary by CDN and
// whichever one fronts this deployment needs to be configured to validate
// tokens in this shape (or this function swapped for that CDN's scheme).
// Pure computation, no I/O, so it's a plain function rather than a port.
func signPlaybackToken(secret, path string, ttl time.Duration) (token string, expiresAt time.Time) {
	expiresAt = time.Now().Add(ttl)
	expiresUnix := strconv.FormatInt(expiresAt.Unix(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(path))
	mac.Write([]byte(expiresUnix))
	token = hex.EncodeToString(mac.Sum(nil))
	return token, expiresAt
}

func (u *ingestUsecase) GetPlaybackURL(ctx context.Context, roomID string) (*model.PlaybackInfo, error) {
	room, err := u.rooms.GetByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}

	var path string
	isLive := room.Status == model.RoomStatusLive
	if isLive {
		path = room.StreamKey + "/master.m3u8"
	} else {
		recordings, err := u.vods.ListByRoom(ctx, roomID)
		if err != nil {
			return nil, fmt.Errorf("list recordings: %w", err)
		}
		if len(recordings) == 0 {
			return nil, fmt.Errorf("room %q has no live stream or recordings to play back", roomID)
		}
		path = recordings[0].ObjectKey // ListByRoom orders newest first
	}

	ttl := u.cfg.Playback.TokenTTLDuration()
	token, expiresAt := signPlaybackToken(u.cfg.Playback.SigningSecret, path, ttl)

	return &model.PlaybackInfo{
		URL:       fmt.Sprintf("%s/%s?token=%s&expires=%d", u.cfg.Playback.CDNBaseURL, path, token, expiresAt.Unix()),
		ExpiresAt: expiresAt,
		IsLive:    isLive,
	}, nil
}
