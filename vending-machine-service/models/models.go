package models

import (
	"time"
)

type Machine struct {
	ID              string    `db:"machine_id"`
	Location        string    `db:"location"`
	Model           string    `db:"model"`
	Status          string    `db:"status"`
	LastMaintenance time.Time `db:"last_maintenance_at"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type Category struct {
	ID           string    `db:"category_id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	DisplayOrder int       `db:"display_order"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Product struct {
	ID          string    `db:"product_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	PriceCents  int       `db:"price_cents"`
	CategoryID  string    `db:"category_id"`
	ImageURL    string    `db:"image_url"`
	Barcode     string    `db:"barcode"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	Category   *Category          `db:"-"`
	Attributes []ProductAttribute `db:"-"`
}

type ProductAttribute struct {
	ID        string `db:"attribute_id"`
	ProductID string `db:"product_id"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}

type Inventory struct {
	ID             string    `db:"inventory_id"`
	MachineID      string    `db:"machine_id"`
	ProductID      string    `db:"product_id"`
	SlotIdentifier string    `db:"slot_identifier"`
	Quantity       int       `db:"quantity"`
	MaxCapacity    int       `db:"max_capacity"`
	LowThreshold   int       `db:"low_threshold"`
	LastRestocked  time.Time `db:"last_restocked_at"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`

	Product *Product `db:"-"`
	Machine *Machine `db:"-"`
}

type Session struct {
	ID        string    `db:"session_id"`
	MachineID string    `db:"machine_id"`
	StartedAt time.Time `db:"started_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Machine *Machine `db:"-"`
}

type Reservation struct {
	ID          string    `db:"reservation_id"`
	SessionID   string    `db:"session_id"`
	InventoryID string    `db:"inventory_id"`
	ExpiresAt   time.Time `db:"expires_at"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	Session   *Session   `db:"-"`
	Inventory *Inventory `db:"-"`
}

type Payment struct {
	ID              string    `db:"payment_id"`
	SessionID       string    `db:"session_id"`
	AmountCents     int       `db:"amount_cents"`
	Currency        string    `db:"currency"`
	PaymentMethod   string    `db:"payment_method"`
	PaymentStatus   string    `db:"payment_status"`
	TransactionID   string    `db:"transaction_id"`
	PaymentMetadata string    `db:"payment_metadata"`
	CompletedAt     time.Time `db:"completed_at"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`

	Session *Session `db:"-"`
}

type Order struct {
	ID            string    `db:"order_id"`
	SessionID     string    `db:"session_id"`
	ReservationID string    `db:"reservation_id"`
	PaymentID     string    `db:"payment_id"`
	Status        string    `db:"status"`
	FulfilledAt   time.Time `db:"fulfilled_at"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`

	Session     *Session     `db:"-"`
	Reservation *Reservation `db:"-"`
	Payment     *Payment     `db:"-"`
}

type Event struct {
	ID            string    `db:"event_id"`
	EventType     string    `db:"event_type"`
	RelatedEntity string    `db:"related_entity"`
	EntityID      string    `db:"entity_id"`
	MachineID     string    `db:"machine_id"`
	Data          string    `db:"data"`
	OccurredAt    time.Time `db:"occurred_at"`
}

type MaintenanceLog struct {
	ID              string    `db:"log_id"`
	MachineID       string    `db:"machine_id"`
	TechnicianID    string    `db:"technician_id"`
	MaintenanceType string    `db:"maintenance_type"`
	Notes           string    `db:"notes"`
	PerformedAt     time.Time `db:"performed_at"`

	Machine *Machine `db:"-"`
}
