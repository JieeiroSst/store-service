package model

import "time"

type WebhookEvent struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	WebhookID   string    `json:"webhook_id" gorm:"uniqueIndex"`
	Topic       string    `json:"topic"`
	ShopDomain  string    `json:"shop_domain"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (WebhookEvent) TableName() string { return "webhook_events" }

const (
	TopicProductsCreate  = "products/create"
	TopicProductsUpdate  = "products/update"
	TopicProductsDelete  = "products/delete"
	TopicOrdersCreate    = "orders/create"
	TopicOrdersUpdated   = "orders/updated"
	TopicOrdersPaid      = "orders/paid"
	TopicOrdersCancelled = "orders/cancelled"
	TopicOrdersFulfilled = "orders/fulfilled"
)
