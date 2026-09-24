package repository

import (
	"time"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
)

type categoryEntity struct {
	ID                 int64  `gorm:"primaryKey"`
	ParentID           *int64 `gorm:"index"`
	Name               string `gorm:"not null"`
	Slug               string `gorm:"not null;uniqueIndex"`
	Description        string
	Image              string
	IsPublic           bool `gorm:"not null;default:true"`
	AncestorsArePublic bool `gorm:"not null;default:true"`
	MetaTitle          string
	MetaDescription    string

	Parent *categoryEntity `gorm:"foreignKey:ParentID;constraint:OnDelete:RESTRICT"`
}

func (categoryEntity) TableName() string { return "catalogue_category" }

func (e categoryEntity) toModel() model.Category {
	return model.Category{
		ID: e.ID, ParentID: e.ParentID, Name: e.Name, Slug: e.Slug,
		Description: e.Description, Image: e.Image, IsPublic: e.IsPublic,
		AncestorsArePublic: e.AncestorsArePublic,
		MetaTitle:          e.MetaTitle, MetaDescription: e.MetaDescription,
	}
}

func categoryToEntity(c *model.Category) categoryEntity {
	return categoryEntity{
		ID: c.ID, ParentID: c.ParentID, Name: c.Name, Slug: c.Slug,
		Description: c.Description, Image: c.Image, IsPublic: c.IsPublic,
		AncestorsArePublic: c.AncestorsArePublic,
		MetaTitle:          c.MetaTitle, MetaDescription: c.MetaDescription,
	}
}

type optionEntity struct {
	ID       int64  `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	Code     string `gorm:"not null;uniqueIndex"`
	Type     string `gorm:"not null"`
	Required bool   `gorm:"not null"`
}

func (optionEntity) TableName() string { return "catalogue_option" }

func (e optionEntity) toModel() model.Option {
	return model.Option{ID: e.ID, Name: e.Name, Code: e.Code, Type: model.OptionType(e.Type), Required: e.Required}
}

func optionToEntity(o *model.Option) optionEntity {
	return optionEntity{ID: o.ID, Name: o.Name, Code: o.Code, Type: string(o.Type), Required: o.Required}
}

type productClassEntity struct {
	ID               int64  `gorm:"primaryKey"`
	Name             string `gorm:"not null"`
	Slug             string `gorm:"not null;uniqueIndex"`
	RequiresShipping bool   `gorm:"not null"`
	TrackStock       bool   `gorm:"not null"`
}

func (productClassEntity) TableName() string { return "catalogue_productclass" }

type productClassOptionEntity struct {
	ProductClassID int64 `gorm:"primaryKey"`
	OptionID       int64 `gorm:"primaryKey"`

	ProductClass productClassEntity `gorm:"foreignKey:ProductClassID;constraint:OnDelete:CASCADE"`
	Option       optionEntity       `gorm:"foreignKey:OptionID;constraint:OnDelete:CASCADE"`
}

func (productClassOptionEntity) TableName() string { return "catalogue_productclass_options" }

type productEntity struct {
	ID              int64   `gorm:"primaryKey"`
	Structure       string  `gorm:"not null"`
	UPC             *string `gorm:"uniqueIndex"`
	Title           string  `gorm:"not null"`
	Slug            string  `gorm:"not null;index"`
	Description     string
	Rating          *float64
	IsDiscountable  bool `gorm:"not null"`
	IsPublic        bool `gorm:"not null;default:true"`
	ParentID        *int64
	ProductClassID  *int64
	MetaTitle       string
	MetaDescription string
	DateCreated     time.Time `gorm:"autoCreateTime"`
	DateUpdated     time.Time `gorm:"autoUpdateTime"`

	Parent       *productEntity      `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	ProductClass *productClassEntity `gorm:"foreignKey:ProductClassID;constraint:OnDelete:RESTRICT"`
}

func (productEntity) TableName() string { return "catalogue_product" }

type productCategoryEntity struct {
	ProductID  int64 `gorm:"primaryKey"`
	CategoryID int64 `gorm:"primaryKey;index"`

	Product  productEntity  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Category categoryEntity `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}

func (productCategoryEntity) TableName() string { return "catalogue_productcategory" }

type productOptionEntity struct {
	ProductID int64 `gorm:"primaryKey"`
	OptionID  int64 `gorm:"primaryKey"`

	Product productEntity `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Option  optionEntity  `gorm:"foreignKey:OptionID;constraint:OnDelete:CASCADE"`
}

func (productOptionEntity) TableName() string { return "catalogue_product_product_options" }

type productImageEntity struct {
	ID           int64 `gorm:"primaryKey"`
	ProductID    int64 `gorm:"not null;index"`
	Original     string
	Caption      string
	DisplayOrder int
	DateCreated  time.Time `gorm:"autoCreateTime"`

	Product productEntity `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}

func (productImageEntity) TableName() string { return "catalogue_productimage" }

type recommendationEntity struct {
	ID            int64 `gorm:"primaryKey"`
	PrimaryID     int64 `gorm:"not null;uniqueIndex:uq_recommendation"`
	RecommendedID int64 `gorm:"not null;uniqueIndex:uq_recommendation"`
	Ranking       int   `gorm:"not null"`

	Primary     productEntity `gorm:"foreignKey:PrimaryID;constraint:OnDelete:CASCADE"`
	Recommended productEntity `gorm:"foreignKey:RecommendedID;constraint:OnDelete:CASCADE"`
}

func (recommendationEntity) TableName() string { return "catalogue_productrecommendation" }

func Entities() []any {
	return []any{
		&categoryEntity{}, &optionEntity{}, &productClassEntity{}, &productClassOptionEntity{},
		&productEntity{}, &productCategoryEntity{}, &productOptionEntity{},
		&productImageEntity{}, &recommendationEntity{},
	}
}
