package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

// CatalogService is what a hotel's manager maintains: room types, physical rooms, the nightly
// inventory and rates, and extra services. Reads are public (for published hotels); writes are for
// the hotel's manager or an admin.
type CatalogService struct {
	repo     outbound.HotelRepository
	soldOut  *SoldOutCache // told when inventory changes; may be nil
	promoter Promoter      // offers new rooms to the waiting list; may be nil
	now      func() time.Time
}

// WithPromoter makes new inventory be offered to the waiting list.
func (s *CatalogService) WithPromoter(p Promoter) *CatalogService {
	s.promoter = p
	return s
}

var _ inbound.CatalogUseCase = (*CatalogService)(nil)

func NewCatalogService(repo outbound.HotelRepository, soldOut *SoldOutCache) *CatalogService {
	return &CatalogService{repo: repo, soldOut: soldOut, now: time.Now}
}

func (s *CatalogService) today() time.Time { return s.now().UTC().Truncate(24 * time.Hour) }

// published loads a hotel for a public read.
func (s *CatalogService) published(ctx context.Context, id int64) (domain.Hotel, error) {
	h, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Hotel{}, err
	}
	if h.Status != domain.HotelActive {
		return domain.Hotel{}, domain.ErrNotFound
	}
	return h, nil
}

func (s *CatalogService) manageable(ctx context.Context, actor inbound.Principal, id int64) (domain.Hotel, error) {
	return manageableHotel(ctx, s.repo, actor, id)
}

// ownRoomType loads a room type and checks it belongs to the hotel.
func (s *CatalogService) ownRoomType(ctx context.Context, hotelID, id int64) (domain.RoomType, error) {
	t, err := s.repo.GetRoomType(ctx, id)
	if err != nil {
		return domain.RoomType{}, err
	}
	if t.HotelID != hotelID {
		return domain.RoomType{}, domain.ErrNotFound
	}
	return t, nil
}

func decimalOK(s string) bool {
	v, err := strconv.ParseFloat(s, 64)
	return err == nil && v >= 0
}

// ---- room types

func (s *CatalogService) RoomTypes(ctx context.Context, hotelID int64) ([]domain.RoomType, error) {
	h, err := s.published(ctx, hotelID)
	if err != nil {
		return nil, err
	}
	return activeRoomTypes(h.RoomTypes), nil
}

func toRoomType(in inbound.RoomTypeInput, cur domain.RoomType) (domain.RoomType, error) {
	if trim(in.Name) == "" {
		return domain.RoomType{}, fmt.Errorf("%w: name is required", domain.ErrInvalid)
	}
	if in.Capacity < 1 || in.Capacity > 10 {
		return domain.RoomType{}, fmt.Errorf("%w: capacity must be between 1 and 10 guests per room", domain.ErrInvalid)
	}
	if in.SizeM2 < 0 {
		return domain.RoomType{}, fmt.Errorf("%w: size cannot be negative", domain.ErrInvalid)
	}
	t := cur
	t.Name, t.Description, t.Capacity, t.Bed, t.SizeM2, t.View, t.Amenities =
		trim(in.Name), in.Description, in.Capacity, trim(in.Bed), in.SizeM2, trim(in.View), in.Amenities
	if in.Active != nil {
		t.Active = *in.Active
	}
	return t, nil
}

func (s *CatalogService) CreateRoomType(ctx context.Context, actor inbound.Principal, hotelID int64, in inbound.RoomTypeInput) (domain.RoomType, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.RoomType{}, err
	}
	t, err := toRoomType(in, domain.RoomType{HotelID: hotelID, Active: true})
	if err != nil {
		return domain.RoomType{}, err
	}
	return s.repo.CreateRoomType(ctx, t)
}

func (s *CatalogService) UpdateRoomType(ctx context.Context, actor inbound.Principal, hotelID, id int64, in inbound.RoomTypeInput) (domain.RoomType, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.RoomType{}, err
	}
	cur, err := s.ownRoomType(ctx, hotelID, id)
	if err != nil {
		return domain.RoomType{}, err
	}
	t, err := toRoomType(in, cur)
	if err != nil {
		return domain.RoomType{}, err
	}
	return s.repo.UpdateRoomType(ctx, t)
}

// ---- rooms

func (s *CatalogService) Rooms(ctx context.Context, actor inbound.Principal, hotelID, roomTypeID int64) ([]domain.Room, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return nil, err
	}
	return s.repo.Rooms(ctx, hotelID, roomTypeID)
}

func (s *CatalogService) toRoom(ctx context.Context, hotelID int64, in inbound.RoomInput, cur domain.Room) (domain.Room, error) {
	if trim(in.Name) == "" {
		return domain.Room{}, fmt.Errorf("%w: name is required", domain.ErrInvalid)
	}
	if _, err := s.ownRoomType(ctx, hotelID, in.RoomTypeID); err != nil {
		return domain.Room{}, fmt.Errorf("%w: room_type_id is not a room type of this hotel", domain.ErrInvalid)
	}
	r := cur
	r.HotelID, r.RoomTypeID, r.Name, r.Floor = hotelID, in.RoomTypeID, trim(in.Name), in.Floor
	if in.Available != nil {
		r.Available = *in.Available
	}
	return r, nil
}

func (s *CatalogService) CreateRoom(ctx context.Context, actor inbound.Principal, hotelID int64, in inbound.RoomInput) (domain.Room, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.Room{}, err
	}
	r, err := s.toRoom(ctx, hotelID, in, domain.Room{Available: true})
	if err != nil {
		return domain.Room{}, err
	}
	return s.repo.CreateRoom(ctx, r)
}

func (s *CatalogService) UpdateRoom(ctx context.Context, actor inbound.Principal, hotelID, id int64, in inbound.RoomInput) (domain.Room, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.Room{}, err
	}
	cur, err := s.repo.GetRoom(ctx, id)
	if err != nil {
		return domain.Room{}, err
	}
	if cur.HotelID != hotelID {
		return domain.Room{}, domain.ErrNotFound
	}
	r, err := s.toRoom(ctx, hotelID, in, cur)
	if err != nil {
		return domain.Room{}, err
	}
	return s.repo.UpdateRoom(ctx, r)
}

// ---- inventory and rates

func (s *CatalogService) SetInventory(ctx context.Context, actor inbound.Principal, c inbound.InventoryCommand) error {
	if _, err := s.manageable(ctx, actor, c.HotelID); err != nil {
		return err
	}
	if _, err := s.ownRoomType(ctx, c.HotelID, c.RoomTypeID); err != nil {
		return err
	}
	if err := validRange(c.From, c.To); err != nil {
		return err
	}
	if c.From.Before(s.today()) {
		return fmt.Errorf("%w: cannot change nights that have already begun", domain.ErrInvalid)
	}
	if c.Total != nil && (*c.Total < 0 || *c.Total > 10000) {
		return fmt.Errorf("%w: total rooms must be between 0 and 10000", domain.ErrInvalid)
	}
	if c.Rate != nil && *c.Rate != "" && !decimalOK(*c.Rate) {
		return fmt.Errorf("%w: rate must be a non-negative number", domain.ErrInvalid)
	}
	if c.Total == nil && c.Rate == nil {
		return fmt.Errorf("%w: give a total, a rate, or both", domain.ErrInvalid)
	}
	err := s.repo.SetInventory(ctx, outbound.SetInventoryParams{
		HotelID: c.HotelID, RoomTypeID: c.RoomTypeID, From: c.From, To: c.To, Total: c.Total, Rate: c.Rate,
	})
	if err == nil {
		s.soldOut.Release(c.RoomTypeID) // more rooms or new rates: "sold out" may no longer be true
		if s.promoter != nil {
			_, _ = s.promoter.Promote(context.WithoutCancel(ctx), c.RoomTypeID)
		}
	}
	return err
}

func (s *CatalogService) Inventory(ctx context.Context, hotelID, roomTypeID int64, from, to time.Time) ([]domain.InventoryDay, error) {
	if err := validRange(from, to); err != nil {
		return nil, err
	}
	if _, err := s.published(ctx, hotelID); err != nil {
		return nil, err
	}
	if _, err := s.ownRoomType(ctx, hotelID, roomTypeID); err != nil {
		return nil, err
	}
	return s.repo.Inventory(ctx, hotelID, roomTypeID, from, to)
}

func (s *CatalogService) Quotes(ctx context.Context, hotelID int64, from, to time.Time, rooms int) ([]domain.Quote, error) {
	if err := validRange(from, to); err != nil {
		return nil, err
	}
	if rooms < 1 || rooms > maxRooms {
		return nil, fmt.Errorf("%w: rooms must be between 1 and %d", domain.ErrInvalid, maxRooms)
	}
	if _, err := s.published(ctx, hotelID); err != nil {
		return nil, err
	}
	return s.repo.Quotes(ctx, hotelID, from, to, rooms)
}

// ---- extra services

func (s *CatalogService) Services(ctx context.Context, hotelID int64) ([]domain.HotelService, error) {
	if _, err := s.published(ctx, hotelID); err != nil {
		return nil, err
	}
	return s.repo.Services(ctx, hotelID, true)
}

func toService(in inbound.ServiceInput, cur domain.HotelService) (domain.HotelService, error) {
	if trim(in.Name) == "" {
		return domain.HotelService{}, fmt.Errorf("%w: name is required", domain.ErrInvalid)
	}
	if !decimalOK(in.Price) {
		return domain.HotelService{}, fmt.Errorf("%w: price must be a non-negative number", domain.ErrInvalid)
	}
	if !in.Unit.Valid() {
		return domain.HotelService{}, fmt.Errorf("%w: unit must be %q or %q", domain.ErrInvalid, domain.UnitStay, domain.UnitNight)
	}
	v := cur
	v.Name, v.Description, v.Price, v.Unit = trim(in.Name), in.Description, in.Price, in.Unit
	if in.Active != nil {
		v.Active = *in.Active
	}
	return v, nil
}

func (s *CatalogService) CreateService(ctx context.Context, actor inbound.Principal, hotelID int64, in inbound.ServiceInput) (domain.HotelService, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.HotelService{}, err
	}
	v, err := toService(in, domain.HotelService{HotelID: hotelID, Active: true})
	if err != nil {
		return domain.HotelService{}, err
	}
	return s.repo.SaveService(ctx, v)
}

func (s *CatalogService) UpdateService(ctx context.Context, actor inbound.Principal, hotelID, id int64, in inbound.ServiceInput) (domain.HotelService, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.HotelService{}, err
	}
	all, err := s.repo.Services(ctx, hotelID, false)
	if err != nil {
		return domain.HotelService{}, err
	}
	for _, cur := range all {
		if cur.ID == id {
			v, err := toService(in, cur)
			if err != nil {
				return domain.HotelService{}, err
			}
			return s.repo.SaveService(ctx, v)
		}
	}
	return domain.HotelService{}, domain.ErrNotFound
}

// ---- promo codes

var codePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)

func toPromotion(in inbound.PromotionInput, cur domain.Promotion) (domain.Promotion, error) {
	code := trim(in.Code)
	if !codePattern.MatchString(code) {
		return domain.Promotion{}, fmt.Errorf("%w: the code must be 3-32 letters, digits, - or _", domain.ErrInvalid)
	}
	if (in.PercentOff > 0) == (trim(in.AmountOff) != "") {
		return domain.Promotion{}, fmt.Errorf("%w: give either percent_off or amount_off", domain.ErrInvalid)
	}
	if in.PercentOff < 0 || in.PercentOff > 100 {
		return domain.Promotion{}, fmt.Errorf("%w: percent_off must be between 1 and 100", domain.ErrInvalid)
	}
	if in.AmountOff != "" {
		if v, err := strconv.ParseFloat(in.AmountOff, 64); err != nil || v <= 0 {
			return domain.Promotion{}, fmt.Errorf("%w: amount_off must be a positive number", domain.ErrInvalid)
		}
	}
	if in.MinNights < 0 || in.MaxUses < 0 {
		return domain.Promotion{}, fmt.Errorf("%w: min_nights and max_uses cannot be negative", domain.ErrInvalid)
	}
	if in.ValidFrom != nil && in.ValidTo != nil && in.ValidTo.Before(*in.ValidFrom) {
		return domain.Promotion{}, fmt.Errorf("%w: valid_to is before valid_from", domain.ErrInvalid)
	}
	p := cur
	p.Code, p.Description, p.PercentOff, p.AmountOff = code, in.Description, in.PercentOff, trim(in.AmountOff)
	p.MinNights, p.MaxUses, p.ValidFrom, p.ValidTo = max(in.MinNights, 1), in.MaxUses, in.ValidFrom, in.ValidTo
	if in.Active != nil {
		p.Active = *in.Active
	}
	if p.MaxUses > 0 && p.MaxUses < p.UsedCount {
		return domain.Promotion{}, fmt.Errorf("%w: %d guests have already used this code, so max_uses cannot be lower", domain.ErrInvalid, p.UsedCount)
	}
	return p, nil
}

func (s *CatalogService) Promotions(ctx context.Context, actor inbound.Principal, hotelID int64) ([]domain.Promotion, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return nil, err
	}
	return s.repo.Promotions(ctx, hotelID)
}

func (s *CatalogService) CreatePromotion(ctx context.Context, actor inbound.Principal, hotelID int64, in inbound.PromotionInput) (domain.Promotion, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.Promotion{}, err
	}
	p, err := toPromotion(in, domain.Promotion{HotelID: hotelID, Active: true})
	if err != nil {
		return domain.Promotion{}, err
	}
	return s.repo.SavePromotion(ctx, p)
}

func (s *CatalogService) UpdatePromotion(ctx context.Context, actor inbound.Principal, hotelID, id int64, in inbound.PromotionInput) (domain.Promotion, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.Promotion{}, err
	}
	all, err := s.repo.Promotions(ctx, hotelID)
	if err != nil {
		return domain.Promotion{}, err
	}
	for _, cur := range all {
		if cur.ID == id {
			p, err := toPromotion(in, cur)
			if err != nil {
				return domain.Promotion{}, err
			}
			return s.repo.SavePromotion(ctx, p)
		}
	}
	return domain.Promotion{}, domain.ErrNotFound
}

// ---- report

const maxReportDays = 92

func (s *CatalogService) Report(ctx context.Context, actor inbound.Principal, hotelID int64, from, to time.Time) (domain.HotelReport, error) {
	if _, err := s.manageable(ctx, actor, hotelID); err != nil {
		return domain.HotelReport{}, err
	}
	if !to.After(from) || to.Sub(from) > maxReportDays*24*time.Hour {
		return domain.HotelReport{}, fmt.Errorf("%w: the range must be between 1 and %d days", domain.ErrInvalid, maxReportDays)
	}
	return s.repo.Report(ctx, hotelID, from, to)
}
