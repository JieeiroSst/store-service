package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

const (
	maxPageSize     = 100
	defaultPageSize = 20
	maxRangeDays    = 366
)

type HomestayService struct {
	repo    outbound.HomestayRepository
	wallets outbound.WalletGateway
	tiers   *HostTiers
}

var _ inbound.HomestayUseCase = (*HomestayService)(nil)

func NewHomestayService(repo outbound.HomestayRepository, wallets outbound.WalletGateway, tiers *HostTiers) *HomestayService {
	return &HomestayService{repo: repo, wallets: wallets, tiers: tiers}
}

func (s *HomestayService) checkWallet(ctx context.Context, id string, owner int64, enforceOwner bool) error {
	if id == "" {
		return nil
	}
	if s.wallets == nil {
		return fmt.Errorf("%w: wallet payment is not enabled, so no wallet can be set", domain.ErrInvalid)
	}
	w, err := s.wallets.Get(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("%w: wallet %s does not exist", domain.ErrInvalid, id)
	}
	if err != nil {
		return err
	}
	if w.Status != "ACTIVE" {
		return fmt.Errorf("%w: wallet %s is %s and cannot receive payments", domain.ErrInvalid, id, strings.ToLower(w.Status))
	}
	if enforceOwner && w.UserID != owner {
		return fmt.Errorf("%w: wallet %s is not yours", domain.ErrForbidden, id)
	}
	return nil
}

func canManage(a inbound.Principal, h domain.Homestay) bool {
	return a.IsAdmin() || (a.UserID != 0 && h.HostID == a.UserID)
}

func (s *HomestayService) manageable(ctx context.Context, actor inbound.Principal, id int64) (domain.Homestay, error) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Homestay{}, err
	}
	if !canManage(actor, h) {
		if h.Status != domain.HomestayActive {
			return domain.Homestay{}, domain.ErrNotFound
		}
		return domain.Homestay{}, domain.ErrForbidden
	}
	return h, nil
}

func requireAdmin(p inbound.Principal) error {
	if !p.IsAdmin() {
		return domain.ErrForbidden
	}
	return nil
}

func (s *HomestayService) View(ctx context.Context, viewer *inbound.Principal, id int64) (domain.Homestay, error) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Homestay{}, err
	}
	if h.Status != domain.HomestayActive && (viewer == nil || !canManage(*viewer, h)) {
		return domain.Homestay{}, domain.ErrNotFound
	}
	return h, nil
}

func (s *HomestayService) Mine(ctx context.Context, actor inbound.Principal, limit, offset int) ([]domain.Homestay, error) {
	limit, offset = page(limit, offset)
	return s.repo.List(ctx, domain.HomestayFilter{HostID: actor.UserID, IncludeInactive: true, Sort: "newest", Limit: limit, Offset: offset})
}

func (s *HomestayService) ForReview(ctx context.Context, actor inbound.Principal, status domain.HomestayStatus, limit, offset int) ([]domain.Homestay, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	if status == 0 {
		status = domain.HomestayPending
	}
	if status.Name() == "" {
		return nil, fmt.Errorf("%w: unknown status", domain.ErrInvalid)
	}
	limit, offset = page(limit, offset)
	return s.repo.List(ctx, domain.HomestayFilter{Status: status, IncludeInactive: true, Sort: "oldest", Limit: limit, Offset: offset})
}

func (s *HomestayService) Approve(ctx context.Context, actor inbound.Principal, id int64) (domain.Homestay, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Homestay{}, err
	}
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Homestay{}, err
	}
	if h.Status != domain.HomestayPending && h.Status != domain.HomestayRejected {
		return domain.Homestay{}, fmt.Errorf("%w: homestay is %s, not waiting for approval", domain.ErrConflict, h.Status.Name())
	}
	if err := s.repo.SetReview(ctx, id, domain.HomestayActive, "", actor.UserID); err != nil {
		return domain.Homestay{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *HomestayService) Reject(ctx context.Context, actor inbound.Principal, id int64, reason string) (domain.Homestay, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Homestay{}, err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return domain.Homestay{}, fmt.Errorf("%w: give the owner a reason", domain.ErrInvalid)
	}
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Homestay{}, err
	}
	if h.Status != domain.HomestayPending {
		return domain.Homestay{}, fmt.Errorf("%w: homestay is %s, not waiting for approval", domain.ErrConflict, h.Status.Name())
	}
	if err := s.repo.SetReview(ctx, id, domain.HomestayRejected, reason, actor.UserID); err != nil {
		return domain.Homestay{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *HomestayService) List(ctx context.Context, f domain.HomestayFilter) (domain.HomestayPage, error) {
	limit, _ := page(f.Limit, 0)
	f.Offset, f.IncludeInactive = 0, false
	if f.Type != 0 && !domain.PropertyType(f.Type).Valid() {
		return domain.HomestayPage{}, fmt.Errorf("%w: type must be 1 (homestay), 2 (hotel) or 3 (rental house)", domain.ErrInvalid)
	}
	if f.Model != "" && !f.Model.Valid() {
		return domain.HomestayPage{}, fmt.Errorf("%w: unknown rental model %q", domain.ErrInvalid, f.Model)
	}
	for _, p := range []string{f.MinPrice, f.MaxPrice} {
		if p == "" {
			continue
		}
		if v, err := strconv.ParseFloat(p, 64); err != nil || v < 0 {
			return domain.HomestayPage{}, fmt.Errorf("%w: prices must be non-negative numbers", domain.ErrInvalid)
		}
	}
	if (!f.CheckIn.IsZero() || !f.CheckOut.IsZero()) && (f.CheckIn.IsZero() || !f.CheckOut.After(f.CheckIn)) {
		return domain.HomestayPage{}, fmt.Errorf("%w: checkin and checkout must both be given, checkout after checkin", domain.ErrInvalid)
	}
	sortBy := f.Sort
	switch sortBy {
	case "":
		sortBy = "recommended"
	case "recommended", "rating", "price_asc", "price_desc", "newest":
	default:
		return domain.HomestayPage{}, fmt.Errorf("%w: sort must be recommended, rating, price_asc, price_desc or newest", domain.ErrInvalid)
	}
	if f.After != nil && f.After.Sort != sortBy {
		return domain.HomestayPage{}, fmt.Errorf("%w: cursor belongs to a different sort", domain.ErrInvalid)
	}
	cursor := f.After

	if sortBy == "recommended" {
		f.Sort, f.Limit, f.After = "rating", rankCandidates, nil
		list, err := s.repo.List(ctx, f)
		if err != nil {
			return domain.HomestayPage{}, err
		}
		hosts := make([]int64, 0, len(list))
		for _, h := range list {
			hosts = append(hosts, h.HostID)
		}
		ranked := rank(list, s.tiers.Tiers(ctx, hosts))
		if cursor != nil {
			ranked = after(ranked, *cursor)
		}
		pg := domain.HomestayPage{Items: []domain.Homestay{}}
		for i, r := range ranked {
			if i == limit {
				last := ranked[limit-1]
				pg.Next = &domain.HomestayCursor{Sort: sortBy, ID: last.h.ID, Count: last.h.ReviewCount, Score: last.score}
				break
			}
			pg.Items = append(pg.Items, r.h)
		}
		return pg, nil
	}

	f.Sort, f.Limit = sortBy, limit+1
	list, err := s.repo.List(ctx, f)
	if err != nil {
		return domain.HomestayPage{}, err
	}
	pg := domain.HomestayPage{Items: list}
	if len(list) > limit {
		pg.Items = list[:limit]
		pg.Next = cursorAfter(sortBy, f.Model, list[limit-1])
	}
	return pg, nil
}

func cursorAfter(sortBy string, model domain.RentalModel, h domain.Homestay) *domain.HomestayCursor {
	c := &domain.HomestayCursor{Sort: sortBy, ID: h.ID, Rating: h.Rating, Count: h.ReviewCount}
	if sortBy == "price_asc" || sortBy == "price_desc" {
		if model == "" {
			model = domain.ModelDay
		}
		for _, r := range h.Rates {
			if r.Model == model && r.Active {
				c.Price = r.Price
			}
		}
	}
	return c
}

func (s *HomestayService) Rates(ctx context.Context, id int64) ([]domain.Rate, error) {
	if _, err := s.repo.Get(ctx, id); err != nil {
		return nil, err
	}
	rates, err := s.repo.Rates(ctx, id)
	if err != nil {
		return nil, err
	}
	active := rates[:0]
	for _, r := range rates {
		if r.Active {
			active = append(active, r)
		}
	}
	return active, nil
}

func (s *HomestayService) SetRate(ctx context.Context, actor inbound.Principal, r domain.Rate) error {
	if _, err := s.manageable(ctx, actor, r.HomestayID); err != nil {
		return err
	}
	if !r.Model.Valid() {
		return fmt.Errorf("%w: model must be day, week, month or year", domain.ErrInvalid)
	}
	if p, err := strconv.ParseFloat(r.Price, 64); err != nil || p < 0 {
		return fmt.Errorf("%w: price must be a non-negative number", domain.ErrInvalid)
	}
	r.Currency = strings.ToUpper(strings.TrimSpace(r.Currency))
	if len(r.Currency) != 3 {
		return fmt.Errorf("%w: currency must be a 3-letter code", domain.ErrInvalid)
	}
	if r.MinPeriods == 0 {
		r.MinPeriods = 1
	}
	if r.MinPeriods < 1 || (r.Model == domain.ModelDay && r.MinPeriods != 1) || (r.Model.IsLease() && r.MinPeriods > r.Model.MaxPeriods()) {
		return fmt.Errorf("%w: min_periods out of range for %s", domain.ErrInvalid, r.Model)
	}
	return s.repo.SetRate(ctx, r, actor.UserID)
}

func (s *HomestayService) DeleteRate(ctx context.Context, actor inbound.Principal, id int64, model domain.RentalModel) error {
	if _, err := s.manageable(ctx, actor, id); err != nil {
		return err
	}
	if !model.Valid() {
		return fmt.Errorf("%w: model must be day, week, month or year", domain.ErrInvalid)
	}
	return s.repo.DeleteRate(ctx, id, model)
}

func (s *HomestayService) Availability(ctx context.Context, id int64, from, to time.Time) ([]domain.Slot, error) {
	if err := validRange(from, to); err != nil {
		return nil, err
	}
	if _, err := s.repo.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.Availability(ctx, id, from, to)
}

func (s *HomestayService) ListAmenities(ctx context.Context) ([]domain.Amenity, error) {
	return s.repo.ListAmenities(ctx)
}

func (s *HomestayService) Create(ctx context.Context, actor inbound.Principal, in inbound.HomestayInput) (domain.Homestay, error) {
	h, err := toHomestay(in)
	if err != nil {
		return domain.Homestay{}, err
	}
	switch {
	case !actor.IsAdmin():
		if in.Status != 0 {
			return domain.Homestay{}, fmt.Errorf("%w: status is set by approval, leave it out", domain.ErrInvalid)
		}
		h.Status = domain.HomestayPending
	case in.Status == 0:
		h.Status = domain.HomestayActive
	case in.Status != domain.HomestayActive && in.Status != domain.HomestayInactive:
		return domain.Homestay{}, fmt.Errorf("%w: status must be active or inactive", domain.ErrInvalid)
	default:
		h.Status = in.Status
	}
	if in.WalletID != nil {
		h.WalletID = strings.TrimSpace(*in.WalletID)
	}
	if err := s.checkWallet(ctx, h.WalletID, actor.UserID, !actor.IsAdmin()); err != nil {
		return domain.Homestay{}, err
	}
	return s.repo.Create(ctx, h, actor.UserID)
}

func (s *HomestayService) Update(ctx context.Context, actor inbound.Principal, id int64, in inbound.HomestayInput) (domain.Homestay, error) {
	current, err := s.manageable(ctx, actor, id)
	if err != nil {
		return domain.Homestay{}, err
	}
	h, err := toHomestay(in)
	if err != nil {
		return domain.Homestay{}, err
	}
	h.ID = id

	want := in.Status
	if want != 0 && want != domain.HomestayActive && want != domain.HomestayInactive {
		return domain.Homestay{}, fmt.Errorf("%w: status can only be set to active or inactive; approval decides the rest", domain.ErrInvalid)
	}
	switch current.Status {
	case domain.HomestayPending, domain.HomestayRejected:
		if want != 0 {
			return domain.Homestay{}, fmt.Errorf("%w: homestay is %s; it cannot be switched on before approval", domain.ErrConflict, current.Status.Name())
		}
		h.Status = domain.HomestayPending
		if current.Status == domain.HomestayRejected && current.HostID != actor.UserID {
			h.Status = domain.HomestayRejected
		}
	default:
		h.Status = current.Status
		if want != 0 {
			h.Status = want
		}
	}

	if in.WalletID == nil {
		h.WalletID = current.WalletID
	} else {
		h.WalletID = strings.TrimSpace(*in.WalletID)
		if h.WalletID != current.WalletID {
			if err := s.checkWallet(ctx, h.WalletID, current.HostID, !actor.IsAdmin()); err != nil {
				return domain.Homestay{}, err
			}
		}
	}
	return s.repo.Update(ctx, h, actor.UserID)
}

func (s *HomestayService) Deactivate(ctx context.Context, actor inbound.Principal, id int64) error {
	h, err := s.manageable(ctx, actor, id)
	if err != nil {
		return err
	}
	if h.Status != domain.HomestayActive {
		return fmt.Errorf("%w: homestay is %s; only a published one can be paused", domain.ErrConflict, h.Status.Name())
	}
	return s.repo.SetStatus(ctx, id, domain.HomestayInactive, actor.UserID)
}

func (s *HomestayService) SetAvailability(ctx context.Context, actor inbound.Principal, c inbound.SetAvailabilityCommand) error {
	if _, err := s.manageable(ctx, actor, c.HomestayID); err != nil {
		return err
	}
	if err := validRange(c.From, c.To); err != nil {
		return err
	}
	if c.Status != domain.SlotAvailable && c.Status != domain.SlotBlocked {
		return fmt.Errorf("%w: status must be available or blocked", domain.ErrInvalid)
	}
	if c.Status == domain.SlotAvailable && c.Price == "" {
		rates, err := s.repo.Rates(ctx, c.HomestayID)
		if err != nil {
			return err
		}
		for _, r := range rates {
			if r.Model == domain.ModelDay && r.Active {
				c.Price = r.Price
			}
		}
		if c.Price == "" {
			return fmt.Errorf("%w: give a price or set a day rate first", domain.ErrInvalid)
		}
	}
	if c.Status == domain.SlotAvailable {
		p, err := strconv.ParseFloat(c.Price, 64)
		if err != nil || p < 0 {
			return fmt.Errorf("%w: price must be a non-negative number", domain.ErrInvalid)
		}
	}
	return s.repo.SetAvailability(ctx, c.HomestayID, c.From, c.To, c.Price, c.Status)
}

func (s *HomestayService) CreateAmenity(ctx context.Context, actor inbound.Principal, name, icon string) (domain.Amenity, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Amenity{}, err
	}
	name, icon = strings.TrimSpace(name), strings.TrimSpace(icon)
	if name == "" || icon == "" {
		return domain.Amenity{}, fmt.Errorf("%w: name and icon are required", domain.ErrInvalid)
	}
	return s.repo.CreateAmenity(ctx, domain.Amenity{Name: name, Icon: icon})
}

var phonePattern = regexp.MustCompile(`^[0-9+()\-\s.]{8,20}$`)

func toHomestay(in inbound.HomestayInput) (domain.Homestay, error) {
	name, address, phone := strings.TrimSpace(in.Name), strings.TrimSpace(in.Address), strings.TrimSpace(in.PhoneNumber)
	if name == "" {
		return domain.Homestay{}, fmt.Errorf("%w: name is required", domain.ErrInvalid)
	}
	if !domain.PropertyType(in.Type).Valid() {
		return domain.Homestay{}, fmt.Errorf("%w: type must be 1 (homestay), 2 (hotel) or 3 (rental house)", domain.ErrInvalid)
	}
	if address == "" {
		return domain.Homestay{}, fmt.Errorf("%w: address is required", domain.ErrInvalid)
	}
	if !phonePattern.MatchString(phone) {
		return domain.Homestay{}, fmt.Errorf("%w: a contact phone number is required (8-20 digits, + - ( ) allowed)", domain.ErrInvalid)
	}
	if in.Guests <= 0 {
		return domain.Homestay{}, fmt.Errorf("%w: guests must be positive", domain.ErrInvalid)
	}
	if in.Bedrooms < 0 || in.Bathrooms < 0 {
		return domain.Homestay{}, fmt.Errorf("%w: bedrooms and bathrooms cannot be negative", domain.ErrInvalid)
	}
	return domain.Homestay{
		Name: name, Description: in.Description, Type: in.Type,
		PhoneNumber: phone, Address: address,
		WardID: in.WardID, DistrictID: in.DistrictID, ProvinceID: in.ProvinceID,
		Images: in.Images, Guests: in.Guests, Bedrooms: in.Bedrooms, Bathrooms: in.Bathrooms,
		AmenityIDs: in.AmenityIDs,
	}, nil
}

func validRange(from, to time.Time) error {
	if !to.After(from) {
		return fmt.Errorf("%w: end date must be after start date", domain.ErrInvalid)
	}
	if to.Sub(from) > maxRangeDays*24*time.Hour {
		return fmt.Errorf("%w: range cannot exceed %d days", domain.ErrInvalid, maxRangeDays)
	}
	return nil
}

func page(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
