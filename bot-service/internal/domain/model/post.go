package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

type Media struct {
	URL  string    `json:"url"`
	Type MediaType `json:"type"`
	// Thumbnail is only meaningful for Type == MediaTypeVideo.
	Thumbnail string `json:"thumbnail,omitempty"`
}

type MediaKind string

const (
	MediaKindTextOnly    MediaKind = "text_only"
	MediaKindSingleImage MediaKind = "single_image"
	MediaKindMultiImage  MediaKind = "multi_image"
	MediaKindVideo       MediaKind = "video"
	MediaKindReel        MediaKind = "reel"
)

func DeriveMediaKind(media []Media) MediaKind {
	switch {
	case len(media) == 0:
		return MediaKindTextOnly
	case len(media) > 1:
		return MediaKindMultiImage
	case media[0].Type == MediaTypeVideo:
		return MediaKindVideo
	default:
		return MediaKindSingleImage
	}
}

type PostStatus string

const (
	PostStatusDraft         PostStatus = "draft"
	PostStatusPendingReview PostStatus = "pending_review"
	PostStatusScheduled     PostStatus = "scheduled"
	PostStatusPublishing    PostStatus = "publishing"
	PostStatusPublished     PostStatus = "published"
	PostStatusFailed        PostStatus = "failed"
	PostStatusActive        PostStatus = "active"
	PostStatusCancelled     PostStatus = "cancelled"
)

type ChannelPublishResult struct {
	Channel      Channel   `json:"channel"`
	ExternalID   string    `json:"external_id,omitempty"`
	PublishedURL string    `json:"published_url,omitempty"`
	Error        string    `json:"error,omitempty"`
	PublishedAt  time.Time `json:"published_at"`
}

type Post struct {
	ID              string      `gorm:"primaryKey" json:"id"`
	Title           string      `json:"title"`
	Text            string      `json:"text"`
	Hashtags        StringList  `gorm:"type:jsonb" json:"hashtags"`
	Media           MediaList   `gorm:"type:jsonb" json:"media"`
	MediaKind       MediaKind   `json:"media_kind,omitempty"`
	Channels        ChannelList `gorm:"type:jsonb" json:"channels"`
	ScheduledAt     time.Time   `json:"scheduled_at"`
	Timezone        string      `json:"timezone,omitempty"`
	CronExpr        string      `json:"cron_expr,omitempty"`
	MaxRunsPerDay   int         `json:"max_runs_per_day,omitempty"`
	MaxRunsPerMonth int         `json:"max_runs_per_month,omitempty"`
	RunsToday       int         `json:"runs_today,omitempty"`
	RunsThisMonth   int         `json:"runs_this_month,omitempty"`
	LastRunDate     string      `json:"last_run_date,omitempty"`
	LastRunMonth    string      `json:"last_run_month,omitempty"`
	LastRunAt       *time.Time  `json:"last_run_at,omitempty"`
	Status          PostStatus  `json:"status"`
	RejectReason    string      `json:"reject_reason,omitempty"`
	Results         ResultList  `gorm:"type:jsonb" json:"results"`
	Campaign        string      `json:"campaign,omitempty"`
	CreatedBy       string      `json:"created_by,omitempty"`
	ApprovedBy      string      `json:"approved_by,omitempty"`
	StatusChangedBy string      `json:"status_changed_by,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func (Post) TableName() string { return "posts" }

func (p Post) IsRecurring() bool { return p.CronExpr != "" }

func (p Post) HasSchedule() bool { return p.CronExpr != "" || !p.ScheduledAt.IsZero() }

type MediaList []Media

func (m MediaList) Value() (driver.Value, error) { return marshalJSONColumn(m) }
func (m *MediaList) Scan(v any) error            { return scanJSONColumn(v, m) }

type ChannelList []Channel

func (c ChannelList) Value() (driver.Value, error) { return marshalJSONColumn(c) }
func (c *ChannelList) Scan(v any) error            { return scanJSONColumn(v, c) }

type ResultList []ChannelPublishResult

func (r ResultList) Value() (driver.Value, error) { return marshalJSONColumn(r) }
func (r *ResultList) Scan(v any) error            { return scanJSONColumn(v, r) }

type StringList []string

func (s StringList) Value() (driver.Value, error) { return marshalJSONColumn(s) }
func (s *StringList) Scan(v any) error            { return scanJSONColumn(v, s) }

func marshalJSONColumn(v any) (driver.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func scanJSONColumn(v any, out any) error {
	if v == nil {
		return nil
	}
	switch b := v.(type) {
	case []byte:
		return json.Unmarshal(b, out)
	case string:
		return json.Unmarshal([]byte(b), out)
	default:
		return errors.New("model: unsupported scan source for JSON column")
	}
}
