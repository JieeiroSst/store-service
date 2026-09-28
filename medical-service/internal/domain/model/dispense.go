package model

import "time"

type Dispense struct {
	ID              int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	PatientRef      string         `json:"patient_ref" gorm:"index"`
	PrescriptionRef string         `json:"prescription_ref"`
	DispensedBy     string         `json:"dispensed_by"`
	OverrideReason  string         `json:"override_reason"`
	Warnings        []Warning      `json:"warnings" gorm:"serializer:json;type:text"`
	Items           []DispenseItem `json:"items" gorm:"foreignKey:DispenseID"`
	CreatedAt       time.Time      `json:"created_at"`
}

func (Dispense) TableName() string { return "dispenses" }

type DispenseItem struct {
	ID          int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	DispenseID  int64  `json:"dispense_id"`
	MedicineID  int64  `json:"medicine_id"`
	BatchID     int64  `json:"batch_id"`
	BatchNumber string `json:"batch_number"`
	ExpiryDate  Date   `json:"expiry_date" gorm:"type:date"`
	Quantity    int    `json:"quantity"`
}

func (DispenseItem) TableName() string { return "dispense_items" }
