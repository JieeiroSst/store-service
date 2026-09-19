package model

// Order.Status is persisted using the NATS subject name it was decided
// under ("order.success" / "order.reject"), matching the pre-existing
// stored values.
const (
	OrderStatusSuccess = "order.success"
	OrderStatusReject  = "order.reject"
)

type Order struct {
	ID          int     `json:"id" form:"id"`
	TableName   string  `json:"table_name" form:"table_name"`
	Status      string  `json:"status" form:"status"`
	KitchenID   int     `json:"kitchen_id"`
	TotalAmount float64 `json:"total_amount" form:"total_amount"`
}

type Delivery struct {
	Name      string `json:"name" form:"name"`
	Address   string `json:"address" form:"address"`
	KitchenID int    `json:"kitchen_id"`
}

type AuthCart struct {
	Order         Order    `json:"order" form:"order"`
	Delivery      Delivery `json:"delivery" form:"delivery"`
	PaymentMethod string   `json:"payment_method" form:"payment_method"`
}
