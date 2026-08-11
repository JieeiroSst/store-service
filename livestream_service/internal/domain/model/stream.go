package model

import "time"

type StreamStatus string

const (
	StreamStatusPending StreamStatus = "pending"
	StreamStatusLive    StreamStatus = "live"
	StreamStatusEnded   StreamStatus = "ended"
)

type Stream struct {
	ID         string       `gorm:"primaryKey" json:"id"`
	RoomID     string       `gorm:"column:room_id;index" json:"roomId"`
	NodeID     string       `gorm:"column:node_id" json:"nodeId"`
	Status     StreamStatus `json:"status"`
	StartedAt  *time.Time   `gorm:"column:started_at" json:"startedAt,omitempty"`
	EndedAt    *time.Time   `gorm:"column:ended_at" json:"endedAt,omitempty"`
	PeakViewer int          `gorm:"column:peak_viewer" json:"peakViewer"`
	CreatedAt  time.Time    `json:"createdAt"`
}

func (Stream) TableName() string { return "streams" }

type IngestEndpoint struct {
	RTMPURL   string
	NodeID    string
	StreamKey string
}
