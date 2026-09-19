package model

import "time"

// Post is what the DB and earlier code called "New" (Vietnamese "bài viết
// mới" abbreviated) - renamed because "New"/"new" as a Go identifier
// shadows the built-in and reads as an adjective, not a noun, everywhere
// it's used. TableName keeps the underlying MySQL table as "news" so this
// rename needs no data migration.
type Post struct {
	ID          string `json:"id" gorm:"primaryKey"`
	AuthorID    string `json:"author_id" gorm:"not null;index"`
	Name        string `json:"name" gorm:"not null"`
	Content     string `json:"content" gorm:"type:text"`
	Description string `json:"description" gorm:"type:text"`
	// MediaID is a pointer, not a plain string: a text-only post has no
	// media at all, and media_id must be SQL NULL for that (not "") or
	// the fk_news_media foreign key constraint rejects the insert - an
	// empty string is never a valid media.id to reference.
	MediaID   *string   `json:"media_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Categories still joins through new_categories.new_id/category_id -
	// those column names predate this Go-level New->Post rename and are
	// kept as-is to avoid a data migration; joinForeignKey/joinReferences
	// below point at PostCategory's Go field names, which still match.
	Categories []Category `json:"categories,omitempty" gorm:"many2many:new_categories;joinForeignKey:NewID;joinReferences:CategoryID"`
	// Media is the post's single cover image/attachment (see MediaID) -
	// the original code declared this as a slice with no working
	// association (Media had no foreign key back to its post at all), so
	// Preload("Medias") never actually returned anything.
	Media *Media `json:"media,omitempty" gorm:"foreignKey:MediaID"`
}

func (Post) TableName() string { return "news" }

// CursorKey implements CursorKeyed - the post list is ordered and
// paginated by (created_at, id) DESC (see model.Cursor's doc comment).
func (p Post) CursorKey() (time.Time, string) { return p.CreatedAt, p.ID }

type CreatePostInput struct {
	AuthorID    string
	Name        string
	Content     string
	Description string
	CategoryID  string
}

type UpdatePostInput struct {
	Name        string
	Content     string
	Description string
}
