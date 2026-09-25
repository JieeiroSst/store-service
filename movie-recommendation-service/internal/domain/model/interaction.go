package model

import (
	"fmt"
	"time"
)

type InteractionType string

const (
	Watch   InteractionType = "watch"
	Like    InteractionType = "like"
	Dislike InteractionType = "dislike"
	Rating  InteractionType = "rating"
	View    InteractionType = "view"
)

const maxIDLen = 128

type Interaction struct {
	UserID  string          `json:"user_id"`
	VideoID string          `json:"video_id"`
	Type    InteractionType `json:"type"`
	Value   float64         `json:"value"`
	At      time.Time       `json:"at"`
}

func (i Interaction) Validate() error {
	if i.UserID == "" || len(i.UserID) > maxIDLen {
		return fmt.Errorf("user_id must be 1-%d characters", maxIDLen)
	}
	if i.VideoID == "" || len(i.VideoID) > maxIDLen {
		return fmt.Errorf("video_id must be 1-%d characters", maxIDLen)
	}
	switch i.Type {
	case View, Like, Dislike:
	case Watch:
		if i.Value < 0 || i.Value > 1 {
			return fmt.Errorf("watch value must be within 0..1")
		}
	case Rating:
		if i.Value < 1 || i.Value > 5 {
			return fmt.Errorf("rating value must be within 1..5")
		}
	default:
		return fmt.Errorf("unknown type %q", i.Type)
	}
	return nil
}

func (i Interaction) Weight() float64 {
	switch i.Type {
	case View:
		return 1
	case Watch:
		return 3 * i.Value
	case Like:
		return 4
	case Dislike:
		return -4
	case Rating:
		return (i.Value - 3) * 2
	}
	return 0
}
