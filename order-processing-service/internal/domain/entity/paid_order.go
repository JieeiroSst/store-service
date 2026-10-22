package entity

import "fmt"

const PaidOrderWorkflowName = "PaidOrderWorkflow"

type PaidOrderLine struct {
	ProductID     int     `json:"product_id"`
	StockrecordID int     `json:"stockrecord_id"`
	Title         string  `json:"title"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
}

type PaidOrderInput struct {
	BasketID          int             `json:"basket_id"`
	UserID            int             `json:"user_id"`
	Currency          string          `json:"currency"`
	Total             float64         `json:"total"`
	PaymentReference  string          `json:"payment_reference"`
	ShippingAddressID int             `json:"shipping_address_id"`
	Lines             []PaidOrderLine `json:"lines"`
}

func (in PaidOrderInput) Number() string { return fmt.Sprintf("KS%08d", in.BasketID) }

func (in PaidOrderInput) Validate() error {
	switch {
	case in.BasketID <= 0:
		return fmt.Errorf("basket_id is required")
	case in.UserID <= 0:
		return fmt.Errorf("user_id is required")
	case in.PaymentReference == "":
		return fmt.Errorf("payment_reference is required for a paid order")
	case len(in.Lines) == 0:
		return fmt.Errorf("order has no lines")
	case in.Total <= 0:
		return fmt.Errorf("total must be positive")
	}
	return nil
}

type PaidOrderRecord struct {
	ID      int    `json:"id"`
	Number  string `json:"number"`
	Created bool   `json:"created"`
}

type PaidOrderResult struct {
	OrderID  int    `json:"order_id"`
	Number   string `json:"number"`
	Status   string `json:"status"`
	Notified bool   `json:"notified"`
}
