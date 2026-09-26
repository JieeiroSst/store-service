package service

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type VenueService struct {
	venues outbound.VenueRepository
}

var _ inbound.VenueUseCase = (*VenueService)(nil)

func NewVenueService(v outbound.VenueRepository) *VenueService { return &VenueService{venues: v} }

func (s *VenueService) Templates() []domain.VenueTemplate { return domain.VenueTemplates() }

func countSections(m domain.VenueMap, seats []domain.GeneratedSeat) []inbound.SectionCount {
	idx := map[string]int{}
	var out []inbound.SectionCount
	for _, sec := range m.Sections {
		idx[sec.Key] = len(out)
		name := sec.Name
		if name == "" {
			name = sec.Key
		}
		out = append(out, inbound.SectionCount{Key: sec.Key, Name: name})
	}
	for _, s := range seats {
		c := &out[idx[s.Section]]
		c.Seats++
		if s.Accessible {
			c.Accessible++
		}
	}
	return out
}

func blueprint(in inbound.VenueInput) (domain.VenueMap, []domain.GeneratedSeat, error) {
	var m domain.VenueMap
	switch {
	case in.Template != "" && in.Map != nil:
		return m, nil, invalid("give a template or a map, not both")
	case in.Template != "":
		var err error
		if m, err = domain.BuildVenueTemplate(in.Template, in.Params); err != nil {
			return m, nil, err
		}
	case in.Map != nil:
		m = *in.Map
	default:
		return m, nil, invalid("give a template or a map")
	}
	seats, err := m.Generate()
	return m, seats, err
}

func (s *VenueService) Preview(_ context.Context, in inbound.VenueInput) (inbound.VenuePreview, error) {
	m, seats, err := blueprint(in)
	if err != nil {
		return inbound.VenuePreview{}, err
	}
	return inbound.VenuePreview{Map: m, Seats: seats, Sections: countSections(m, seats), Total: len(seats), View: m.View(seats, nil)}, nil
}

func (s *VenueService) build(in inbound.VenueInput, v domain.Venue) (domain.Venue, error) {
	in.Name, in.City, in.Address, in.Description = trim(in.Name), trim(in.City), trim(in.Address), trim(in.Description)
	if len(in.Name) < 2 || len(in.Name) > 120 || len(in.City) > 120 || len(in.Address) > 300 || len(in.Description) > 5000 {
		return v, invalid("name must be 2-120 characters, and city, address and description of a sensible length")
	}
	m, seats, err := blueprint(in)
	if err != nil {
		return v, err
	}
	v.Name, v.City, v.Address, v.Description = in.Name, in.City, in.Address, in.Description
	v.Map, v.Seats, v.Sections = m, len(seats), len(m.Sections)
	return v, nil
}

func (s *VenueService) Create(ctx context.Context, actor inbound.Principal, in inbound.VenueInput) (domain.Venue, error) {
	v, err := s.build(in, domain.Venue{OwnerID: actor.UserID})
	if err != nil {
		return domain.Venue{}, err
	}
	return s.venues.Create(ctx, v)
}

func (s *VenueService) readable(ctx context.Context, actor inbound.Principal, id int64) (domain.Venue, error) {
	v, err := s.venues.Get(ctx, id)
	if err != nil {
		return v, err
	}
	if !v.Shared && v.OwnerID != actor.UserID && !actor.IsAdmin() {
		return domain.Venue{}, domain.ErrNotFound
	}
	return v, nil
}

func (s *VenueService) writable(ctx context.Context, actor inbound.Principal, id int64) (domain.Venue, error) {
	v, err := s.readable(ctx, actor, id)
	if err != nil {
		return v, err
	}
	if v.OwnerID != actor.UserID && !actor.IsAdmin() {
		return domain.Venue{}, domain.ErrForbidden
	}
	return v, nil
}

func (s *VenueService) Get(ctx context.Context, actor inbound.Principal, id int64) (domain.Venue, error) {
	return s.readable(ctx, actor, id)
}

func (s *VenueService) Update(ctx context.Context, actor inbound.Principal, id int64, in inbound.VenueInput) (domain.Venue, error) {
	cur, err := s.writable(ctx, actor, id)
	if err != nil {
		return domain.Venue{}, err
	}
	v, err := s.build(in, cur)
	if err != nil {
		return domain.Venue{}, err
	}
	return s.venues.Update(ctx, v)
}

func (s *VenueService) Delete(ctx context.Context, actor inbound.Principal, id int64) error {
	if _, err := s.writable(ctx, actor, id); err != nil {
		return err
	}
	return s.venues.Delete(ctx, id)
}

func (s *VenueService) List(ctx context.Context, actor inbound.Principal, shared bool, afterID int64, limit int) (domain.Page[domain.Venue], error) {
	limit = page(limit)
	f := outbound.VenueFilter{AfterID: afterID, Limit: limit + 1}
	if shared {
		f.Shared = true
	} else {
		f.OwnerID = actor.UserID
	}
	rows, err := s.venues.List(ctx, f)
	if err != nil {
		return domain.Page[domain.Venue]{}, err
	}
	return domain.PageOf(rows, limit, func(v domain.Venue) int64 { return v.ID }), nil
}

func (s *VenueService) Copy(ctx context.Context, actor inbound.Principal, id int64, name string) (domain.Venue, error) {
	v, err := s.readable(ctx, actor, id)
	if err != nil {
		return domain.Venue{}, err
	}
	if name = trim(name); name == "" {
		name = "Copy of " + v.Name
	}
	return s.venues.Create(ctx, domain.Venue{OwnerID: actor.UserID, Name: name, City: v.City, Address: v.Address, Description: v.Description,
		Map: v.Map, Seats: v.Seats, Sections: v.Sections})
}

func (s *VenueService) SetShared(ctx context.Context, actor inbound.Principal, id int64, shared bool) (domain.Venue, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Venue{}, err
	}
	if err := s.venues.SetShared(ctx, id, shared); err != nil {
		return domain.Venue{}, err
	}
	return s.venues.Get(ctx, id)
}

// ---- using a venue for an event

func (s *EventService) WithVenues(v outbound.VenueRepository) *EventService {
	s.venues = v
	return s
}

func (s *EventService) ApplyVenueMap(ctx context.Context, actor inbound.Principal, eventID int64, in inbound.ApplyVenueInput) (inbound.ApplyVenueResult, error) {
	if s.venues == nil {
		return inbound.ApplyVenueResult{}, invalid("venues are not enabled")
	}
	e, err := s.load(ctx, actor, eventID)
	if err != nil {
		return inbound.ApplyVenueResult{}, err
	}
	if e.Status == domain.EventCancelled {
		return inbound.ApplyVenueResult{}, fmt.Errorf("%w: a cancelled event cannot be edited", domain.ErrConflict)
	}
	v, err := s.venues.Get(ctx, in.VenueID)
	if err != nil {
		return inbound.ApplyVenueResult{}, err
	}
	if !v.Shared && v.OwnerID != actor.UserID && !actor.IsAdmin() {
		return inbound.ApplyVenueResult{}, domain.ErrNotFound
	}
	if len(in.Assignments) == 0 {
		return inbound.ApplyVenueResult{}, invalid("say which ticket type sells each section: assignments is empty")
	}
	sections := map[string]bool{}
	for _, sec := range v.Map.Sections {
		sections[sec.Key] = true
	}
	seated := map[int64]bool{}
	for _, t := range e.TicketTypes {
		if t.Seated {
			seated[t.ID] = true
		}
	}
	typeOf := map[string]int64{}
	for _, a := range in.Assignments {
		switch {
		case !sections[a.Section]:
			return inbound.ApplyVenueResult{}, invalid("the venue has no section %q", a.Section)
		case typeOf[a.Section] != 0:
			return inbound.ApplyVenueResult{}, invalid("section %q is assigned twice", a.Section)
		case !seated[a.TicketTypeID]:
			return inbound.ApplyVenueResult{}, invalid("ticket type %d is not a seated ticket type of this event", a.TicketTypeID)
		}
		typeOf[a.Section] = a.TicketTypeID
	}

	all, err := v.Map.Generate()
	if err != nil {
		return inbound.ApplyVenueResult{}, err
	}
	groups := map[int64][]domain.GeneratedSeat{}
	var chosen []domain.GeneratedSeat
	for _, seat := range all {
		if t, ok := typeOf[seat.Section]; ok {
			groups[t] = append(groups[t], seat)
			chosen = append(chosen, seat)
		}
	}
	sub := v.Map
	sub.Sections = nil
	for _, sec := range v.Map.Sections {
		if typeOf[sec.Key] != 0 {
			sub.Sections = append(sub.Sections, sec)
		}
	}
	var ordered []domain.SeatGroup
	for _, a := range in.Assignments {
		if g, ok := groups[a.TicketTypeID]; ok {
			ordered = append(ordered, domain.SeatGroup{TicketTypeID: a.TicketTypeID, Seats: g})
			delete(groups, a.TicketTypeID)
		}
	}
	n, err := s.events.ApplyVenueMap(ctx, eventID, v.ID, sub.View(chosen, typeOf), ordered, in.Replace)
	if err != nil {
		return inbound.ApplyVenueResult{}, err
	}
	s.changed(eventID)
	return inbound.ApplyVenueResult{Seats: n, Sections: countSections(sub, chosen)}, nil
}
