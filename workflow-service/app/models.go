package app

type Order struct {
	ID             string  `json:"id"`
	CustomerID     string  `json:"customer_id"`
	Amount         float64 `json:"amount"`
	Status         string  `json:"status"`
	PaymentID      string  `json:"payment_id,omitempty"`
	TrackingNumber string  `json:"tracking_number,omitempty"`
}
