package model

// PlaceOrder is the request shape for a consumer placing an order: a
// selected menu that the kitchen (via PublishKitchenCreate), the order book
// (via PublishOrderCreated) and accounting (via PublishCartAuthorization)
// all need to hear about. DeliveryName/DeliveryAddress and PaymentMethod
// are optional — an empty value just means accounting-service records an
// empty delivery / defaults the payment method.
type PlaceOrder struct {
	ID              int        `json:"id" form:"id"`
	Name            string     `json:"name"`
	KitchenID       int        `json:"kitchen_id"`
	Menu            []MenuFood `json:"menu" form:"menu"`
	OrderID         int        `json:"order_id"`
	DeliveryName    string     `json:"delivery_name"`
	DeliveryAddress string     `json:"delivery_address"`
	PaymentMethod   string     `json:"payment_method"`
}

// MenuFood.Money is a client-quoted price, kept only for consumer-service's
// own display purposes and for the amount quoted to accounting-service at
// authorization time — order-service does NOT trust it, and instead
// re-prices every item from kitchen-service's catalog (see
// order-service's port.FoodPricer) before persisting Order.TotalAmount.
type MenuFood struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Category Category `json:"category"`
	Money    float64  `json:"money"`
	Quantity int      `json:"quantity"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Order is published on "order.created". It carries line items rather than
// a pre-computed TotalAmount — order-service prices them itself from
// kitchen-service's catalog rather than trusting a client-supplied total.
type Order struct {
	ID        int         `json:"id" form:"id"`
	TableName string      `json:"table_name" form:"table_name"`
	KitchenID int         `json:"kitchen_id"`
	Items     []OrderItem `json:"items"`
}

type OrderItem struct {
	FoodID   int `json:"food_id"`
	Quantity int `json:"quantity"`
}

// CartAuthorization is published on "accounting.authorize" so
// accounting-service can accept/reject the order by operating hours,
// dispatch delivery, and open a payment record. TotalAmount here is the
// client-quoted figure (sum of MenuFood.Money*Quantity) — accounting-service
// receives it independently of order-service's server-verified total, since
// reconciling the two would need a full saga; out of scope for now. Field
// shape matches accounting-service's own model.AuthCart wire format; Status
// is decided by accounting-service itself and left unset here.
type CartAuthorization struct {
	Order         AuthOrder    `json:"order"`
	Delivery      AuthDelivery `json:"delivery"`
	PaymentMethod string       `json:"payment_method"`
}

type AuthOrder struct {
	ID          int     `json:"id"`
	TableName   string  `json:"table_name"`
	KitchenID   int     `json:"kitchen_id"`
	TotalAmount float64 `json:"total_amount"`
}

type AuthDelivery struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	KitchenID int    `json:"kitchen_id"`
}
