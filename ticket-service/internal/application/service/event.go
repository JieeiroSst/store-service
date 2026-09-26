package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

const (
	maxSeatsPerCall = 20_000
	maxSeatRows     = 200
	maxTicketTypes  = 50
	seatPitchX      = 30
	seatPitchY      = 34
)

type EventService struct {
	events  outbound.EventRepository
	tickets outbound.TicketRepository
	venues  outbound.VenueRepository
	orders  outbound.OrderRepository
	notify  *Notifier
	views   *ttlCache[int64, domain.Event]
	waiters TicketsBack
	now     func() time.Time
}

var _ inbound.EventUseCase = (*EventService)(nil)

func (s *EventService) WithNotices(orders outbound.OrderRepository, n *Notifier) *EventService {
	s.orders, s.notify = orders, n
	return s
}

func (s *EventService) WithWaitlist(w TicketsBack) *EventService {
	s.waiters = w
	return s
}

func NewEventService(events outbound.EventRepository, tickets outbound.TicketRepository, viewTTL time.Duration) *EventService {
	return &EventService{events: events, tickets: tickets, views: newTTLCache[int64, domain.Event](viewTTL, 10_000), now: time.Now}
}

func (s *EventService) validate(in inbound.EventInput, creating bool) (inbound.EventInput, error) {
	in.Title, in.Description, in.City, in.Venue = trim(in.Title), trim(in.Description), trim(in.City), trim(in.Venue)
	in.Address, in.BannerURL, in.WalletID = trim(in.Address), trim(in.BannerURL), trim(in.WalletID)
	in.Category = strings.ToLower(trim(in.Category))
	in.Currency = strings.ToUpper(trim(in.Currency))
	if in.Currency == "" {
		in.Currency = "VND"
	}
	switch {
	case len(in.Title) < 3 || len(in.Title) > 200:
		return in, invalid("title must be 3-200 characters")
	case len(in.Description) > 20_000:
		return in, invalid("description is too long")
	case !domain.ValidCategory(in.Category):
		return in, invalid("category must be one of %s", strings.Join(domain.Categories, ", "))
	case in.City == "" || in.Venue == "":
		return in, invalid("city and venue are required")
	case len(in.Currency) != 3:
		return in, invalid("currency must be a 3-letter code")
	case !in.EndsAt.After(in.StartsAt):
		return in, invalid("the event must end after it starts")
	case creating && !in.StartsAt.After(s.now()):
		return in, invalid("the event must start in the future")
	case in.ResaleCapPercent < 0 || in.ResaleCapPercent > 300:
		return in, invalid("resale_cap_percent must be between 0 and 300")
	case in.RefundCutoffHours < 0 || in.RefundCutoffHours > 24*365:
		return in, invalid("refund_cutoff_hours must be between 0 and %d", 24*365)
	}
	return in, nil
}

func (s *EventService) load(ctx context.Context, actor inbound.Principal, id int64) (domain.Event, error) {
	e, err := s.events.Get(ctx, id)
	if err != nil {
		return e, err
	}
	if !canManage(actor, e) {
		return domain.Event{}, notYours(e)
	}
	return e, nil
}

func (s *EventService) changed(id int64) { s.views.Forget(id) }

func (s *EventService) Create(ctx context.Context, actor inbound.Principal, in inbound.EventInput) (domain.Event, error) {
	in, err := s.validate(in, true)
	if err != nil {
		return domain.Event{}, err
	}
	return s.events.Create(ctx, domain.Event{
		OrganizerID: actor.UserID, Title: in.Title, Description: in.Description, Category: in.Category, City: in.City,
		Venue: in.Venue, Address: in.Address, BannerURL: in.BannerURL, StartsAt: in.StartsAt, EndsAt: in.EndsAt,
		Currency: in.Currency, WalletID: in.WalletID, RefundCutoffHours: in.RefundCutoffHours, Transferable: in.Transferable, ResaleCapPercent: in.ResaleCapPercent, Status: domain.EventDraft,
	})
}

func (s *EventService) Update(ctx context.Context, actor inbound.Principal, id int64, in inbound.EventInput) (domain.Event, error) {
	e, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Event{}, err
	}
	if e.Status == domain.EventCancelled {
		return domain.Event{}, fmt.Errorf("%w: a cancelled event cannot be edited", domain.ErrConflict)
	}
	in, err = s.validate(in, false)
	if err != nil {
		return domain.Event{}, err
	}
	if e.Status != domain.EventDraft && in.Currency != e.Currency {
		return domain.Event{}, invalid("the currency cannot change once the event has been submitted")
	}
	multi := len(e.Sessions) > 1
	if multi && (!in.StartsAt.Equal(e.StartsAt) || !in.EndsAt.Equal(e.EndsAt)) {
		return domain.Event{}, invalid("this event has %d showtimes: change their dates with the session endpoints (the event's dates are their first start and last end)", len(e.Sessions))
	}
	if !multi && e.Status == domain.EventPublished && !in.EndsAt.After(s.now()) {
		return domain.Event{}, invalid("the event cannot end in the past")
	}
	var moved *domain.Session
	if !multi && len(e.Sessions) == 1 && (!in.StartsAt.Equal(e.StartsAt) || !in.EndsAt.Equal(e.EndsAt)) {
		sess := e.Sessions[0]
		moved = &sess
	}
	e.Title, e.Description, e.Category, e.City, e.Venue = in.Title, in.Description, in.Category, in.City, in.Venue
	e.Address, e.BannerURL, e.StartsAt, e.EndsAt = in.Address, in.BannerURL, in.StartsAt, in.EndsAt
	e.Currency, e.WalletID, e.RefundCutoffHours, e.Transferable = in.Currency, in.WalletID, in.RefundCutoffHours, in.Transferable
	e.ResaleCapPercent = in.ResaleCapPercent
	out, err := s.events.Update(ctx, e)
	if err == nil {
		s.changed(id)
		if moved != nil {
			s.tellMoved(ctx, out, *moved, in.StartsAt)
		}
	}
	return out, err
}

func (s *EventService) tellMoved(ctx context.Context, e domain.Event, old domain.Session, newStart time.Time) {
	if s.orders == nil || s.notify == nil || old.StartsAt.Equal(newStart) {
		return
	}
	buyers, err := s.orders.BuyersOfSession(ctx, old.ID, 5000)
	if err != nil {
		return
	}
	body := fmt.Sprintf("%s was moved from %s to %s. Your tickets stay valid.", e.Title, s.notify.Format(old.StartsAt), s.notify.Format(newStart))
	for _, u := range buyers {
		s.notify.Send(ctx, u, "showtime_moved", "The date of your event changed", body, 0, e.ID)
	}
}

func (s *EventService) Duplicate(ctx context.Context, actor inbound.Principal, id int64, in inbound.DuplicateInput) (domain.Event, error) {
	if _, err := s.load(ctx, actor, id); err != nil {
		return domain.Event{}, err
	}
	in.Title = trim(in.Title)
	switch {
	case in.Title != "" && (len(in.Title) < 3 || len(in.Title) > 200):
		return domain.Event{}, invalid("title must be 3-200 characters")
	case !in.EndsAt.After(in.StartsAt):
		return domain.Event{}, invalid("the event must end after it starts")
	case !in.StartsAt.After(s.now()):
		return domain.Event{}, invalid("the new showtime must start in the future")
	}
	return s.events.Duplicate(ctx, id, outbound.DuplicateSpec{Title: in.Title, StartsAt: in.StartsAt, EndsAt: in.EndsAt, CopyPromotions: in.CopyPromotions})
}

func (s *EventService) status(ctx context.Context, id int64, from []domain.EventStatus, to domain.EventStatus, note string) (domain.Event, error) {
	e, err := s.events.SetStatus(ctx, id, from, to, note)
	if errors.Is(err, domain.ErrConflict) {
		return domain.Event{}, fmt.Errorf("%w: the event changed while you were acting on it; look at it again", domain.ErrConflict)
	}
	if err == nil {
		s.changed(id)
	}
	return e, err
}

func (s *EventService) Submit(ctx context.Context, actor inbound.Principal, id int64) (domain.Event, error) {
	e, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Event{}, err
	}
	active := 0
	for _, t := range e.TicketTypes {
		if t.Active && t.Total > 0 {
			active++
		}
	}
	if active == 0 {
		return domain.Event{}, invalid("add at least one ticket type with tickets before submitting")
	}
	if !e.EndsAt.After(s.now()) {
		return domain.Event{}, invalid("the event is already over")
	}
	return s.status(ctx, id, []domain.EventStatus{domain.EventDraft, domain.EventRejected}, domain.EventPending, "")
}

func (s *EventService) Approve(ctx context.Context, actor inbound.Principal, id int64) (domain.Event, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Event{}, err
	}
	return s.status(ctx, id, []domain.EventStatus{domain.EventPending}, domain.EventPublished, "")
}

func (s *EventService) Reject(ctx context.Context, actor inbound.Principal, id int64, note string) (domain.Event, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Event{}, err
	}
	if note = trim(note); note == "" {
		return domain.Event{}, invalid("tell the organizer why the event was rejected")
	}
	return s.status(ctx, id, []domain.EventStatus{domain.EventPending}, domain.EventRejected, note)
}

func (s *EventService) Cancel(ctx context.Context, actor inbound.Principal, id int64, reason string) (domain.Event, error) {
	e, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Event{}, err
	}
	reason = trim(reason)
	if e.Status == domain.EventPublished && reason == "" {
		return domain.Event{}, invalid("tell ticket holders why the event was cancelled")
	}
	return s.status(ctx, id, []domain.EventStatus{domain.EventDraft, domain.EventPending, domain.EventRejected, domain.EventPublished}, domain.EventCancelled, reason)
}

func (s *EventService) SetFeatured(ctx context.Context, actor inbound.Principal, id int64, featured bool) (domain.Event, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Event{}, err
	}
	if err := s.events.SetFeatured(ctx, id, featured); err != nil {
		return domain.Event{}, err
	}
	s.changed(id)
	return s.events.Get(ctx, id)
}

func (s *EventService) View(ctx context.Context, viewer *inbound.Principal, id int64) (domain.Event, error) {
	e, err := s.views.Get(id, func() (domain.Event, error) { return s.events.Get(ctx, id) })
	if err != nil {
		return domain.Event{}, err
	}
	managed := viewer != nil && canManage(*viewer, e)
	if e.Status != domain.EventPublished && e.Status != domain.EventCancelled && !managed {
		return domain.Event{}, domain.ErrNotFound
	}
	if e.Status == domain.EventCancelled && !managed {
		e.TicketTypes = nil
	}
	if !managed {
		visible := make([]domain.TicketType, 0, len(e.TicketTypes))
		for _, t := range e.TicketTypes {
			if t.Active {
				visible = append(visible, t)
			}
		}
		e.TicketTypes = visible
		e.WalletID = ""
		e.VenueID = 0
	}
	e.SoldOut = allSoldOut(e.TicketTypes)
	return e, nil
}

func allSoldOut(types []domain.TicketType) bool {
	n := 0
	for _, t := range types {
		if !t.Active {
			continue
		}
		n++
		if t.Available > 0 {
			return false
		}
	}
	return n > 0
}

func (s *EventService) ListMine(ctx context.Context, actor inbound.Principal, afterID int64, limit int) (domain.Page[domain.Event], error) {
	limit = page(limit)
	rows, err := s.events.ListByOrganizer(ctx, actor.UserID, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.Event]{}, err
	}
	return domain.PageOf(rows, limit, func(e domain.Event) int64 { return e.ID }), nil
}

func (s *EventService) ReviewQueue(ctx context.Context, actor inbound.Principal, afterID int64, limit int) (domain.Page[domain.Event], error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Page[domain.Event]{}, err
	}
	limit = page(limit)
	rows, err := s.events.ListByStatus(ctx, domain.EventPending, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.Event]{}, err
	}
	return domain.PageOf(rows, limit, func(e domain.Event) int64 { return e.ID }), nil
}

// ---- ticket types

func (s *EventService) validateType(in inbound.TicketTypeInput) (inbound.TicketTypeInput, error) {
	in.Name, in.Description = trim(in.Name), trim(in.Description)
	if in.MinPerOrder == 0 {
		in.MinPerOrder = 1
	}
	if in.MaxPerOrder == 0 {
		in.MaxPerOrder = 10
	}
	switch {
	case in.Name == "" || len(in.Name) > 100:
		return in, invalid("ticket type name must be 1-100 characters")
	case in.Price < 0:
		return in, invalid("price cannot be negative")
	case !in.Seated && in.Total < 1:
		return in, invalid("total must be at least 1")
	case in.Total < 0 || in.Total > 10_000_000:
		return in, invalid("total is out of range")
	case in.MinPerOrder < 1 || in.MaxPerOrder < in.MinPerOrder || in.MaxPerOrder > 100:
		return in, invalid("per-order limits must satisfy 1 <= min <= max <= 100")
	case in.MaxPerUser < 0 || in.MaxPerUser > 1000:
		return in, invalid("max_per_user must be between 0 and 1000")
	case in.SaleStartsAt != nil && in.SaleEndsAt != nil && !in.SaleEndsAt.After(*in.SaleStartsAt):
		return in, invalid("ticket sales must end after they start")
	}
	return in, nil
}

func (s *EventService) CreateTicketType(ctx context.Context, actor inbound.Principal, eventID int64, in inbound.TicketTypeInput) (domain.TicketType, error) {
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return domain.TicketType{}, err
	}
	if e.Status == domain.EventCancelled {
		return domain.TicketType{}, fmt.Errorf("%w: a cancelled event cannot sell tickets", domain.ErrConflict)
	}
	if len(e.TicketTypes) >= maxTicketTypes {
		return domain.TicketType{}, invalid("an event has at most %d ticket types", maxTicketTypes)
	}
	in, err = s.validateType(in)
	if err != nil {
		return domain.TicketType{}, err
	}
	if in.Seated {
		in.Total = 0
	}
	if in.SessionID, err = pickSession(e, in.SessionID); err != nil {
		return domain.TicketType{}, err
	}
	t, err := s.events.CreateTicketType(ctx, typeOf(eventID, 0, in))
	if err == nil {
		s.changed(eventID)
	}
	return t, err
}

func typeOf(eventID, id int64, in inbound.TicketTypeInput) domain.TicketType {
	return domain.TicketType{
		ID: id, EventID: eventID, Name: in.Name, Description: in.Description, Price: in.Price, Total: in.Total,
		Available: in.Total, MinPerOrder: in.MinPerOrder, MaxPerOrder: in.MaxPerOrder, MaxPerUser: in.MaxPerUser,
		SaleStartsAt: in.SaleStartsAt, SaleEndsAt: in.SaleEndsAt, Seated: in.Seated, Active: in.Active, SortOrder: in.SortOrder, SessionID: in.SessionID,
	}
}

func (s *EventService) UpdateTicketType(ctx context.Context, actor inbound.Principal, eventID, typeID int64, in inbound.TicketTypeInput) (domain.TicketType, error) {
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return domain.TicketType{}, err
	}
	if e.Status == domain.EventCancelled {
		return domain.TicketType{}, fmt.Errorf("%w: a cancelled event cannot sell tickets", domain.ErrConflict)
	}
	cur, err := s.events.TicketType(ctx, typeID)
	if err != nil {
		return domain.TicketType{}, err
	}
	if cur.EventID != eventID {
		return domain.TicketType{}, domain.ErrNotFound
	}
	if in.SessionID != 0 && in.SessionID != cur.SessionID {
		return domain.TicketType{}, invalid("a ticket type stays with its showtime; create another for the other one")
	}
	in.SessionID = cur.SessionID
	if in.Seated != cur.Seated {
		return domain.TicketType{}, invalid("a ticket type cannot change between seated and general admission")
	}
	if cur.Seated {
		in.Total = cur.Total
	}
	in, err = s.validateType(in)
	if err != nil {
		return domain.TicketType{}, err
	}
	t, err := s.events.UpdateTicketType(ctx, typeOf(eventID, typeID, in))
	if errors.Is(err, domain.ErrConflict) {
		return domain.TicketType{}, fmt.Errorf("%w: total cannot go below the tickets already held or sold", domain.ErrConflict)
	}
	if err == nil {
		s.changed(eventID)
		if s.waiters != nil && t.Total > cur.Total {
			s.waiters.TicketsBack(ctx, typeID, t.Total-cur.Total)
		}
	}
	return t, err
}

func (s *EventService) GenerateSeats(ctx context.Context, actor inbound.Principal, eventID, typeID int64, l inbound.SeatLayout) (int, error) {
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return 0, err
	}
	if e.Status == domain.EventCancelled {
		return 0, fmt.Errorf("%w: a cancelled event cannot sell tickets", domain.ErrConflict)
	}
	t, err := s.events.TicketType(ctx, typeID)
	if err != nil {
		return 0, err
	}
	if t.EventID != eventID {
		return 0, domain.ErrNotFound
	}
	if !t.Seated {
		return 0, invalid("this ticket type is general admission; create a seated one to have seats")
	}
	l.Section = trim(l.Section)
	first := strings.ToUpper(trim(l.FirstRow))
	if first == "" {
		first = "A"
	}
	for _, c := range first {
		if c < 'A' || c > 'Z' {
			return 0, invalid("first_row must be letters, like A or AA")
		}
	}
	switch {
	case l.Rows < 1 || l.Rows > maxSeatRows:
		return 0, invalid("rows must be between 1 and %d", maxSeatRows)
	case l.SeatsPerRow < 1 || l.SeatsPerRow > 200:
		return 0, invalid("seats_per_row must be between 1 and 200")
	case l.Rows*l.SeatsPerRow > maxSeatsPerCall:
		return 0, invalid("at most %d seats per call", maxSeatsPerCall)
	}
	start := domain.RowIndex(first)
	seats := make([]domain.Seat, 0, l.Rows*l.SeatsPerRow)
	for r := 0; r < l.Rows; r++ {
		for n := 1; n <= l.SeatsPerRow; n++ {
			x, y := l.OffsetX+float64(n-1)*seatPitchX, l.OffsetY+float64(r)*seatPitchY
			seats = append(seats, domain.Seat{TicketTypeID: typeID, Section: l.Section, Row: domain.RowLabel(start + r), Number: n, X: &x, Y: &y})
		}
	}
	added, err := s.events.AddSeats(ctx, typeID, seats)
	if err == nil {
		s.changed(eventID)
	}
	return added, err
}

func (s *EventService) SetSeatMap(ctx context.Context, actor inbound.Principal, eventID, typeID int64, l domain.SeatMapLayout, seats []domain.SeatPosition) (int, error) {
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return 0, err
	}
	if e.Status == domain.EventCancelled {
		return 0, fmt.Errorf("%w: a cancelled event cannot be edited", domain.ErrConflict)
	}
	t, err := s.events.TicketType(ctx, typeID)
	if err != nil {
		return 0, err
	}
	if t.EventID != eventID {
		return 0, domain.ErrNotFound
	}
	if !t.Seated {
		return 0, invalid("this ticket type is general admission and has no seat map")
	}
	if err := l.Validate(); err != nil {
		return 0, err
	}
	if len(seats) > maxSeatsPerCall {
		return 0, invalid("at most %d seats per call", maxSeatsPerCall)
	}
	for _, p := range seats {
		if p.Row == "" || p.Number < 1 || p.X < 0 || p.Y < 0 || p.X > l.Width || p.Y > l.Height {
			return 0, invalid("seat %s-%d must have a row, a number and a place inside the canvas", p.Row, p.Number)
		}
	}
	n, err := s.events.SetSeatMap(ctx, typeID, l, seats)
	if err == nil {
		s.changed(eventID)
	}
	return n, err
}

// ---- promotions

func (s *EventService) validatePromo(in inbound.PromotionInput) (inbound.PromotionInput, error) {
	in.Code = domain.NormalizeCode(in.Code)
	if in.MinTickets == 0 {
		in.MinTickets = 1
	}
	if n := len(in.Code); n < 3 || n > 32 || strings.Trim(in.Code, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-") != "" {
		return in, invalid("code must be 3-32 letters, digits, - or _")
	}
	switch {
	case in.Kind == domain.PromoPercent && (in.Value < 1 || in.Value > 100):
		return in, invalid("a percent discount must be between 1 and 100")
	case in.Kind == domain.PromoFixed && in.Value < 1:
		return in, invalid("a fixed discount must be positive")
	case in.Kind != domain.PromoPercent && in.Kind != domain.PromoFixed:
		return in, invalid("kind must be percent or fixed")
	case in.MaxUses < 0 || in.MinTickets < 1:
		return in, invalid("max_uses cannot be negative and min_tickets must be at least 1")
	case in.ValidFrom != nil && in.ValidTo != nil && !in.ValidTo.After(*in.ValidFrom):
		return in, invalid("the promo must end after it starts")
	}
	return in, nil
}

func (s *EventService) ListPromotions(ctx context.Context, actor inbound.Principal, eventID int64) ([]domain.Promotion, error) {
	if _, err := s.load(ctx, actor, eventID); err != nil {
		return nil, err
	}
	return s.events.Promotions(ctx, eventID)
}

func (s *EventService) CreatePromotion(ctx context.Context, actor inbound.Principal, eventID int64, in inbound.PromotionInput) (domain.Promotion, error) {
	if _, err := s.load(ctx, actor, eventID); err != nil {
		return domain.Promotion{}, err
	}
	in, err := s.validatePromo(in)
	if err != nil {
		return domain.Promotion{}, err
	}
	p, err := s.events.CreatePromotion(ctx, promoOf(eventID, 0, in))
	if errors.Is(err, domain.ErrConflict) {
		return domain.Promotion{}, fmt.Errorf("%w: this event already has a promo code %s", domain.ErrConflict, in.Code)
	}
	return p, err
}

func promoOf(eventID, id int64, in inbound.PromotionInput) domain.Promotion {
	return domain.Promotion{ID: id, EventID: eventID, Code: in.Code, Kind: in.Kind, Value: in.Value, MaxUses: in.MaxUses,
		MinTickets: in.MinTickets, ValidFrom: in.ValidFrom, ValidTo: in.ValidTo, Active: in.Active}
}

func (s *EventService) UpdatePromotion(ctx context.Context, actor inbound.Principal, eventID, promoID int64, in inbound.PromotionInput) (domain.Promotion, error) {
	if _, err := s.load(ctx, actor, eventID); err != nil {
		return domain.Promotion{}, err
	}
	in, err := s.validatePromo(in)
	if err != nil {
		return domain.Promotion{}, err
	}
	p, err := s.events.UpdatePromotion(ctx, promoOf(eventID, promoID, in))
	if errors.Is(err, domain.ErrConflict) {
		return domain.Promotion{}, fmt.Errorf("%w: max_uses cannot be below the times the code was already used, and the code must be unique in the event", domain.ErrConflict)
	}
	return p, err
}

// ---- reporting

func (s *EventService) Report(ctx context.Context, actor inbound.Principal, eventID int64) (domain.EventReport, error) {
	if _, err := s.load(ctx, actor, eventID); err != nil {
		return domain.EventReport{}, err
	}
	return s.events.Report(ctx, eventID)
}

func (s *EventService) Attendees(ctx context.Context, actor inbound.Principal, eventID, afterID int64, limit int) (domain.Page[domain.Attendee], error) {
	if _, err := s.load(ctx, actor, eventID); err != nil {
		return domain.Page[domain.Attendee]{}, err
	}
	limit = page(limit)
	rows, err := s.tickets.Attendees(ctx, eventID, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.Attendee]{}, err
	}
	return domain.PageOf(rows, limit, func(a domain.Attendee) int64 { return a.ID }), nil
}

func pickSession(e domain.Event, want int64) (int64, error) {
	if want != 0 {
		for _, s := range e.Sessions {
			if s.ID == want {
				if s.Status != domain.SessionScheduled {
					return 0, fmt.Errorf("%w: that showtime was cancelled", domain.ErrConflict)
				}
				return s.ID, nil
			}
		}
		return 0, invalid("session %d is not a showtime of this event", want)
	}
	for _, s := range e.Sessions {
		if s.Status == domain.SessionScheduled {
			return s.ID, nil
		}
	}
	return 0, fmt.Errorf("%w: every showtime of this event was cancelled", domain.ErrConflict)
}

// ---- sessions

const maxSessions = 60

func (s *EventService) validateSession(in inbound.SessionInput, creating bool) (inbound.SessionInput, error) {
	in.Label = trim(in.Label)
	switch {
	case len(in.Label) > 100:
		return in, invalid("label is too long")
	case !in.EndsAt.After(in.StartsAt):
		return in, invalid("a showtime must end after it starts")
	case creating && !in.StartsAt.After(s.now()):
		return in, invalid("a new showtime must start in the future")
	case in.EndsAt.Sub(in.StartsAt) > 14*24*time.Hour:
		return in, invalid("a showtime cannot last more than 14 days")
	}
	return in, nil
}

func sessionOf(e domain.Event, id int64) (domain.Session, bool) {
	for _, x := range e.Sessions {
		if x.ID == id {
			return x, true
		}
	}
	return domain.Session{}, false
}

func (s *EventService) CreateSession(ctx context.Context, actor inbound.Principal, eventID int64, in inbound.SessionInput) (domain.Session, error) {
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return domain.Session{}, err
	}
	if e.Status == domain.EventCancelled {
		return domain.Session{}, fmt.Errorf("%w: a cancelled event cannot get showtimes", domain.ErrConflict)
	}
	if len(e.Sessions) >= maxSessions {
		return domain.Session{}, invalid("an event has at most %d showtimes", maxSessions)
	}
	if in, err = s.validateSession(in, true); err != nil {
		return domain.Session{}, err
	}
	if in.CopyFrom != 0 {
		if _, ok := sessionOf(e, in.CopyFrom); !ok {
			return domain.Session{}, invalid("copy_from must be a showtime of this event")
		}
	}
	out, err := s.events.CreateSession(ctx, domain.Session{EventID: eventID, StartsAt: in.StartsAt, EndsAt: in.EndsAt, Label: in.Label}, in.CopyFrom)
	if err == nil {
		s.changed(eventID)
	}
	return out, err
}

func (s *EventService) UpdateSession(ctx context.Context, actor inbound.Principal, eventID, sessionID int64, in inbound.SessionInput) (domain.Session, error) {
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return domain.Session{}, err
	}
	cur, ok := sessionOf(e, sessionID)
	if !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	if cur.Status != domain.SessionScheduled {
		return domain.Session{}, fmt.Errorf("%w: a cancelled showtime cannot be edited", domain.ErrConflict)
	}
	if in, err = s.validateSession(in, false); err != nil {
		return domain.Session{}, err
	}
	if e.Status == domain.EventPublished && !in.EndsAt.After(s.now()) {
		return domain.Session{}, invalid("a showtime cannot be moved into the past")
	}
	out, err := s.events.UpdateSession(ctx, domain.Session{ID: sessionID, EventID: eventID, StartsAt: in.StartsAt, EndsAt: in.EndsAt, Label: in.Label})
	if err == nil {
		s.changed(eventID)
		s.tellMoved(ctx, e, cur, in.StartsAt)
	}
	return out, err
}

func (s *EventService) DeleteSession(ctx context.Context, actor inbound.Principal, eventID, sessionID int64) error {
	if _, err := s.load(ctx, actor, eventID); err != nil {
		return err
	}
	err := s.events.DeleteSession(ctx, eventID, sessionID)
	if err == nil {
		s.changed(eventID)
	}
	return err
}

func (s *EventService) CancelSession(ctx context.Context, actor inbound.Principal, eventID, sessionID int64, reason string) (domain.Session, error) {
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return domain.Session{}, err
	}
	if _, ok := sessionOf(e, sessionID); !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	if reason = trim(reason); reason == "" {
		return domain.Session{}, invalid("tell ticket holders why the showtime was cancelled")
	}
	out, err := s.events.CancelSession(ctx, eventID, sessionID, reason)
	if errors.Is(err, domain.ErrConflict) {
		return domain.Session{}, fmt.Errorf("%w: the showtime is already cancelled", domain.ErrConflict)
	}
	if err == nil {
		s.changed(eventID)
	}
	return out, err
}
