package model

type New struct {
	Id          string
	AuthorId    string
	Name        string
	Content     string
	Description string
	MediaId     string
	Categories  []Category `gorm:"many2many:new_categories;"`
	Medias      []Media
}
