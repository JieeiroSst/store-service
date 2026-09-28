package model

import (
	"sort"
	"time"
)

type Batch struct {
	ID               int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	MedicineID       int64     `json:"medicine_id"`
	BatchNumber      string    `json:"batch_number"`
	ExpiryDate       Date      `json:"expiry_date" gorm:"type:date"`
	Quantity         int       `json:"quantity"`
	ReceivedQuantity int       `json:"received_quantity"`
	UnitCost         int64     `json:"unit_cost"`
	Supplier         string    `json:"supplier"`
	ReceivedAt       time.Time `json:"received_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Batch) TableName() string { return "batches" }

func (b Batch) UsableOn(today Date) bool {
	return b.Quantity > 0 && b.ExpiryDate > today
}

type Allocation struct {
	Batch    Batch
	Quantity int
}

func AllocateFEFO(batches []Batch, quantity int, today Date) (allocs []Allocation, shortfall int) {
	usable := make([]Batch, 0, len(batches))
	for _, b := range batches {
		if b.UsableOn(today) {
			usable = append(usable, b)
		}
	}
	sort.SliceStable(usable, func(i, j int) bool {
		if usable[i].ExpiryDate != usable[j].ExpiryDate {
			return usable[i].ExpiryDate < usable[j].ExpiryDate
		}
		return usable[i].ID < usable[j].ID
	})

	remaining := quantity
	for _, b := range usable {
		if remaining == 0 {
			break
		}
		take := min(b.Quantity, remaining)
		allocs = append(allocs, Allocation{Batch: b, Quantity: take})
		remaining -= take
	}
	return allocs, remaining
}

type StockSummary struct {
	MedicineID int64   `json:"medicine_id"`
	Available  int     `json:"available"`
	Expired    int     `json:"expired"`
	Batches    []Batch `json:"batches"`
}

func SummarizeStock(medicineID int64, batches []Batch, today Date) StockSummary {
	s := StockSummary{MedicineID: medicineID, Batches: batches}
	for _, b := range batches {
		if b.UsableOn(today) {
			s.Available += b.Quantity
		} else if b.Quantity > 0 {
			s.Expired += b.Quantity
		}
	}
	return s
}
