package model

import "time"

type OrderStatus string

const (
	OrderStatusCreated        OrderStatus = "created"
	OrderStatusConfirmed      OrderStatus = "confirmed"
	OrderStatusPreparing      OrderStatus = "preparing"
	OrderStatusReadyForPickup OrderStatus = "ready_for_pickup"
	OrderStatusPickedUp       OrderStatus = "picked_up"
	OrderStatusDelivered      OrderStatus = "delivered"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

func (s OrderStatus) Valid() bool {
	switch s {
	case OrderStatusCreated, OrderStatusConfirmed, OrderStatusPreparing, OrderStatusReadyForPickup,
		OrderStatusPickedUp, OrderStatusDelivered, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func (s OrderStatus) Terminal() bool {
	return s == OrderStatusDelivered || s == OrderStatusCancelled
}

var orderStatusTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusCreated:        {OrderStatusConfirmed, OrderStatusCancelled},
	OrderStatusConfirmed:      {OrderStatusPreparing, OrderStatusCancelled},
	OrderStatusPreparing:      {OrderStatusReadyForPickup, OrderStatusCancelled},
	OrderStatusReadyForPickup: {OrderStatusPickedUp, OrderStatusCancelled},
	OrderStatusPickedUp:       {OrderStatusDelivered},
	OrderStatusDelivered:      {},
	OrderStatusCancelled:      {},
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	for _, allowed := range orderStatusTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

const (
	PaymentStatusPending    = "pending"
	PaymentStatusAuthorized = "authorized"
	PaymentStatusCaptured   = "captured"
	PaymentStatusRefunded   = "refunded"
	PaymentStatusFailed     = "failed"
)

type Order struct {
	ID                    int64       `gorm:"column:id;primaryKey" json:"id"`
	CustomerID            string      `gorm:"column:customer_id" json:"customer_id"`
	RestaurantID          string      `gorm:"column:restaurant_id" json:"restaurant_id"`
	DriverID              *string     `gorm:"column:driver_id" json:"driver_id,omitempty"`
	DeliveryAddressID     string      `gorm:"column:delivery_address_id" json:"delivery_address_id"`
	Status                OrderStatus `gorm:"column:status" json:"status"`
	PlacedAt              time.Time   `gorm:"column:placed_at" json:"placed_at"`
	EstimatedDeliveryTime *time.Time  `gorm:"column:estimated_delivery_time" json:"estimated_delivery_time,omitempty"`
	ActualDeliveryTime    *time.Time  `gorm:"column:actual_delivery_time" json:"actual_delivery_time,omitempty"`
	Subtotal              float64     `gorm:"column:subtotal" json:"subtotal"`
	DeliveryFee           float64     `gorm:"column:delivery_fee" json:"delivery_fee"`
	ServiceFee            float64     `gorm:"column:service_fee" json:"service_fee"`
	Tax                   float64     `gorm:"column:tax" json:"tax"`
	Tip                   float64     `gorm:"column:tip" json:"tip"`
	TotalAmount           float64     `gorm:"column:total_amount" json:"total_amount"`
	PaymentMethodID       string      `gorm:"column:payment_method_id" json:"payment_method_id"`
	PaymentStatus         string      `gorm:"column:payment_status" json:"payment_status"`
	SpecialInstructions   string      `gorm:"column:special_instructions" json:"special_instructions,omitempty"`
	CreatedAt             time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt             time.Time   `gorm:"column:updated_at" json:"updated_at"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string { return "orders" }

func (o *Order) Recalculate() {
	var subtotal float64
	for _, item := range o.Items {
		subtotal += item.LineTotal()
	}
	o.Subtotal = subtotal
	o.TotalAmount = o.Subtotal + o.DeliveryFee + o.ServiceFee + o.Tax + o.Tip
}

type OrderItem struct {
	ID                  int64     `gorm:"column:id;primaryKey" json:"id"`
	OrderID             int64     `gorm:"column:order_id" json:"order_id"`
	MenuItemID          string    `gorm:"column:menu_item_id" json:"menu_item_id"`
	Name                string    `gorm:"column:name" json:"name"`
	Quantity            int       `gorm:"column:quantity" json:"quantity"`
	UnitPrice           float64   `gorm:"column:unit_price" json:"unit_price"`
	Customizations      string    `gorm:"column:customizations" json:"customizations,omitempty"`
	SpecialInstructions string    `gorm:"column:special_instructions" json:"special_instructions,omitempty"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"created_at"`
}

func (OrderItem) TableName() string { return "order_items" }

func (i OrderItem) LineTotal() float64 {
	return i.UnitPrice * float64(i.Quantity)
}
