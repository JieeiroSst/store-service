package model

import "time"

type TicketStatus string

const (
	TicketActive    TicketStatus = "ACTIVE"
	TicketCompleted TicketStatus = "COMPLETED"
)

type Ticket struct {
	TicketID     string       `gorm:"column:history_id;primaryKey"`
	VehiclePlate string       `gorm:"column:vehicle_plate"`
	SpotID       string       `gorm:"column:spot_id"`
	EntryGateID  *string      `gorm:"column:entry_gate_id"`
	ExitGateID   *string      `gorm:"column:exit_gate_id"`
	ParkedTime   time.Time    `gorm:"column:parked_time"`
	LeaveTime    *time.Time   `gorm:"column:leave_time"`
	Status       TicketStatus `gorm:"column:status"`
}

func (Ticket) TableName() string { return "parking_history" }

func (t Ticket) Duration() time.Duration {
	if t.LeaveTime == nil {
		return 0
	}
	return t.LeaveTime.Sub(t.ParkedTime)
}
