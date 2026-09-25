package model

import "time"

const (
	ReasonTrending        = "trending"
	ReasonBecauseWatched  = "because_you_watched"
	ReasonSimilarContent  = "similar_content"
	ReasonWatchedTogether = "watched_together"
	ReasonNewRelease      = "new_release"
	ReasonContinue        = "continue_watching"
	ReasonHistory         = "history"
	ReasonForYou          = "for_you"
)

type Recommendation struct {
	Video    Video      `json:"video"`
	Score    float64    `json:"score"`
	Reason   string     `json:"reason"`
	Because  *Video     `json:"because,omitempty"`
	Progress *float64   `json:"progress,omitempty"`
	At       *time.Time `json:"at,omitempty"`
}

type Section struct {
	ID    string           `json:"id"`
	Title string           `json:"title"`
	Items []Recommendation `json:"items"`
}

type Home struct {
	Sections []Section `json:"sections"`
}
