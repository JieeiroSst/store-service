package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const (
	defaultCheckIn         = "14:00"
	defaultCheckOut        = "12:00"
	defaultFreeCancelHours = 48
)

var (
	phonePattern = regexp.MustCompile(`^[0-9+()\-\s.]{8,20}$`)
	timePattern  = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
)

type HotelService struct {
	repo    outbound.HotelRepository
	wallets outbound.WalletGateway // nil: wallet payment is not enabled
	tiers   *ManagerTiers          // nil: rank by ratings only
	notify  *Notifier              // nil: nobody is notified
}

var _ inbound.HotelUseCase = (*HotelService)(nil)

func NewHotelService(repo outbound.HotelRepository, wallets outbound.WalletGateway, tiers *ManagerTiers) *HotelService {
	return &HotelService{repo: repo, wallets: wallets, tiers: tiers}
}

func (s *HotelService) WithNotifier(n *Notifier) *HotelService {
	s.notify = n
	return s
}

func (s *HotelService) checkWallet(ctx context.Context, id string, owner int64, enforceOwner bool) error {
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

func (s *HotelService) manageable(ctx context.Context, actor inbound.Principal, id int64) (domain.Hotel, error) {
	return manageableHotel(ctx, s.repo, actor, id)
}

func manageableHotel(ctx context.Context, repo outbound.HotelRepository, actor inbound.Principal, id int64) (domain.Hotel, error) {
	h, err := repo.Get(ctx, id)
	if err != nil {
		return domain.Hotel{}, err
	}
	if !canManage(actor, h) {
		if h.Status != domain.HotelActive {
			return domain.Hotel{}, domain.ErrNotFound
		}
		return domain.Hotel{}, domain.ErrForbidden
	}
	return h, nil
}

func (s *HotelService) View(ctx context.Context, viewer *inbound.Principal, id int64) (domain.Hotel, error) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Hotel{}, err
	}
	if h.Status != domain.HotelActive && (viewer == nil || !canManage(*viewer, h)) {
		return domain.Hotel{}, domain.ErrNotFound
	}
	if viewer == nil || !canManage(*viewer, h) {
		h.RoomTypes = activeRoomTypes(h.RoomTypes) // switched-off room types are the manager's business
	}
	return h, nil
}

func activeRoomTypes(in []domain.RoomType) []domain.RoomType {
	out := in[:0:0]
	for _, t := range in {
		if t.Active {
			out = append(out, t)
		}
	}
	return out
}

func (s *HotelService) Search(ctx context.Context, f domain.HotelFilter) (domain.HotelPage, error) {
	limit, _ := page(f.Limit, 0)
	f.Offset, f.IncludeInactive, f.OwnerID, f.Status = 0, false, 0, 0 // the public sees published hotels only
	f.City, f.Query = trim(f.City), trim(f.Query)
	if (!f.CheckIn.IsZero() || !f.CheckOut.IsZero()) && (f.CheckIn.IsZero() || !f.CheckOut.After(f.CheckIn)) {
		return domain.HotelPage{}, fmt.Errorf("%w: checkin and checkout must both be given, checkout after checkin", domain.ErrInvalid)
	}
	if f.Rooms < 0 || f.Guests < 0 {
		return domain.HotelPage{}, fmt.Errorf("%w: rooms and guests cannot be negative", domain.ErrInvalid)
	}
	if len(f.Amenities) > 20 {
		return domain.HotelPage{}, fmt.Errorf("%w: at most 20 amenities can be asked for", domain.ErrInvalid)
	}
	for _, r := range []string{f.MinRate, f.MaxRate} {
		if r == "" {
			continue
		}
		if v, err := strconv.ParseFloat(r, 64); err != nil || v < 0 {
			return domain.HotelPage{}, fmt.Errorf("%w: rates must be non-negative numbers", domain.ErrInvalid)
		}
	}
	sortBy := f.Sort
	switch sortBy {
	case "":
		sortBy = "recommended"
	case "recommended", "rating", "newest":
	default:
		return domain.HotelPage{}, fmt.Errorf("%w: sort must be recommended, rating or newest", domain.ErrInvalid)
	}
	if f.After != nil && f.After.Sort != sortBy {
		return domain.HotelPage{}, fmt.Errorf("%w: cursor belongs to a different sort", domain.ErrInvalid)
	}
	cursor := f.After

	if sortBy == "recommended" {
		f.Sort, f.Limit, f.After = "rating", rankCandidates, nil
		list, err := s.repo.List(ctx, f)
		if err != nil {
			return domain.HotelPage{}, err
		}
		owners := make([]int64, 0, len(list))
		for _, h := range list {
			owners = append(owners, h.OwnerID)
		}
		ranked := rank(list, s.tiers.Tiers(ctx, owners))
		if cursor != nil {
			ranked = after(ranked, *cursor)
		}
		pg := domain.HotelPage{Items: []domain.Hotel{}}
		for i, r := range ranked {
			if i == limit {
				last := ranked[limit-1]
				pg.Next = &domain.HotelCursor{Sort: sortBy, ID: last.h.ID, Count: last.h.ReviewCount, Score: last.score}
				break
			}
			pg.Items = append(pg.Items, r.h)
		}
		return pg, nil
	}

	f.Sort, f.Limit = sortBy, limit+1
	list, err := s.repo.List(ctx, f)
	if err != nil {
		return domain.HotelPage{}, err
	}
	pg := domain.HotelPage{Items: list}
	if len(list) > limit {
		pg.Items = list[:limit]
		last := list[limit-1]
		pg.Next = &domain.HotelCursor{Sort: sortBy, ID: last.ID, Rating: last.Rating, Count: last.ReviewCount}
	}
	return pg, nil
}

func (s *HotelService) listByCursor(ctx context.Context, f domain.HotelFilter, sortBy string, after *domain.HotelCursor, limit int) (domain.HotelPage, error) {
	limit, _ = page(limit, 0)
	if after != nil && after.Sort != sortBy {
		return domain.HotelPage{}, fmt.Errorf("%w: cursor belongs to a different list", domain.ErrInvalid)
	}
	f.Sort, f.After, f.Limit = sortBy, after, limit+1
	list, err := s.repo.List(ctx, f)
	if err != nil {
		return domain.HotelPage{}, err
	}
	pg := domain.HotelPage{Items: list}
	if len(list) > limit {
		pg.Items = list[:limit]
		pg.Next = &domain.HotelCursor{Sort: sortBy, ID: list[limit-1].ID}
	}
	return pg, nil
}

func (s *HotelService) Mine(ctx context.Context, actor inbound.Principal, after *domain.HotelCursor, limit int) (domain.HotelPage, error) {
	return s.listByCursor(ctx, domain.HotelFilter{OwnerID: actor.UserID, IncludeInactive: true}, "newest", after, limit)
}

func (s *HotelService) ForReview(ctx context.Context, actor inbound.Principal, status domain.HotelStatus, after *domain.HotelCursor, limit int) (domain.HotelPage, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.HotelPage{}, err
	}
	if status == 0 {
		status = domain.HotelPending
	}
	if status.Name() == "" {
		return domain.HotelPage{}, fmt.Errorf("%w: unknown status", domain.ErrInvalid)
	}
	return s.listByCursor(ctx, domain.HotelFilter{Status: status, IncludeInactive: true}, "oldest", after, limit)
}

func (s *HotelService) Approve(ctx context.Context, actor inbound.Principal, id int64) (domain.Hotel, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Hotel{}, err
	}
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Hotel{}, err
	}
	if h.Status != domain.HotelPending && h.Status != domain.HotelRejected {
		return domain.Hotel{}, fmt.Errorf("%w: hotel is %s, not waiting for verification", domain.ErrConflict, h.Status.Name())
	}
	if err := s.repo.SetReview(ctx, id, domain.HotelActive, ""); err != nil {
		return domain.Hotel{}, err
	}
	s.notify.Send(ctx, h.OwnerID, "hotel.approved", h.Name+" is live",
		"Your hotel has been verified and can now be found and reserved by guests.", 0, id)
	return s.repo.Get(ctx, id)
}

func (s *HotelService) Reject(ctx context.Context, actor inbound.Principal, id int64, reason string) (domain.Hotel, error) {
	if err := requireAdmin(actor); err != nil {
		return domain.Hotel{}, err
	}
	if reason = trim(reason); reason == "" {
		return domain.Hotel{}, fmt.Errorf("%w: give the manager a reason", domain.ErrInvalid)
	}
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Hotel{}, err
	}
	if h.Status != domain.HotelPending {
		return domain.Hotel{}, fmt.Errorf("%w: hotel is %s, not waiting for verification", domain.ErrConflict, h.Status.Name())
	}
	if err := s.repo.SetReview(ctx, id, domain.HotelRejected, reason); err != nil {
		return domain.Hotel{}, err
	}
	s.notify.Send(ctx, h.OwnerID, "hotel.rejected", h.Name+" was not approved",
		"Reason: "+reason+". Fix it and it goes back into the review queue.", 0, id)
	return s.repo.Get(ctx, id)
}

func (s *HotelService) Register(ctx context.Context, actor inbound.Principal, in inbound.HotelInput) (domain.Hotel, error) {
	h, err := toHotel(in, nil)
	if err != nil {
		return domain.Hotel{}, err
	}
	switch {
	case !actor.IsAdmin():
		if in.Status != 0 {
			return domain.Hotel{}, fmt.Errorf("%w: status is set by verification, leave it out", domain.ErrInvalid)
		}
		h.Status = domain.HotelPending
	case in.Status == 0:
		h.Status = domain.HotelActive
	case in.Status != domain.HotelActive && in.Status != domain.HotelInactive:
		return domain.Hotel{}, fmt.Errorf("%w: status must be active or inactive", domain.ErrInvalid)
	default:
		h.Status = in.Status
	}
	if in.WalletID != nil {
		h.WalletID = trim(*in.WalletID)
	}
	if err := s.checkWallet(ctx, h.WalletID, actor.UserID, !actor.IsAdmin()); err != nil {
		return domain.Hotel{}, err
	}
	h.OwnerID = actor.UserID
	return s.repo.Create(ctx, h)
}

func (s *HotelService) Update(ctx context.Context, actor inbound.Principal, id int64, in inbound.HotelInput) (domain.Hotel, error) {
	current, err := s.manageable(ctx, actor, id)
	if err != nil {
		return domain.Hotel{}, err
	}
	h, err := toHotel(in, &current)
	if err != nil {
		return domain.Hotel{}, err
	}
	h.ID, h.OwnerID = id, current.OwnerID
	if h.Currency != current.Currency {
		return domain.Hotel{}, fmt.Errorf("%w: currency cannot be changed after registration: rates and payments are in %s", domain.ErrConflict, current.Currency)
	}

	want := in.Status
	if want != 0 && want != domain.HotelActive && want != domain.HotelInactive {
		return domain.Hotel{}, fmt.Errorf("%w: status can only be set to active or inactive; verification decides the rest", domain.ErrInvalid)
	}
	switch current.Status {
	case domain.HotelPending, domain.HotelRejected:
		if want != 0 {
			return domain.Hotel{}, fmt.Errorf("%w: hotel is %s; it cannot be switched on before verification", domain.ErrConflict, current.Status.Name())
		}
		h.Status = domain.HotelPending // a fixed, rejected hotel is resubmitted
		if current.Status == domain.HotelRejected && current.OwnerID != actor.UserID {
			h.Status = domain.HotelRejected // an admin's edit does not resubmit it for the manager
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
		h.WalletID = trim(*in.WalletID)
		if h.WalletID != current.WalletID { // only re-check a changed wallet, against the manager's identity
			if err := s.checkWallet(ctx, h.WalletID, current.OwnerID, !actor.IsAdmin()); err != nil {
				return domain.Hotel{}, err
			}
		}
	}
	return s.repo.Update(ctx, h)
}

func (s *HotelService) Deactivate(ctx context.Context, actor inbound.Principal, id int64) error {
	h, err := s.manageable(ctx, actor, id)
	if err != nil {
		return err
	}
	if h.Status != domain.HotelActive {
		return fmt.Errorf("%w: hotel is %s; only a published one can be paused", domain.ErrConflict, h.Status.Name())
	}
	return s.repo.SetStatus(ctx, id, domain.HotelInactive)
}

func toHotel(in inbound.HotelInput, current *domain.Hotel) (domain.Hotel, error) {
	invalid := func(format string, a ...any) (domain.Hotel, error) {
		return domain.Hotel{}, fmt.Errorf("%w: "+format, append([]any{domain.ErrInvalid}, a...)...)
	}
	name, city, address, phone := trim(in.Name), trim(in.City), trim(in.Address), trim(in.PhoneNumber)
	if name == "" {
		return invalid("name is required")
	}
	if in.Stars != domain.RequiredStars {
		return invalid("only %d-star hotels are managed here", domain.RequiredStars)
	}
	if city == "" {
		return invalid("city is required")
	}
	if address == "" {
		return invalid("address is required")
	}
	if !phonePattern.MatchString(phone) {
		return invalid("a contact phone number is required (8-20 digits, + - ( ) allowed)")
	}
	currency := strings.ToUpper(trim(in.Currency))
	if len(currency) != 3 {
		return invalid("currency must be a 3-letter code")
	}
	checkIn, checkOut := trim(in.CheckInTime), trim(in.CheckOutTime)
	free := defaultFreeCancelHours
	if current != nil {
		if checkIn == "" {
			checkIn = current.CheckInTime
		}
		if checkOut == "" {
			checkOut = current.CheckOutTime
		}
		free = current.FreeCancelHours
	}
	if checkIn == "" {
		checkIn = defaultCheckIn
	}
	if checkOut == "" {
		checkOut = defaultCheckOut
	}
	if !timePattern.MatchString(checkIn) || !timePattern.MatchString(checkOut) {
		return invalid("check-in and check-out times must look like 14:00")
	}
	if in.FreeCancelHours != nil {
		free = *in.FreeCancelHours
	}
	if free < 0 || free > 24*30 {
		return invalid("free cancellation window must be between 0 and 720 hours")
	}
	var tiers []domain.CancellationTier
	if current != nil {
		tiers = current.CancellationTiers
	}
	if in.CancellationTiers != nil {
		tiers = *in.CancellationTiers 
		if err := domain.ValidatePolicy(tiers); err != nil {
			return domain.Hotel{}, err
		}
	}
	return domain.Hotel{
		Name: name, Description: in.Description, Stars: in.Stars, City: city, Address: address,
		PhoneNumber: phone, Email: trim(in.Email), Images: in.Images, Amenities: in.Amenities,
		CheckInTime: checkIn, CheckOutTime: checkOut, Currency: currency, FreeCancelHours: free, CancellationTiers: tiers,
	}, nil
}
