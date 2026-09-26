package inbound

import (
	"context"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type TicketListQuery struct {
	EventID int64
	Status  domain.TicketStatus
	AfterID int64
	Limit   int
}

type HolderInput struct {
	Name  string
	Email string
}

type TransferCommand struct {
	ToUserID int64
	ToEmail  string
	Message  string
}

type TicketUseCase interface {
	List(ctx context.Context, actor Principal, q TicketListQuery) (domain.Page[domain.Ticket], error)
	Get(ctx context.Context, actor Principal, id int64) (domain.Ticket, error)
	History(ctx context.Context, actor Principal, id int64) ([]domain.TicketEvent, error)
	SetHolder(ctx context.Context, actor Principal, id int64, in HolderInput) (domain.Ticket, error)
	Reissue(ctx context.Context, actor Principal, id int64) (domain.Ticket, error)
	Offer(ctx context.Context, actor Principal, ticketID int64, c TransferCommand) (domain.Transfer, error)
	Transfers(ctx context.Context, actor Principal, incoming bool, afterID int64, limit int) (domain.Page[domain.Transfer], error)
	AcceptTransfer(ctx context.Context, actor Principal, transferID int64) (domain.Ticket, error)
	DeclineTransfer(ctx context.Context, actor Principal, transferID int64) (domain.Transfer, error)
	CancelTransfer(ctx context.Context, actor Principal, transferID int64) (domain.Transfer, error)

	Void(ctx context.Context, actor Principal, id int64, reason string) (domain.Ticket, error)
	RevertCheckIn(ctx context.Context, actor Principal, id int64) (domain.Ticket, error)
	Lookup(ctx context.Context, actor Principal, eventID int64, code string) (domain.Ticket, error)

	ExpireTickets(ctx context.Context) (int, error)
	ExpireTransfers(ctx context.Context) (int, error)
}

type ScanItem struct {
	Code      string
	ScannedAt *time.Time
}

type ScanResult struct {
	Code   string
	Result string
	Ticket domain.Ticket
}

type StaffUseCase interface {
	Add(ctx context.Context, actor Principal, eventID, userID int64) error
	Remove(ctx context.Context, actor Principal, eventID, userID int64) error
	List(ctx context.Context, actor Principal, eventID int64) ([]domain.StaffMember, error)
	Events(ctx context.Context, actor Principal) ([]domain.Event, error)
}

type WaitlistUseCase interface {
	Join(ctx context.Context, actor Principal, eventID, typeID int64) error
	Leave(ctx context.Context, actor Principal, typeID int64) error
	Mine(ctx context.Context, actor Principal) ([]domain.WaitlistEntry, error)
}

type ResaleQuery struct {
	EventID       int64
	TypeID        int64
	CheapestFirst bool
	AfterID       int64
	Limit         int
}

type ResaleUseCase interface {
	Browse(ctx context.Context, q ResaleQuery) (domain.Page[domain.ResaleListing], error)
	Mine(ctx context.Context, actor Principal, status domain.ResaleStatus, afterID int64, limit int) (domain.Page[domain.ResaleListing], error)
	List(ctx context.Context, actor Principal, ticketID, price int64) (domain.ResaleListing, error)
	Cancel(ctx context.Context, actor Principal, listingID int64) (domain.ResaleListing, error)
	Buy(ctx context.Context, actor Principal, listingID int64) (domain.Ticket, error)
	Recover(ctx context.Context) (int, error)
}
