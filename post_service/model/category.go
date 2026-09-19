package model

import "time"

type Category struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Posts []Post `json:"posts,omitempty" gorm:"many2many:new_categories;joinForeignKey:CategoryID;joinReferences:NewID"`
}

func (Category) TableName() string { return "categories" }

// CursorKey implements CursorKeyed - see model.Cursor's doc comment.
func (c Category) CursorKey() (time.Time, string) { return c.CreatedAt, c.ID }

type CreateCategoryInput struct {
	Name        string
	Description string
}

type UpdateCategoryInput struct {
	Name        string
	Description string
}
