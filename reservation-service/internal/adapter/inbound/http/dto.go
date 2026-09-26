package http

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const dateLayout = "2006-01-02"

func parseDate(field, v string) (time.Time, error) {
	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s must be YYYY-MM-DD", domain.ErrInvalid, field)
	}
	return t, nil
}

// ---- hotels

type hotelRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Stars           int      `json:"stars"` // must be 5
	City            string   `json:"city"`
	Address         string   `json:"address"`
	PhoneNumber     string   `json:"phone_number"`
	Email           string   `json:"email"`
	Images          []string `json:"images"`
	Amenities       []string `json:"amenities"`
	CheckInTime     string   `json:"check_in_time"`  // "15:00"; default 14:00
	CheckOutTime    string   `json:"check_out_time"` // default 12:00
	Currency        string   `json:"currency"`
	FreeCancelHours *int     `json:"free_cancel_hours"` // refund without fee until this long before check-in (default 48)
	// CancellationTiers is the cancellation policy, e.g. [{"hours_before":72,"fee_percent":0},{"hours_before":24,"fee_percent":50}]:
	// free 72h or more before check-in, half the price 24-72h before, non-refundable after that. Omit to keep the
	// current one; [] goes back to "free until free_cancel_hours before check-in".
	CancellationTiers *[]tierDTO `json:"cancellation_tiers"`
	// WalletID is the payment-wallet-service wallet that receives this hotel's payments. On update,
	// omit it to keep the current one, or send "" to remove it.
	WalletID *string `json:"wallet_id"`
	Status   int16   `json:"status"` // on update only: 1 active, 2 inactive
}

type tierDTO struct {
	HoursBefore int `json:"hours_before"`
	FeePercent  int `json:"fee_percent"`
}

type hotelResponse struct {
	ID                 int64              `json:"id"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	Stars              int                `json:"stars"`
	City               string             `json:"city"`
	Address            string             `json:"address"`
	PhoneNumber        string             `json:"phone_number"`
	Email              string             `json:"email,omitempty"`
	Images             []string           `json:"images"`
	Amenities          []string           `json:"amenities"`
	CheckInTime        string             `json:"check_in_time"`
	CheckOutTime       string             `json:"check_out_time"`
	Currency           string             `json:"currency"`
	FreeCancelHours    int                `json:"free_cancel_hours"`
	CancellationPolicy []tierDTO          `json:"cancellation_policy"` // the policy that applies
	Status             int16              `json:"status"`              // 1 active, 2 inactive, 3 pending verification, 4 rejected
	StatusName         string             `json:"status_name"`
	ReviewNote         string             `json:"review_note,omitempty"` // the admin's reason when rejected
	ManagerID          int64              `json:"manager_id"`
	AcceptsWallet      bool               `json:"accepts_wallet_payment"`
	Rating             float64            `json:"rating"` // average stars, 0 when unreviewed
	ReviewCount        int                `json:"review_count"`
	RoomTypes          []roomTypeResponse `json:"room_types,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
}

func toHotelResponse(h domain.Hotel) hotelResponse {
	r := hotelResponse{
		ID: h.ID, Name: h.Name, Description: h.Description, Stars: h.Stars, City: h.City, Address: h.Address,
		PhoneNumber: h.PhoneNumber, Email: h.Email, Images: nonNil(h.Images), Amenities: nonNil(h.Amenities),
		CheckInTime: h.CheckInTime, CheckOutTime: h.CheckOutTime, Currency: h.Currency, FreeCancelHours: h.FreeCancelHours,
		CancellationPolicy: toTierDTOs(h.EffectivePolicy()),
		Status:             int16(h.Status), StatusName: h.Status.Name(), ReviewNote: h.ReviewNote, ManagerID: h.OwnerID,
		AcceptsWallet: h.WalletID != "", Rating: math.Round(h.Rating*100) / 100, ReviewCount: h.ReviewCount, CreatedAt: h.CreatedAt,
	}
	for _, t := range h.RoomTypes {
		r.RoomTypes = append(r.RoomTypes, toRoomTypeResponse(t))
	}
	return r
}

func toTierDTOs(in []domain.CancellationTier) []tierDTO {
	out := make([]tierDTO, len(in))
	for i, t := range in {
		out[i] = tierDTO{t.HoursBefore, t.FeePercent}
	}
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// hotelPageResponse is one page of a search. To get the next page, repeat the same request with
// cursor set to next_cursor; it is absent on the last page.
type hotelPageResponse struct {
	Items      []hotelResponse `json:"items"`
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// pageResponse is one page of a list. To get the next page, repeat the same request with cursor set to
// next_cursor; it is absent on the last page.
type pageResponse[T any] struct {
	Items      []T    `json:"items"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func toPage[S, T any](pg domain.Page[S], convert func(S) T) pageResponse[T] {
	out := pageResponse[T]{Items: make([]T, len(pg.Items)), HasMore: pg.NextID != 0, NextCursor: encodeIDCursor(pg.NextID)}
	for i, x := range pg.Items {
		out.Items[i] = convert(x)
	}
	return out
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

// ---- room types, rooms, inventory, services

type roomTypeRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Capacity    int      `json:"capacity"` // guests per room
	Bed         string   `json:"bed"`
	SizeM2      int      `json:"size_m2"`
	View        string   `json:"view"`
	Amenities   []string `json:"amenities"`
	Active      *bool    `json:"active"`
}

type roomTypeResponse struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Capacity    int      `json:"capacity"`
	Bed         string   `json:"bed,omitempty"`
	SizeM2      int      `json:"size_m2,omitempty"`
	View        string   `json:"view,omitempty"`
	Amenities   []string `json:"amenities"`
	Active      bool     `json:"active"`
}

func toRoomTypeResponse(t domain.RoomType) roomTypeResponse {
	return roomTypeResponse{ID: t.ID, Name: t.Name, Description: t.Description, Capacity: t.Capacity, Bed: t.Bed,
		SizeM2: t.SizeM2, View: t.View, Amenities: nonNil(t.Amenities), Active: t.Active}
}

type roomRequest struct {
	RoomTypeID int64  `json:"room_type_id"`
	Name       string `json:"name"`
	Floor      int    `json:"floor"`
	Available  *bool  `json:"available"`
}

type roomResponse struct {
	ID         int64  `json:"id"`
	RoomTypeID int64  `json:"room_type_id"`
	Name       string `json:"name"`
	Floor      int    `json:"floor"`
	Available  bool   `json:"available"`
}

type inventoryRequest struct {
	From string `json:"from"`
	To   string `json:"to"` // exclusive
	// Total is how many rooms are for sale each night; omit it to use the number of available rooms.
	Total *int `json:"total_inventory"`
	// Rate is the price per room per night; omit it to keep the current rates.
	Rate *json.Number `json:"rate"`
	// RateOff takes the nights off sale (they cannot be reserved until a rate is set again).
	RateOff bool `json:"rate_off"`
}

type inventoryResponse struct {
	Date     string `json:"date"`
	Total    int    `json:"total_inventory"`
	Reserved int    `json:"total_reserved"`
	Left     int    `json:"left"`
	Rate     string `json:"rate,omitempty"`
}

type quoteResponse struct {
	RoomType  roomTypeResponse `json:"room_type"`
	Nights    int              `json:"nights"`
	Available int              `json:"rooms_available"` // for every night of the stay
	Total     string           `json:"total,omitempty"` // all rooms, all nights; absent when not for sale
	Bookable  bool             `json:"bookable"`        // enough rooms and every night for sale
}

type serviceRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       json.Number `json:"price"`
	Unit        string      `json:"unit"` // "stay" or "night"
	Active      *bool       `json:"active"`
}

type serviceResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Price       string `json:"price"`
	Unit        string `json:"unit"`
	Active      bool   `json:"active"`
}

func toServiceResponse(s domain.HotelService) serviceResponse {
	return serviceResponse{ID: s.ID, Name: s.Name, Description: s.Description, Price: s.Price, Unit: string(s.Unit), Active: s.Active}
}

// ---- reservations

type addonRequest struct {
	ServiceID int64 `json:"service_id"`
	Quantity  int   `json:"quantity"`
}

type reservationRequest struct {
	HotelID         int64          `json:"hotel_id"`
	RoomTypeID      int64          `json:"room_type_id"`
	CheckIn         string         `json:"check_in"`
	CheckOut        string         `json:"check_out"`
	Rooms           int            `json:"rooms"`
	Adults          int            `json:"adults"`
	Children        int            `json:"children"`
	SpecialRequests string         `json:"special_requests"`
	Services        []addonRequest `json:"services"`
	PromoCode       string         `json:"promo_code"` // a code given out by the hotel: a discount on the room charge
	RequestID       string         `json:"request_id"`
}

type addonResponse struct {
	ServiceID int64  `json:"service_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice string `json:"unit_price"`
	Amount    string `json:"amount"`
}

type reservationResponse struct {
	ID              int64           `json:"id"`
	HotelID         int64           `json:"hotel_id"`
	RoomTypeID      int64           `json:"room_type_id"`
	GuestID         int64           `json:"guest_id"`
	CheckIn         string          `json:"check_in"`
	CheckOut        string          `json:"check_out"`
	Nights          int             `json:"nights"`
	Rooms           int             `json:"rooms"`
	Adults          int             `json:"adults"`
	Children        int             `json:"children"`
	Status          int16           `json:"status"` // 1 pending, 2 paid, 3 canceled, 4 rejected, 5 refunded
	StatusName      string          `json:"status_name"`
	StatusNote      string          `json:"status_note,omitempty"`
	Currency        string          `json:"currency"`
	RoomTotal       string          `json:"room_total"`
	PromoCode       string          `json:"promo_code,omitempty"`
	Discount        string          `json:"discount,omitempty"` // taken off room_total by the promo code
	AddonsTotal     string          `json:"services_total"`
	TotalAmount     string          `json:"total_amount"`               // room_total - discount + services_total
	FeeAmount       string          `json:"cancellation_fee,omitempty"` // kept by the hotel when cancelled late
	RefundAmount    string          `json:"refund_amount,omitempty"`    // paid back when a paid reservation was cancelled
	Stay            string          `json:"stay_status"`                // not_arrived, checked_in, checked_out, no_show
	CheckedInAt     *time.Time      `json:"checked_in_at,omitempty"`
	CheckedOutAt    *time.Time      `json:"checked_out_at,omitempty"`
	SpecialRequests string          `json:"special_requests,omitempty"`
	Services        []addonResponse `json:"services,omitempty"`
	RequestID       string          `json:"request_id"`
	PaymentMethod   string          `json:"payment_method,omitempty"`
	PaymentRef      string          `json:"payment_ref,omitempty"`
	PaidAt          *time.Time      `json:"paid_at,omitempty"`
	// ExpiresAt is when the held rooms are released if the reservation is still unpaid.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func toReservationResponse(x domain.Reservation) reservationResponse {
	r := reservationResponse{
		ID: x.ID, HotelID: x.HotelID, RoomTypeID: x.RoomTypeID, GuestID: x.GuestID,
		CheckIn: x.Start.Format(dateLayout), CheckOut: x.End.Format(dateLayout), Nights: x.Nights(),
		Rooms: x.Rooms, Adults: x.Adults, Children: x.Children, Status: int16(x.Status), StatusName: x.Status.Name(),
		StatusNote: x.StatusNote, Currency: x.Currency, RoomTotal: x.RoomTotal, AddonsTotal: x.AddonsTotal,
		TotalAmount: x.TotalAmount, PromoCode: x.PromoCode, Stay: x.Stay.Name(), CheckedInAt: x.CheckedInAt, CheckedOutAt: x.CheckedOutAt, SpecialRequests: x.SpecialRequests, RequestID: x.RequestID,
		PaymentMethod: string(x.PaymentMethod), PaymentRef: x.PaymentRef, PaidAt: x.PaidAt, CreatedAt: x.CreatedAt,
	}
	if x.Status == domain.ReservationPending {
		r.ExpiresAt = x.ExpiresAt
	}
	if x.PromoID != 0 {
		r.Discount = x.DiscountAmount
	}
	if x.Status == domain.ReservationRefunded {
		r.FeeAmount, r.RefundAmount = x.FeeAmount, x.RefundAmount
	}
	for _, a := range x.Addons {
		r.Services = append(r.Services, addonResponse{ServiceID: a.ServiceID, Name: a.Name, Quantity: a.Quantity, UnitPrice: a.UnitPrice, Amount: a.Amount})
	}
	return r
}

type payRequest struct {
	Method   string `json:"method"`   // "wallet" or "gateway"
	Provider string `json:"provider"` // gateway only: paypal, stripe, ...
}

// ---- reviews, wallet, loyalty

type reviewRequest struct {
	ReservationID int64  `json:"reservation_id"`
	Rating        int    `json:"rating"` // 1-5
	Comment       string `json:"comment"`
}

type reviewResponse struct {
	ID            int64     `json:"id"`
	HotelID       int64     `json:"hotel_id"`
	UserID        int64     `json:"user_id"`
	ReservationID int64     `json:"reservation_id"`
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type loyaltyResponse struct {
	Points   int64  `json:"points"`
	Tier     int    `json:"tier"`
	TierName string `json:"tier_name"`
}

type walletRequest struct {
	Currency string `json:"currency"`
}

type walletResponse struct {
	ID       string `json:"wallet_id"`
	Balance  int64  `json:"balance"` // minor units
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

type walletTxnResponse struct {
	ID          string    `json:"transaction_id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	ReferenceID string    `json:"reference_id,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type cancellationQuoteResponse struct {
	Status      string    `json:"status_name"`
	HoursLeft   float64   `json:"hours_before_check_in"`
	FeePercent  int       `json:"fee_percent"`
	Fee         string    `json:"fee"`    // the hotel would keep this
	Refund      string    `json:"refund"` // and this would be paid back
	Cancellable bool      `json:"cancellable"`
	Reason      string    `json:"reason,omitempty"`
	Policy      []tierDTO `json:"policy"`
}

type notificationResponse struct {
	ID            int64     `json:"id"`
	Kind          string    `json:"kind"`
	Title         string    `json:"title"`
	Body          string    `json:"body"`
	ReservationID int64     `json:"reservation_id,omitempty"`
	HotelID       int64     `json:"hotel_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	Read          bool      `json:"read"`
}

func toNotificationResponse(n domain.Notification) notificationResponse {
	return notificationResponse{ID: n.ID, Kind: n.Kind, Title: n.Title, Body: n.Body, ReservationID: n.ReservationID,
		HotelID: n.HotelID, CreatedAt: n.CreatedAt, Read: n.ReadAt != nil}
}

type promotionRequest struct {
	Code        string       `json:"code"`
	Description string       `json:"description"`
	PercentOff  int          `json:"percent_off"` // 1-100, or use amount_off
	AmountOff   *json.Number `json:"amount_off"`  // a fixed amount off the room charge
	MinNights   int          `json:"min_nights"`
	MaxUses     int          `json:"max_uses"`   // 0 = unlimited
	ValidFrom   string       `json:"valid_from"` // YYYY-MM-DD: the first check-in date the code works for
	ValidTo     string       `json:"valid_to"`   // the last check-in date
	Active      *bool        `json:"active"`
}

type promotionResponse struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	PercentOff  int    `json:"percent_off,omitempty"`
	AmountOff   string `json:"amount_off,omitempty"`
	MinNights   int    `json:"min_nights"`
	MaxUses     int    `json:"max_uses,omitempty"`
	UsedCount   int    `json:"used_count"`
	ValidFrom   string `json:"valid_from,omitempty"`
	ValidTo     string `json:"valid_to,omitempty"`
	Active      bool   `json:"active"`
}

func toPromotionResponse(p domain.Promotion) promotionResponse {
	r := promotionResponse{ID: p.ID, Code: p.Code, Description: p.Description, PercentOff: p.PercentOff, AmountOff: p.AmountOff,
		MinNights: p.MinNights, MaxUses: p.MaxUses, UsedCount: p.UsedCount, Active: p.Active}
	if p.ValidFrom != nil {
		r.ValidFrom = p.ValidFrom.Format(dateLayout)
	}
	if p.ValidTo != nil {
		r.ValidTo = p.ValidTo.Format(dateLayout)
	}
	return r
}

type reportResponse struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Occupancy struct {
		RoomNightsAvailable int            `json:"room_nights_available"`
		RoomNightsSold      int            `json:"room_nights_sold"`
		OccupancyPercent    float64        `json:"occupancy_percent"`
		Days                []occupancyDay `json:"days"`
	} `json:"occupancy"`
	Revenue struct {
		Currency       string `json:"currency"`
		PaidCount      int    `json:"paid_reservations"`
		PaidRevenue    string `json:"paid_revenue"`
		RefundedCount  int    `json:"refunded_reservations"`
		RefundedAmount string `json:"refunded_amount"`
		FeesKept       string `json:"cancellation_fees_kept"`
		NetRevenue     string `json:"net_revenue"` // paid revenue + cancellation fees kept
	} `json:"revenue"`
	Reservations map[string]int `json:"reservations_by_status"`
	NoShows      int            `json:"no_shows"`
}

type occupancyDay struct {
	Date             string  `json:"date"`
	Rooms            int     `json:"rooms"`
	Reserved         int     `json:"reserved"`
	OccupancyPercent float64 `json:"occupancy_percent"`
}

type historyEntryResponse struct {
	At    time.Time `json:"at"`
	By    string    `json:"by"` // guest, hotel or system
	Event string    `json:"event"`
	From  string    `json:"from_status,omitempty"`
	To    string    `json:"to_status,omitempty"`
	Note  string    `json:"note,omitempty"`
}

type waitRequest struct {
	HotelID    int64  `json:"hotel_id"`
	RoomTypeID int64  `json:"room_type_id"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	Rooms      int    `json:"rooms"`
	Adults     int    `json:"adults"`
	Children   int    `json:"children"`
}

type waitResponse struct {
	ID         int64  `json:"id"`
	HotelID    int64  `json:"hotel_id"`
	RoomTypeID int64  `json:"room_type_id"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	Rooms      int    `json:"rooms"`
	Adults     int    `json:"adults"`
	Children   int    `json:"children"`
	// State: waiting; offered (rooms are held, pay the reservation); fulfilled (paid); lapsed (the offer was
	// not paid in time, or cancelled); left (the guest left the list).
	State         string     `json:"state"`
	ReservationID int64      `json:"reservation_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	OfferedAt     *time.Time `json:"offered_at,omitempty"`
}

func toWaitResponse(e domain.WaitlistEntry) waitResponse {
	state := "waiting"
	switch e.Status {
	case domain.WaitlistCancelled:
		state = "left"
	case domain.WaitlistOffered:
		switch e.OfferStatus {
		case domain.ReservationPending:
			state = "offered"
		case domain.ReservationPaid:
			state = "fulfilled"
		default:
			state = "lapsed"
		}
	}
	return waitResponse{ID: e.ID, HotelID: e.HotelID, RoomTypeID: e.RoomTypeID, CheckIn: e.Start.Format(dateLayout), CheckOut: e.End.Format(dateLayout),
		Rooms: e.Rooms, Adults: e.Adults, Children: e.Children, State: state, ReservationID: e.ReservationID, CreatedAt: e.CreatedAt, OfferedAt: e.OfferedAt}
}
