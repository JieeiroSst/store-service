package model

import "time"

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "ACTIVE"
	ProductStatusDraft    ProductStatus = "DRAFT"
	ProductStatusArchived ProductStatus = "ARCHIVED"
)

func (s ProductStatus) Valid() bool {
	switch s {
	case ProductStatusActive, ProductStatusDraft, ProductStatusArchived:
		return true
	}
	return false
}

type Product struct {
	ID               int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	ShopifyID        string           `json:"shopify_id" gorm:"column:shopify_id;uniqueIndex"`
	Title            string           `json:"title"`
	Handle           string           `json:"handle"`
	Status           ProductStatus    `json:"status"`
	Vendor           string           `json:"vendor"`
	ProductType      string           `json:"product_type"`
	Tags             string           `json:"tags"`
	Variants         []ProductVariant `json:"variants" gorm:"foreignKey:ProductID"`
	ShopifyCreatedAt time.Time        `json:"shopify_created_at"`
	ShopifyUpdatedAt time.Time        `json:"shopify_updated_at"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

func (Product) TableName() string { return "products" }

type ProductVariant struct {
	ID                int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID         int64     `json:"product_id"`
	ShopifyID         string    `json:"shopify_id" gorm:"column:shopify_id;uniqueIndex"`
	Title             string    `json:"title"`
	SKU               string    `json:"sku" gorm:"column:sku"`
	Price             string    `json:"price"`
	InventoryQuantity int       `json:"inventory_quantity"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ProductVariant) TableName() string { return "product_variants" }
