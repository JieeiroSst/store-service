package model

import (
	"time"

	"github.com/lib/pq"
)

type Comment struct {
	ID              string         `json:"id" gorm:"type:uuid;primaryKey"`
	PostID          string         `json:"post_id" gorm:"type:uuid;not null;index"`
	UserID          string         `json:"user_id" gorm:"type:uuid;not null;index"`
	ParentCommentID *string        `json:"parent_comment_id,omitempty" gorm:"type:uuid;index"`
	Content         string         `json:"content" gorm:"type:text;not null"`
	MediaURLs       pq.StringArray `json:"media_urls,omitempty" gorm:"type:text[]"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	Author *Author `json:"author,omitempty" gorm:"-"`
}

func (Comment) TableName() string { return "comments" }

// CursorKey implements CursorKeyed. Unlike posts/follows, comments paginate
// oldest-first (created_at ASC) - a discussion thread reads top-to-bottom
// in the order it happened, not most-recent-first.
func (c Comment) CursorKey() (time.Time, string) { return c.CreatedAt, c.ID }

type CreateCommentInput struct {
	PostID          string
	UserID          string
	ParentCommentID *string
	Content         string
	MediaURLs       []string
}
