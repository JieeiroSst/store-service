package model

import "time"

type Follow struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey"`
	FollowerID string    `json:"follower_id" gorm:"type:uuid;not null;index"`
	FollowedID string    `json:"followed_id" gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Follow) TableName() string { return "follows" }

// CursorKey implements CursorKeyed - followers/following lists are most-
// recent-first, like posts.
func (f Follow) CursorKey() (time.Time, string) { return f.CreatedAt, f.ID }
