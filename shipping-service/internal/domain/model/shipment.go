package model

import (
	"errors"
	"strings"
	"time"
)

type PaymentType string

const (
	PaymentBySender    PaymentType = "SENDER"
	PaymentByRecipient PaymentType = "RECIPIENT"
)

type RequiredNote string

const (
	NoteAllowTry         RequiredNote = "CHOTHUHANG"
	NoteAllowViewNotTry  RequiredNote = "CHOXEMHANGKHONGTHU"
	NoteNotAllowView     RequiredNote = "KHONGCHOXEMHANG"
	DefaultRequiredNote               = NoteAllowViewNotTry
	DefaultPaymentType                = PaymentBySender
	DefaultRouteStrategy              = StrategyBalanced
)

func (n RequiredNote) Valid() bool {
	return n == NoteAllowTry || n == NoteAllowViewNotTry || n == NoteNotAllowView
}

func (p PaymentType) Valid() bool { return p == PaymentBySender || p == PaymentByRecipient }

type Customer struct {
	UserID uint   `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}

func (c Customer) Normalize() (Customer, error) {
	c.Email = strings.TrimSpace(c.Email)
	if c.Email != "" && (!strings.Contains(c.Email, "@") || strings.ContainsAny(c.Email, " \t\r\n")) {
		return c, errors.New("customer email is invalid")
	}
	return c, nil
}

type Shipment struct {
	ID                 int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	Code               string       `json:"code" gorm:"uniqueIndex"`
	ClientService      string       `json:"client_service"`
	ClientOrderCode    string       `json:"client_order_code"`
	Status             Status       `json:"status"`
	CarrierStatus      string       `json:"carrier_status"`
	CarrierOrderCode   string       `json:"carrier_order_code"`
	WarehouseID        int64        `json:"warehouse_id"`
	CarrierShopID      int64        `json:"carrier_shop_id"`
	ServiceID          int          `json:"service_id"`
	ServiceTypeID      int          `json:"service_type_id"`
	ServiceName        string       `json:"service_name"`
	Strategy           Strategy     `json:"strategy"`
	Recipient          Address      `json:"recipient" gorm:"serializer:json;type:text"`
	Parcel             Parcel       `json:"parcel" gorm:"serializer:json;type:text"`
	Customer           Customer     `json:"customer" gorm:"serializer:json;type:text"`
	CODAmount          int64        `json:"cod_amount"`
	InsuranceValue     int64        `json:"insurance_value"`
	PaymentType        PaymentType  `json:"payment_type"`
	RequiredNote       RequiredNote `json:"required_note"`
	Note               string       `json:"note"`
	QuotedFee          int64        `json:"quoted_fee"`
	ShippingFee        int64        `json:"shipping_fee"`
	ExpectedDeliveryAt *time.Time   `json:"expected_delivery_at"`
	DeliveredAt        *time.Time   `json:"delivered_at"`
	CallbackURL        string       `json:"callback_url"`
	PlaceAttempts      int          `json:"place_attempts"`
	PlacingUntil       *time.Time   `json:"-"`
	LastError          string       `json:"last_error"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

func (Shipment) TableName() string { return "shipments" }

type ShipmentEvent struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ShipmentID    int64     `json:"shipment_id"`
	Source        string    `json:"source"`
	FromStatus    Status    `json:"from_status"`
	ToStatus      Status    `json:"to_status"`
	CarrierStatus string    `json:"carrier_status"`
	Reason        string    `json:"reason"`
	Applied       bool      `json:"applied"`
	OccurredAt    time.Time `json:"occurred_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func (ShipmentEvent) TableName() string { return "shipment_events" }

const (
	SourceAPI     = "api"
	SourceCarrier = "carrier"
	SourceWebhook = "webhook"
	SourceSync    = "sync"
)
