package model

import "time"

type AssignmentStatus string

const (
	AssignmentPending   AssignmentStatus = "pending"
	AssignmentAccepted  AssignmentStatus = "accepted"
	AssignmentRejected  AssignmentStatus = "rejected"
	AssignmentCompleted AssignmentStatus = "completed"
)

type DriverAssignment struct {
	ID              int64            `gorm:"column:id;primaryKey" json:"id"`
	DriverID        string           `gorm:"column:driver_id" json:"driver_id"`
	OrderID         int64            `gorm:"column:order_id" json:"order_id"`
	Status          AssignmentStatus `gorm:"column:status" json:"status"`
	AssignedAt      time.Time        `gorm:"column:assigned_at" json:"assigned_at"`
	AcceptedAt      *time.Time       `gorm:"column:accepted_at" json:"accepted_at,omitempty"`
	CompletedAt     *time.Time       `gorm:"column:completed_at" json:"completed_at,omitempty"`
	RejectionReason string           `gorm:"column:rejection_reason" json:"rejection_reason,omitempty"`
}

func (DriverAssignment) TableName() string { return "driver_assignments" }
