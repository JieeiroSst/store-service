package model

import "time"

const (
	ReservationStatusBooked    = "booked"
	ReservationStatusCancelled = "cancelled"
	ReservationStatusCompleted = "completed"
)

// Reservation books a dine-in table ahead of an order being placed against
// it (Order.TableName references the same table naming).
type Reservation struct {
	ID           int       `json:"id,omitempty"`
	TableName    string    `json:"table_name" form:"table_name"`
	CustomerName string    `json:"customer_name" form:"customer_name"`
	PartySize    int       `json:"party_size" form:"party_size"`
	ReservedAt   time.Time `json:"reserved_at" form:"reserved_at"`
	Status       string    `json:"status,omitempty"`
	CreatedDate  time.Time `json:"created_date,omitempty"`
}
