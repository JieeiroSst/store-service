package inbound

import (
	"context"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type VenueInput struct {
	Name        string
	City        string
	Address     string
	Description string
	Template    string
	Params      map[string]int
	Map         *domain.VenueMap
}

type SectionCount struct {
	Key        string
	Name       string
	Seats      int
	Accessible int
}

type VenuePreview struct {
	Map      domain.VenueMap
	Seats    []domain.GeneratedSeat
	Sections []SectionCount
	Total    int
	View     domain.SeatMapLayout
}

type SectionAssignment struct {
	Section      string
	TicketTypeID int64
}

type ApplyVenueInput struct {
	VenueID     int64
	Assignments []SectionAssignment
	Replace     bool
}

type ApplyVenueResult struct {
	Seats    int
	Sections []SectionCount
}

type VenueUseCase interface {
	Templates() []domain.VenueTemplate
	Preview(ctx context.Context, in VenueInput) (VenuePreview, error)
	Create(ctx context.Context, actor Principal, in VenueInput) (domain.Venue, error)
	Get(ctx context.Context, actor Principal, id int64) (domain.Venue, error)
	Update(ctx context.Context, actor Principal, id int64, in VenueInput) (domain.Venue, error)
	Delete(ctx context.Context, actor Principal, id int64) error
	List(ctx context.Context, actor Principal, shared bool, afterID int64, limit int) (domain.Page[domain.Venue], error)
	Copy(ctx context.Context, actor Principal, id int64, name string) (domain.Venue, error)
	SetShared(ctx context.Context, actor Principal, id int64, shared bool) (domain.Venue, error)
}
