package model

import "time"

type MovementType string

const (
	MovementReceive  MovementType = "RECEIVE"
	MovementDispense MovementType = "DISPENSE"
	MovementAdjust   MovementType = "ADJUST"
	MovementDispose  MovementType = "DISPOSE"
)

type StockMovement struct {
	ID         int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	MedicineID int64        `json:"medicine_id"`
	BatchID    int64        `json:"batch_id"`
	Type       MovementType `json:"type"`
	Quantity   int          `json:"quantity"`
	Reason     string       `json:"reason"`
	Reference  string       `json:"reference"`
	CreatedBy  string       `json:"created_by"`
	CreatedAt  time.Time    `json:"created_at"`
}

func (StockMovement) TableName() string { return "stock_movements" }
