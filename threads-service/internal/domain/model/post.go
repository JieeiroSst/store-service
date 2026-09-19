package model

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey"`
	UserID       string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Content      string         `json:"content" gorm:"type:text;not null"`
	MediaURLs    pq.StringArray `json:"media_urls,omitempty" gorm:"type:text[]"`
	LikeCount    int            `json:"like_count" gorm:"not null;default:0"`
	CommentCount int            `json:"comment_count" gorm:"not null;default:0"`
	// RepostOfID is set when this row is a repost (Threads-style "reshare")
	// of another post, rather than an original post. RepostCount lives on
	// the *original* post, incremented whenever someone reposts it.
	RepostOfID  *string   `json:"repost_of_id,omitempty" gorm:"type:uuid;index"`
	RepostCount int       `json:"repost_count" gorm:"not null;default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Author, Tags, and OriginalPost are populated by the application
	// layer from port.UserClient / port.TagRepository / a second
	// PostRepository.GetByID lookup - never persisted directly on this
	// table, hence gorm:"-".
	Author       *Author  `json:"author,omitempty" gorm:"-"`
	Tags         []string `json:"tags,omitempty" gorm:"-"`
	OriginalPost *Post    `json:"original_post,omitempty" gorm:"-"`
}

func (Post) TableName() string { return "posts" }

// CursorKey implements CursorKeyed - every post list (global feed, home
// feed, bookmarks) is ordered and paginated by (created_at, id) DESC.
func (p Post) CursorKey() (time.Time, string) { return p.CreatedAt, p.ID }

type CreatePostInput struct {
	UserID    string
	Content   string
	MediaURLs []string
	Tags      []string
}
