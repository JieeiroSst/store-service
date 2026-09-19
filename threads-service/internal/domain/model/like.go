package model

import "time"

type Like struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	PostID    *string   `json:"post_id,omitempty" gorm:"type:uuid;index"`
	CommentID *string   `json:"comment_id,omitempty" gorm:"type:uuid;index"`
	CreatedAt time.Time `json:"created_at"`
}

func (Like) TableName() string { return "likes" }
