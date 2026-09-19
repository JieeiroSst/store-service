package model

// PostCategory is the posts<->categories join row (table new_categories -
// see Post's doc comment for why the table name still says "new").
type PostCategory struct {
	NewID      string `gorm:"column:new_id;primaryKey"`
	CategoryID string `gorm:"column:category_id;primaryKey"`
}

func (PostCategory) TableName() string { return "new_categories" }
