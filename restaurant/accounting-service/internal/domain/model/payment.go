package model

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"

	PaymentMethodCash = "cash"
)

// Payment opens when an order is accepted (see authCartService.acceptOrder)
// and is settled later via PaymentUsecase.MarkPaid — a real restaurant
// keeps a tab open until the customer actually pays at checkout.
type Payment struct {
	ID      int     `json:"id,omitempty"`
	OrderID int     `json:"order_id"`
	Amount  float64 `json:"amount"`
	Method  string  `json:"method"`
	Status  string  `json:"status"`
}
