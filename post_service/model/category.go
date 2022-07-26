package model

type Category struct {
	Id          string
	Name        string
	Description string
	News        []New `gorm:"many2many:new_categories;"`
}
