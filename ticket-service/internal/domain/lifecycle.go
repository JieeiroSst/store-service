package domain

import (
	"strings"
	"time"
)

type TicketEvent struct {
	ID         int64
	TicketID   int64
	Event      string
	FromStatus TicketStatus
	ToStatus   TicketStatus
	Actor      int64
	Note       string
	CreatedAt  time.Time
}

type TransferStatus int16

const (
	TransferPending   TransferStatus = 1
	TransferAccepted  TransferStatus = 2
	TransferDeclined  TransferStatus = 3
	TransferCancelled TransferStatus = 4
	TransferExpired   TransferStatus = 5
)

func (s TransferStatus) Name() string {
	switch s {
	case TransferPending:
		return "pending"
	case TransferAccepted:
		return "accepted"
	case TransferDeclined:
		return "declined"
	case TransferCancelled:
		return "cancelled"
	case TransferExpired:
		return "expired"
	}
	return "unknown"
}

type Transfer struct {
	ID         int64
	TicketID   int64
	FromUser   int64
	ToUser     int64
	ToEmail    string
	Status     TransferStatus
	Message    string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	ResolvedAt *time.Time
	Ticket     *Ticket
}

func (t Transfer) AddressedTo(userID int64, email string) bool {
	if t.ToUser != 0 && t.ToUser == userID {
		return true
	}
	return t.ToEmail != "" && email != "" && strings.EqualFold(t.ToEmail, email)
}

type StaffMember struct {
	EventID   int64
	UserID    int64
	AddedBy   int64
	CreatedAt time.Time
}

type WaitlistEntry struct {
	TicketTypeID int64
	EventID      int64
	EventTitle   string
	TypeName     string
	CreatedAt    time.Time
}
