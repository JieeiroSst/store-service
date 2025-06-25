package model

import "time"

type CustomerInfo struct {
	CifID        string    `json:"cif_id"`
	FullName     string    `json:"full_name"`
	DateOfBirth  string    `json:"date_of_birth"`
	Gender       string    `json:"gender"`
	Nationality  string    `json:"nationality"`
	PlaceOfBirth string    `json:"place_of_birth"`
	Address      string    `json:"address"`
	IssueDate    string    `json:"issue_date"`
	ExpiryDate   string    `json:"expiry_date"`
	IsDetected   bool      `json:"is_detected"`
	Confidence   float64   `json:"confidence"`
	ProcessedAt  time.Time `json:"processed_at"`
	PhoneNumber  string    `json:"phone_number"`
}
