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

type ShippingAddress struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Street     string `json:"street"`
	Ward       string `json:"ward,omitempty"`
	District   string `json:"district,omitempty"`
	Province   string `json:"province,omitempty"`
	WardCode   string `json:"ward_code"`
	DistrictID int    `json:"district_id"`
	ProvinceID int    `json:"province_id"`
}

func (a *ShippingAddress) Deliverable() bool {
	return a != nil && a.Name != "" && a.Phone != "" && a.Street != "" && a.WardCode != "" && a.DistrictID > 0
}

type ShippingRoute struct {
	WarehouseID int64   `json:"warehouse_id"`
	ServiceID   int     `json:"service_id"`
	Fee         float64 `json:"fee"`
}

type PaidOrderInput struct {
	BasketID          int              `json:"basket_id"`
	UserID            int              `json:"user_id"`
	Currency          string           `json:"currency"`
	Total             float64          `json:"total"`
	ShippingFee       float64          `json:"shipping_fee,omitempty"`
	PaymentReference  string           `json:"payment_reference"`
	ShippingAddressID int              `json:"shipping_address_id"`
	ShippingAddress   *ShippingAddress `json:"shipping_address,omitempty"`
	ShippingRoute     *ShippingRoute   `json:"shipping_route,omitempty"`
	Lines             []PaidOrderLine  `json:"lines"`
}

func (in PaidOrderInput) GoodsTotal() float64 { return max(in.Total-in.ShippingFee, 0) }

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
	case in.ShippingFee < 0 || in.ShippingFee >= in.Total:
		return fmt.Errorf("shipping_fee must be between 0 and the total")
	}
	return nil
}

type PaidOrderRecord struct {
	ID      int    `json:"id"`
	Number  string `json:"number"`
	Created bool   `json:"created"`
}

const (
	ShipmentOutcomeCreated  = "created"
	ShipmentOutcomeRejected = "rejected"
	ShipmentOutcomeSkipped  = "skipped"
)

type ShipmentRequest struct {
	ClientOrderCode string            `json:"client_order_code"`
	Recipient       ShipmentRecipient `json:"recipient"`
	Parcel          ShipmentParcel    `json:"parcel"`
	CODAmount       int64             `json:"cod_amount"`
	InsuranceValue  int64             `json:"insurance_value"`
	PaymentType     string            `json:"payment_type"`
	RequiredNote    string            `json:"required_note"`
	Note            string            `json:"note,omitempty"`
	Strategy        string            `json:"strategy"`
	WarehouseID     int64             `json:"warehouse_id,omitempty"`
	ServiceID       int               `json:"service_id,omitempty"`
	Customer        ShipmentCustomer  `json:"customer"`
}

type ShipmentRecipient struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Street     string `json:"street"`
	WardCode   string `json:"ward_code"`
	DistrictID int    `json:"district_id"`
	ProvinceID int    `json:"province_id"`
}

type ShipmentParcel struct {
	WeightGram int            `json:"weight_gram"`
	LengthCm   int            `json:"length_cm"`
	WidthCm    int            `json:"width_cm"`
	HeightCm   int            `json:"height_cm"`
	Content    string         `json:"content"`
	Items      []ShipmentItem `json:"items"`
}

type ShipmentItem struct {
	Name       string `json:"name"`
	Code       string `json:"code,omitempty"`
	Quantity   int    `json:"quantity"`
	WeightGram int    `json:"weight_gram"`
	Price      int64  `json:"price"`
}

type ShipmentCustomer struct {
	UserID uint `json:"user_id,omitempty"`
}

type Shipment struct {
	ID               int64  `json:"id"`
	Code             string `json:"code"`
	Status           string `json:"status"`
	CarrierOrderCode string `json:"carrier_order_code"`
	ServiceName      string `json:"service_name"`
	QuotedFee        int64  `json:"quoted_fee"`
	LastError        string `json:"last_error"`
}

type ShipmentResult struct {
	Outcome          string `json:"outcome"`
	ShipmentID       int64  `json:"shipment_id,omitempty"`
	Code             string `json:"code,omitempty"`
	Status           string `json:"status,omitempty"`
	CarrierOrderCode string `json:"carrier_order_code,omitempty"`
	Fee              int64  `json:"fee,omitempty"`
	PaidFee          int64  `json:"paid_fee,omitempty"`
	RouteChanged     bool   `json:"route_changed,omitempty"`
	Reason           string `json:"reason,omitempty"`
}

type PaidOrderResult struct {
	OrderID  int             `json:"order_id"`
	Number   string          `json:"number"`
	Status   string          `json:"status"`
	Notified bool            `json:"notified"`
	Shipment *ShipmentResult `json:"shipment,omitempty"`
}
