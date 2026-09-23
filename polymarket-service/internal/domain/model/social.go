package model

import "time"

type Profile struct {
	UserID    string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	Username  string    `gorm:"column:username" json:"username"`
	Bio       string    `gorm:"column:bio" json:"bio,omitempty"`
	AvatarURL string    `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Profile) TableName() string { return "profiles" }

type Comment struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EventID   int64     `gorm:"column:event_id" json:"event_id"`
	ParentID  int64     `gorm:"column:parent_id" json:"parent_id,omitempty"`
	UserID    string    `gorm:"column:user_id" json:"user_id"`
	Body      string    `gorm:"column:body" json:"body"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	Likes int64 `gorm:"column:likes;->" json:"likes"`
}

func (Comment) TableName() string { return "comments" }

type CommentLike struct {
	CommentID int64     `gorm:"column:comment_id;primaryKey"`
	UserID    string    `gorm:"column:user_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (CommentLike) TableName() string { return "comment_likes" }

type Bookmark struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	EventID   int64     `gorm:"column:event_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Bookmark) TableName() string { return "bookmarks" }
