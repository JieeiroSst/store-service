package model

import "time"

type Tag struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(50);unique;not null"`
	CreatedAt time.Time `json:"created_at"`
}

func (Tag) TableName() string { return "tags" }

type PostTag struct {
	PostID string `gorm:"type:uuid;primaryKey"`
	TagID  string `gorm:"type:uuid;primaryKey"`
}

func (PostTag) TableName() string { return "post_tags" }

type TagCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
