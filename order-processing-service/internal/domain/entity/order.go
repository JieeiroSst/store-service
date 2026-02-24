package entity

import "time"

type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Currency    string      `json:"currency"`
	Status      string      `json:"status"`
	ShippingID  string      `json:"shipping_id,omitempty"`
	PaymentID   string      `json:"payment_id,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type PaymentRequest struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type PaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

type InventoryReserveRequest struct {
	OrderID string      `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

type InventoryReserveResponse struct {
	Reserved bool   `json:"reserved"`
	Message  string `json:"message"`
}

type ShippingRequest struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	Address    Address `json:"address"`
}

type ShippingResponse struct {
	ShippingID   string    `json:"shipping_id"`
	TrackingCode string    `json:"tracking_code"`
	EstimatedAt  time.Time `json:"estimated_at"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

type NotificationRequest struct {
	CustomerID string `json:"customer_id"`
	OrderID    string `json:"order_id"`
	Type       string `json:"type"` // email, sms, push
	Message    string `json:"message"`
}

type NotificationResponse struct {
	Sent bool   `json:"sent"`
	ID   string `json:"id"`
}

type OrderWorkflowInput struct {
	Order   Order   `json:"order"`
	Address Address `json:"address"`
}

type OrderWorkflowResult struct {
	OrderID      string `json:"order_id"`
	Status       string `json:"status"`
	PaymentID    string `json:"payment_id,omitempty"`
	ShippingID   string `json:"shipping_id,omitempty"`
	TrackingCode string `json:"tracking_code,omitempty"`
	Message      string `json:"message,omitempty"`
}

type CronResult struct {
	LastRunTime    time.Time `json:"last_run_time"`
	ProcessedCount int       `json:"processed_count"`
}

type StaleOrder struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
