package model

import "time"

// Status values are persisted as-is in existing rows; "peding" is a
// pre-existing typo kept for data compatibility with rows already written.
const (
	OrderStatusPending = "peding"
	OrderStatusCancel  = "cancel"
	OrderStatusSuccess = "success"
)

type Order struct {
	ID          int         `json:"id,omitempty" form:"id"`
	TableName   string      `json:"table_name,omitempty" form:"table_name"`
	Status      string      `json:"status,omitempty" form:"status"`
	KitchenID   int         `json:"kitchen_id,omitempty"`
	TotalAmount float64     `json:"total_amount,omitempty"`
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	CreatedDate time.Time   `json:"created_date,omitempty"`
	UpdatedTime time.Time   `json:"updated_time,omitempty"`
}

// OrderItem is a line item on an Order. UnitPrice is always filled in by
// order-service from kitchen-service's food catalog (see port.FoodPricer),
// never trusted from the caller — that's what makes Order.TotalAmount a
// server-verified figure instead of whatever a client claims it is.
type OrderItem struct {
	ID        int     `json:"id,omitempty"`
	OrderID   int     `json:"order_id,omitempty"`
	FoodID    int     `json:"food_id"`
	FoodName  string  `json:"food_name,omitempty"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price,omitempty"`
}
