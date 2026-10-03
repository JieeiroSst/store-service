package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxChatChars = 2000

type MediaState struct {
	Audio  bool `json:"audio"`
	Video  bool `json:"video"`
	Screen bool `json:"screen"`
}

func DefaultMediaState() MediaState {
	return MediaState{Audio: true, Video: true}
}

func ParseMediaState(data json.RawMessage) (MediaState, error) {
	var m MediaState
	if err := json.Unmarshal(data, &m); err != nil {
		return MediaState{}, fmt.Errorf("%w: media_state data must be {audio, video, screen}", ErrInvalidMessage)
	}
	return m, nil
}

type ChatPayload struct {
	Text   string    `json:"text"`
	SentAt time.Time `json:"sent_at"`
}

func ParseChat(data json.RawMessage, now time.Time) (ChatPayload, error) {
	var in struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return ChatPayload{}, fmt.Errorf("%w: chat data must be {text}", ErrInvalidMessage)
	}
	text := strings.TrimSpace(in.Text)
	if text == "" || !utf8.ValidString(text) || utf8.RuneCountInString(text) > MaxChatChars {
		return ChatPayload{}, fmt.Errorf("%w: chat text must be 1-%d characters", ErrInvalidMessage, MaxChatChars)
	}
	return ChatPayload{Text: text, SentAt: now.UTC()}, nil
}

type Participant struct {
	UserID string     `json:"user_id"`
	Media  MediaState `json:"media"`
}

type RoomInfo struct {
	RoomID       string        `json:"room_id"`
	Capacity     int           `json:"capacity"`
	Participants []Participant `json:"participants"`
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}
