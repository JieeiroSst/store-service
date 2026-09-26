package domain

import "time"

type ResaleStatus int16

const (
	ResaleOpen       ResaleStatus = 1
	ResaleProcessing ResaleStatus = 2 // a buyer claimed it and is paying
	ResaleSold       ResaleStatus = 3
	ResaleCancelled  ResaleStatus = 4
)

func (s ResaleStatus) Name() string {
	switch s {
	case ResaleOpen:
		return "open"
	case ResaleProcessing:
		return "processing"
	case ResaleSold:
		return "sold"
	case ResaleCancelled:
		return "cancelled"
	}
	return "unknown"
}

func ParseResaleStatus(s string) ResaleStatus {
	for _, st := range []ResaleStatus{ResaleOpen, ResaleProcessing, ResaleSold, ResaleCancelled} {
		if st.Name() == s {
			return st
		}
	}
	return 0
}

type ResaleListing struct {
	ID         int64
	TicketID   int64
	EventID    int64
	SellerID   int64
	BuyerID    int64
	Price      int64
	FacePrice  int64
	Fee        int64
	Currency   string
	Status     ResaleStatus
	PaymentRef string
	TypeName   string
	SeatLabel  string
	StartsAt   time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func ResaleCeiling(face int64, capPercent int) int64 { return face * int64(capPercent) / 100 }

func ResaleFee(price int64, feePercent int) int64 {
	if feePercent <= 0 {
		return 0
	}
	return price * int64(feePercent) / 100
}
