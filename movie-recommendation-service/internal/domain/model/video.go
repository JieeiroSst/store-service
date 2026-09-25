package model

import "time"

const StatusReady = "ready"

type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"`
	Duration    float64   `json:"duration"`
	Views       int64     `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
}

func (v Video) Ready() bool { return v.Status == StatusReady }
