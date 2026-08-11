package model

import "time"

// PlaybackInfo is a time-limited, signed URL for a room's current HLS
// output - live master playlist if the room is live, otherwise its most
// recent VOD recording.
type PlaybackInfo struct {
	URL       string
	ExpiresAt time.Time
	IsLive    bool
}
