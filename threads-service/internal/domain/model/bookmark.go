package model

import "time"

type Bookmark struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_bookmarks_user_post"`
	PostID    string    `json:"post_id" gorm:"type:uuid;not null;uniqueIndex:idx_bookmarks_user_post"`
	CreatedAt time.Time `json:"created_at"`
}

func (Bookmark) TableName() string { return "bookmarks" }
