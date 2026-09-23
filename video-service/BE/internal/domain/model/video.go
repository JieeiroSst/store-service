package model

import (
	"regexp"
	"time"
)

var idPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }

type Status string

const (
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusFailed     Status = "failed"
)

type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Status      Status    `json:"status"`
	Duration    float64   `json:"duration"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	CreatedAt   time.Time `json:"created_at"`
	Views       int64     `json:"views"`
}

func (v Video) ObjectKey() string { return "videos/" + v.ID }

func (v Video) ThumbnailKey() string { return "thumbs/" + v.ID + ".jpg" }

func (v Video) HLSPrefix() string { return "hls/" + v.ID + "/" }

func (v Video) ETag() string { return `"` + v.ID + `"` }
