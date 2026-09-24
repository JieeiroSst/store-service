package model

import "time"

type Category struct {
	ID                 int64
	ParentID           *int64
	Name               string
	Slug               string
	Description        string
	Image              string
	IsPublic           bool
	AncestorsArePublic bool
	MetaTitle          string
	MetaDescription    string
}

type ProductClass struct {
	ID               int64
	Name             string
	Slug             string
	RequiresShipping bool
	TrackStock       bool
	OptionIDs        []int64
}

type OptionType string

const (
	OptionText    OptionType = "text"
	OptionInteger OptionType = "integer"
	OptionFloat   OptionType = "float"
	OptionBoolean OptionType = "boolean"
	OptionDate    OptionType = "date"
)

func (t OptionType) Valid() bool {
	switch t {
	case OptionText, OptionInteger, OptionFloat, OptionBoolean, OptionDate:
		return true
	}
	return false
}

type Option struct {
	ID       int64
	Name     string
	Code     string
	Type     OptionType
	Required bool
}

type Structure string

const (
	StructureStandalone Structure = "standalone"
	StructureParent     Structure = "parent"
	StructureChild      Structure = "child"
)

func (s Structure) Valid() bool {
	switch s {
	case StructureStandalone, StructureParent, StructureChild:
		return true
	}
	return false
}

type Product struct {
	ID              int64
	Structure       Structure
	UPC             *string
	Title           string
	Slug            string
	Description     string
	Rating          *float64
	IsDiscountable  bool
	IsPublic        bool
	ParentID        *int64
	ProductClassID  *int64
	MetaTitle       string
	MetaDescription string
	DateCreated     time.Time
	DateUpdated     time.Time

	CategoryIDs     []int64
	OptionIDs       []int64
	Images          []ProductImage
	Recommendations []Recommendation
}

type ProductImage struct {
	ID           int64
	ProductID    int64
	Original     string
	Caption      string
	DisplayOrder int
	DateCreated  time.Time
}

type Recommendation struct {
	RecommendedID int64
	Ranking       int
}

type ProductFilter struct {
	CategoryID *int64
	Query      string
	OnlyPublic bool
	Limit      int
	Offset     int
}
