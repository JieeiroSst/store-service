package model

import "time"

type Order struct {
	ID                int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	ShopifyID         string     `json:"shopify_id" gorm:"column:shopify_id;uniqueIndex"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	FinancialStatus   string     `json:"financial_status"`
	FulfillmentStatus string     `json:"fulfillment_status"`
	Currency          string     `json:"currency"`
	TotalPrice        string     `json:"total_price"`
	CancelledAt       *time.Time `json:"cancelled_at"`
	ProcessedAt       time.Time  `json:"processed_at"`
	LineItems         []LineItem `json:"line_items" gorm:"foreignKey:OrderID"`
	ShopifyCreatedAt  time.Time  `json:"shopify_created_at"`
	ShopifyUpdatedAt  time.Time  `json:"shopify_updated_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

type LineItem struct {
	ID               int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderID          int64     `json:"order_id"`
	ShopifyID        string    `json:"shopify_id" gorm:"column:shopify_id;uniqueIndex"`
	VariantShopifyID string    `json:"variant_shopify_id" gorm:"column:variant_shopify_id"`
	Title            string    `json:"title"`
	SKU              string    `json:"sku" gorm:"column:sku"`
	Quantity         int       `json:"quantity"`
	UnitPrice        string    `json:"unit_price"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (LineItem) TableName() string { return "order_line_items" }
