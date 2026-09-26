package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

const maxBodyBytes = 1 << 20

type Handler struct {
	auth          inbound.AuthUseCase
	hotels        inbound.HotelUseCase
	catalog       inbound.CatalogUseCase
	reservations  inbound.ReservationUseCase
	reviews       inbound.ReviewUseCase
	wallets       inbound.WalletUseCase
	loyalty       inbound.LoyaltyUseCase
	notifications inbound.NotificationUseCase
	wishlist      inbound.WishlistUseCase
	waitlist      inbound.WaitlistUseCase
	log           *slog.Logger
}

func NewHandler(a inbound.AuthUseCase, h inbound.HotelUseCase, c inbound.CatalogUseCase, r inbound.ReservationUseCase,
	rv inbound.ReviewUseCase, w inbound.WalletUseCase, l inbound.LoyaltyUseCase, n inbound.NotificationUseCase, wl inbound.WishlistUseCase, wt inbound.WaitlistUseCase, log *slog.Logger) *Handler {
	return &Handler{auth: a, hotels: h, catalog: c, reservations: r, reviews: rv, wallets: w, loyalty: l, notifications: n, wishlist: wl, waitlist: wt, log: log}
}

func today() time.Time { return time.Now().UTC().Truncate(24 * time.Hour) }

// ---- hotels

func toHotelInput(r hotelRequest) inbound.HotelInput {
	return inbound.HotelInput{
		Name: r.Name, Description: r.Description, Stars: r.Stars, City: r.City, Address: r.Address, PhoneNumber: r.PhoneNumber,
		Email: r.Email, Images: r.Images, Amenities: r.Amenities, CheckInTime: r.CheckInTime, CheckOutTime: r.CheckOutTime,
		Currency: r.Currency, FreeCancelHours: r.FreeCancelHours, WalletID: r.WalletID, Status: domain.HotelStatus(r.Status),
		CancellationTiers: toTiers(r.CancellationTiers),
	}
}

func toTiers(in *[]tierDTO) *[]domain.CancellationTier {
	if in == nil {
		return nil
	}
	out := make([]domain.CancellationTier, len(*in))
	for i, t := range *in {
		out[i] = domain.CancellationTier{HoursBefore: t.HoursBefore, FeePercent: t.FeePercent}
	}
	return &out
}

func (h *Handler) registerHotel(w http.ResponseWriter, r *http.Request) {
	var req hotelRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.hotels.Register(r.Context(), mustPrincipal(r), toHotelInput(req))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toHotelResponse(x))
}

func (h *Handler) updateHotel(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req hotelRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.hotels.Update(r.Context(), mustPrincipal(r), id, toHotelInput(req))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHotelResponse(x))
}

func (h *Handler) deactivateHotel(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.hotels.Deactivate(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getHotel(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var viewer *inbound.Principal
	if p, ok := principalFrom(r.Context()); ok {
		viewer = &p
	}
	x, err := h.hotels.View(r.Context(), viewer, id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHotelResponse(x)) // the wallet id itself is never part of a response
}

// searchHotels lists published hotels a page at a time. Repeat the request with cursor=<next_cursor>
// for the next page.
func (h *Handler) searchHotels(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.HotelFilter{
		City: q.Get("city"), Query: q.Get("q"), Rooms: atoi(q.Get("rooms")), Guests: atoi(q.Get("guests")),
		Sort: q.Get("sort"), Limit: atoi(q.Get("limit")), MinRate: q.Get("min_rate"), MaxRate: q.Get("max_rate"),
	}
	for _, a := range strings.Split(q.Get("amenities"), ",") {
		if a = strings.TrimSpace(a); a != "" {
			f.Amenities = append(f.Amenities, a)
		}
	}
	if c := q.Get("cursor"); c != "" {
		var err error
		if f.After, err = decodeCursor(c); err != nil {
			h.fail(w, err)
			return
		}
	}
	if q.Get("check_in") != "" || q.Get("check_out") != "" {
		var err error
		if f.CheckIn, err = parseDate("check_in", q.Get("check_in")); err != nil {
			h.fail(w, err)
			return
		}
		if f.CheckOut, err = parseDate("check_out", q.Get("check_out")); err != nil {
			h.fail(w, err)
			return
		}
	}
	pg, err := h.hotels.Search(r.Context(), f)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]hotelResponse, len(pg.Items))
	for i, x := range pg.Items {
		out[i] = toHotelResponse(x)
	}
	h.write(w, http.StatusOK, hotelPageResponse{Items: out, HasMore: pg.Next != nil, NextCursor: encodeCursor(pg.Next)})
}

func (h *Handler) hotelCursor(w http.ResponseWriter, r *http.Request) (*domain.HotelCursor, bool) {
	c := r.URL.Query().Get("cursor")
	if c == "" {
		return nil, true
	}
	cur, err := decodeCursor(c)
	if err != nil {
		h.fail(w, err)
		return nil, false
	}
	return cur, true
}

func (h *Handler) writeHotelPage(w http.ResponseWriter, pg domain.HotelPage) {
	out := make([]hotelResponse, len(pg.Items))
	for i, x := range pg.Items {
		out[i] = toHotelResponse(x)
	}
	h.write(w, http.StatusOK, hotelPageResponse{Items: out, HasMore: pg.Next != nil, NextCursor: encodeCursor(pg.Next)})
}

func (h *Handler) myHotels(w http.ResponseWriter, r *http.Request) {
	cur, ok := h.hotelCursor(w, r)
	if !ok {
		return
	}
	pg, err := h.hotels.Mine(r.Context(), mustPrincipal(r), cur, atoi(r.URL.Query().Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeHotelPage(w, pg)
}

// reviewQueue is the admin's list of registrations by status (default: waiting for verification).
func (h *Handler) reviewQueue(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := map[string]domain.HotelStatus{"": 0, "pending": domain.HotelPending, "rejected": domain.HotelRejected,
		"active": domain.HotelActive, "inactive": domain.HotelInactive}
	st, ok := status[q.Get("status")]
	if !ok {
		h.fail(w, domain.ErrInvalid)
		return
	}
	cur, ok := h.hotelCursor(w, r)
	if !ok {
		return
	}
	pg, err := h.hotels.ForReview(r.Context(), mustPrincipal(r), st, cur, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writeHotelPage(w, pg)
}

func (h *Handler) approveHotel(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	x, err := h.hotels.Approve(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHotelResponse(x))
}

func (h *Handler) rejectHotel(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req rejectRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.hotels.Reject(r.Context(), mustPrincipal(r), id, req.Reason)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHotelResponse(x))
}

func (h *Handler) writeHotels(w http.ResponseWriter, list []domain.Hotel) {
	out := make([]hotelResponse, len(list))
	for i, x := range list {
		out[i] = toHotelResponse(x)
	}
	h.write(w, http.StatusOK, out)
}

// ---- room types, rooms, inventory, services

func (h *Handler) listRoomTypes(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	list, err := h.catalog.RoomTypes(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]roomTypeResponse, len(list))
	for i, t := range list {
		out[i] = toRoomTypeResponse(t)
	}
	h.write(w, http.StatusOK, out)
}

func toRoomTypeInput(r roomTypeRequest) inbound.RoomTypeInput {
	return inbound.RoomTypeInput{Name: r.Name, Description: r.Description, Capacity: r.Capacity, Bed: r.Bed,
		SizeM2: r.SizeM2, View: r.View, Amenities: r.Amenities, Active: r.Active}
}

func (h *Handler) createRoomType(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req roomTypeRequest
	if !h.decode(w, r, &req) {
		return
	}
	t, err := h.catalog.CreateRoomType(r.Context(), mustPrincipal(r), id, toRoomTypeInput(req))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toRoomTypeResponse(t))
}

func (h *Handler) updateRoomType(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	tid, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	var req roomTypeRequest
	if !h.decode(w, r, &req) {
		return
	}
	t, err := h.catalog.UpdateRoomType(r.Context(), mustPrincipal(r), id, tid, toRoomTypeInput(req))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toRoomTypeResponse(t))
}

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	list, err := h.catalog.Rooms(r.Context(), mustPrincipal(r), id, int64(atoi(r.URL.Query().Get("room_type_id"))))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]roomResponse, len(list))
	for i, x := range list {
		out[i] = toRoomResponse(x)
	}
	h.write(w, http.StatusOK, out)
}

func toRoomResponse(x domain.Room) roomResponse {
	return roomResponse{ID: x.ID, RoomTypeID: x.RoomTypeID, Name: x.Name, Floor: x.Floor, Available: x.Available}
}

func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req roomRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.catalog.CreateRoom(r.Context(), mustPrincipal(r), id, inbound.RoomInput{RoomTypeID: req.RoomTypeID, Name: req.Name, Floor: req.Floor, Available: req.Available})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toRoomResponse(x))
}

func (h *Handler) updateRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	rid, ok := h.pathID(w, r, "room")
	if !ok {
		return
	}
	var req roomRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.catalog.UpdateRoom(r.Context(), mustPrincipal(r), id, rid, inbound.RoomInput{RoomTypeID: req.RoomTypeID, Name: req.Name, Floor: req.Floor, Available: req.Available})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toRoomResponse(x))
}

func (h *Handler) setInventory(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	tid, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	var req inventoryRequest
	if !h.decode(w, r, &req) {
		return
	}
	from, err := parseDate("from", req.From)
	if err != nil {
		h.fail(w, err)
		return
	}
	to, err := parseDate("to", req.To)
	if err != nil {
		h.fail(w, err)
		return
	}
	cmd := inbound.InventoryCommand{HotelID: id, RoomTypeID: tid, From: from, To: to, Total: req.Total}
	switch {
	case req.RateOff && req.Rate != nil:
		h.fail(w, errors.Join(domain.ErrInvalid, errors.New("give either rate or rate_off, not both")))
		return
	case req.RateOff:
		off := ""
		cmd.Rate = &off
	case req.Rate != nil:
		rate := req.Rate.String()
		cmd.Rate = &rate
	}
	if err := h.catalog.SetInventory(r.Context(), mustPrincipal(r), cmd); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getInventory(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	tid, ok := h.pathID(w, r, "type")
	if !ok {
		return
	}
	from, err := parseDate("from", r.URL.Query().Get("from"))
	if err != nil {
		h.fail(w, err)
		return
	}
	to, err := parseDate("to", r.URL.Query().Get("to"))
	if err != nil {
		h.fail(w, err)
		return
	}
	days, err := h.catalog.Inventory(r.Context(), id, tid, from, to)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]inventoryResponse, len(days))
	for i, d := range days {
		out[i] = inventoryResponse{Date: d.Date.Format(dateLayout), Total: d.Total, Reserved: d.Reserved, Left: d.Left(), Rate: d.Rate}
	}
	h.write(w, http.StatusOK, out)
}

// quotes prices a stay for every room type of the hotel: what it costs and how many rooms are left.
func (h *Handler) quotes(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	q := r.URL.Query()
	in, err := parseDate("check_in", q.Get("check_in"))
	if err != nil {
		h.fail(w, err)
		return
	}
	out, err := parseDate("check_out", q.Get("check_out"))
	if err != nil {
		h.fail(w, err)
		return
	}
	rooms := atoi(q.Get("rooms"))
	if rooms == 0 {
		rooms = 1
	}
	list, err := h.catalog.Quotes(r.Context(), id, in, out, rooms)
	if err != nil {
		h.fail(w, err)
		return
	}
	resp := make([]quoteResponse, len(list))
	for i, x := range list {
		resp[i] = quoteResponse{RoomType: toRoomTypeResponse(x.RoomType), Nights: x.Nights, Available: x.Available, Total: x.Total,
			Bookable: x.Total != "" && x.Available >= rooms}
	}
	h.write(w, http.StatusOK, resp)
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	list, err := h.catalog.Services(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]serviceResponse, len(list))
	for i, s := range list {
		out[i] = toServiceResponse(s)
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) createService(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req serviceRequest
	if !h.decode(w, r, &req) {
		return
	}
	s, err := h.catalog.CreateService(r.Context(), mustPrincipal(r), id, inbound.ServiceInput{
		Name: req.Name, Description: req.Description, Price: req.Price.String(), Unit: domain.AddonUnit(req.Unit), Active: req.Active})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toServiceResponse(s))
}

func (h *Handler) updateService(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	sid, ok := h.pathID(w, r, "service")
	if !ok {
		return
	}
	var req serviceRequest
	if !h.decode(w, r, &req) {
		return
	}
	s, err := h.catalog.UpdateService(r.Context(), mustPrincipal(r), id, sid, inbound.ServiceInput{
		Name: req.Name, Description: req.Description, Price: req.Price.String(), Unit: domain.AddonUnit(req.Unit), Active: req.Active})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toServiceResponse(s))
}

// ---- reservations

func (h *Handler) reserve(w http.ResponseWriter, r *http.Request) {
	var req reservationRequest
	if !h.decode(w, r, &req) {
		return
	}
	in, err := parseDate("check_in", req.CheckIn)
	if err != nil {
		h.fail(w, err)
		return
	}
	out, err := parseDate("check_out", req.CheckOut)
	if err != nil {
		h.fail(w, err)
		return
	}
	if k := r.Header.Get("Idempotency-Key"); k != "" { // the header wins over the body field
		req.RequestID = k
	}
	cmd := inbound.ReserveCommand{HotelID: req.HotelID, RoomTypeID: req.RoomTypeID, Start: in, End: out, Rooms: req.Rooms,
		Adults: req.Adults, Children: req.Children, SpecialRequests: req.SpecialRequests, RequestID: req.RequestID, PromoCode: req.PromoCode}
	for _, a := range req.Services {
		cmd.Addons = append(cmd.Addons, inbound.AddonRequest{ServiceID: a.ServiceID, Quantity: a.Quantity})
	}
	// When a room type is flooded, most requests are for rooms that are already gone. Answer those from
	// memory before authenticating: no user-service call, no database. (A retry of a request that already
	// succeeded is recognised by its request id and never turned away here.)
	if h.reservations.SoldOut(r.Context(), cmd) {
		h.fail(w, fmt.Errorf("%w: sold out for those nights", domain.ErrDatesUnavailable))
		return
	}
	r, err = h.authenticate(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	x, err := h.reservations.Reserve(r.Context(), mustPrincipal(r), cmd)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toReservationResponse(x))
}

func (h *Handler) listReservations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := map[string]domain.ReservationStatus{"": 0, "pending": domain.ReservationPending, "paid": domain.ReservationPaid,
		"canceled": domain.ReservationCanceled, "rejected": domain.ReservationRejected, "refunded": domain.ReservationRefunded}
	st, ok := status[q.Get("status")]
	if !ok {
		h.fail(w, domain.ErrInvalid)
		return
	}
	after, err := decodeIDCursor(q.Get("cursor"))
	if err != nil {
		h.fail(w, err)
		return
	}
	pg, err := h.reservations.List(r.Context(), mustPrincipal(r), inbound.ReservationListQuery{
		AsManager: q.Get("as") == "manager", HotelID: int64(atoi(q.Get("hotel_id"))), Status: st,
		AfterID: after, Limit: atoi(q.Get("limit")),
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(pg, toReservationResponse))
}

func (h *Handler) getReservation(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	x, err := h.reservations.Get(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toReservationResponse(x))
}

func (h *Handler) payReservation(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req payRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.reservations.Pay(r.Context(), mustPrincipal(r), id, inbound.PayCommand{
		Method: domain.PaymentMethod(strings.ToLower(req.Method)), Provider: req.Provider})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toReservationResponse(x))
}

func (h *Handler) cancelReservation(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	x, err := h.reservations.Cancel(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toReservationResponse(x))
}

func (h *Handler) cancellationQuote(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	q, err := h.reservations.CancellationQuote(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, cancellationQuoteResponse{Status: q.Status.Name(), HoursLeft: math.Round(q.HoursLeft*10) / 10,
		FeePercent: q.FeePercent, Fee: q.Fee, Refund: q.Refund, Cancellable: q.Cancellable, Reason: q.Reason, Policy: toTierDTOs(q.Policy)})
}

func (h *Handler) rejectReservation(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req rejectRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.reservations.Reject(r.Context(), mustPrincipal(r), id, req.Reason)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toReservationResponse(x))
}

// ---- reviews

func (h *Handler) listReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	q := r.URL.Query()
	after, err := decodeIDCursor(q.Get("cursor"))
	if err != nil {
		h.fail(w, err)
		return
	}
	pg, err := h.reviews.List(r.Context(), id, after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(pg, toReviewResponse))
}

func (h *Handler) createReview(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req reviewRequest
	if !h.decode(w, r, &req) {
		return
	}
	rv, err := h.reviews.Create(r.Context(), mustPrincipal(r), inbound.ReviewCommand{
		HotelID: id, ReservationID: req.ReservationID, Rating: req.Rating, Comment: req.Comment})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toReviewResponse(rv))
}

func toReviewResponse(rv domain.Review) reviewResponse {
	return reviewResponse{ID: rv.ID, HotelID: rv.HotelID, UserID: rv.UserID, ReservationID: rv.ReservationID,
		Rating: rv.Rating, Comment: rv.Comment, CreatedAt: rv.CreatedAt}
}

// ---- loyalty and wallet (the caller's own, held by payment-wallet-service)

func (h *Handler) loyaltyStatus(w http.ResponseWriter, r *http.Request) {
	s, err := h.loyalty.Status(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, loyaltyResponse{Points: s.Points, Tier: s.Tier, TierName: s.TierName})
}

func (h *Handler) getWallet(w http.ResponseWriter, r *http.Request) {
	x, err := h.wallets.Get(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toWalletResponse(x))
}

func (h *Handler) createWallet(w http.ResponseWriter, r *http.Request) {
	var req walletRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.wallets.Create(r.Context(), mustPrincipal(r), req.Currency)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toWalletResponse(x))
}

func (h *Handler) walletTransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.wallets.Transactions(r.Context(), mustPrincipal(r), atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]walletTxnResponse, len(list))
	for i, t := range list {
		out[i] = walletTxnResponse{ID: t.ID, Type: t.Type, Amount: t.Amount, Currency: t.Currency, Status: t.Status,
			ReferenceID: t.ReferenceID, Description: t.Description, CreatedAt: t.CreatedAt}
	}
	h.write(w, http.StatusOK, out)
}

func toWalletResponse(w domain.Wallet) walletResponse {
	return walletResponse{ID: w.ID, Balance: w.Balance, Currency: w.Currency, Status: w.Status}
}

// ---- helpers

func (h *Handler) decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		h.fail(w, errors.Join(domain.ErrInvalid, err))
		return false
	}
	return true
}

func (h *Handler) pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		h.fail(w, errors.Join(domain.ErrInvalid, errors.New("invalid id")))
		return 0, false
	}
	return id, true
}

func (h *Handler) write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("encode response", "err", err)
	}
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	status, msg := http.StatusInternalServerError, "internal error"
	switch {
	case errors.Is(err, domain.ErrInvalid):
		status, msg = http.StatusBadRequest, cleanMsg(err)
	case errors.Is(err, domain.ErrUnauthorized):
		status, msg = http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		status, msg = http.StatusForbidden, cleanMsg(err)
	case errors.Is(err, domain.ErrNotFound):
		status, msg = http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrDatesUnavailable):
		status, msg = http.StatusConflict, cleanMsg(err)
	case errors.Is(err, domain.ErrBusy):
		w.Header().Set("Retry-After", "1")
		status, msg = http.StatusTooManyRequests, "too many requests right now, try again in a moment"
	case errors.Is(err, domain.ErrInsufficientFunds):
		status, msg = http.StatusPaymentRequired, "insufficient wallet balance"
	case errors.Is(err, domain.ErrPaymentFailed):
		status, msg = http.StatusPaymentRequired, "payment was declined by the provider"
	case errors.Is(err, domain.ErrUpstreamUnavailable):
		status, msg = http.StatusServiceUnavailable, "dependent service unavailable"
	}
	if status == http.StatusInternalServerError || status == http.StatusServiceUnavailable {
		h.log.Error("request failed", "err", err)
	}
	h.write(w, status, map[string]string{"error": msg})
}

// cleanMsg drops the sentinel prefix ("invalid input: ", "conflict: ", "forbidden: ") from wrapped messages.
func cleanMsg(err error) string {
	m := err.Error()
	for _, p := range []string{"invalid input: ", "conflict: ", "forbidden: "} {
		if i := strings.Index(m, p); i >= 0 {
			return m[i+len(p):]
		}
	}
	return m
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// ---- notifications

func (h *Handler) listNotifications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	after, err := decodeIDCursor(q.Get("cursor"))
	if err != nil {
		h.fail(w, err)
		return
	}
	pg, err := h.notifications.List(r.Context(), mustPrincipal(r), q.Get("unread") == "true", after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(pg, toNotificationResponse))
}

func (h *Handler) unreadCount(w http.ResponseWriter, r *http.Request) {
	n, err := h.notifications.UnreadCount(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, map[string]int{"unread": n})
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.notifications.MarkRead(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	n, err := h.notifications.MarkAllRead(r.Context(), mustPrincipal(r))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, map[string]int{"marked": n})
}

// ---- promo codes and the manager's report

func (h *Handler) listPromotions(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	list, err := h.catalog.Promotions(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]promotionResponse, len(list))
	for i, p := range list {
		out[i] = toPromotionResponse(p)
	}
	h.write(w, http.StatusOK, out)
}

func toPromotionInput(req promotionRequest) (inbound.PromotionInput, error) {
	in := inbound.PromotionInput{Code: req.Code, Description: req.Description, PercentOff: req.PercentOff,
		MinNights: req.MinNights, MaxUses: req.MaxUses, Active: req.Active}
	if req.AmountOff != nil {
		in.AmountOff = req.AmountOff.String()
	}
	var err error
	if req.ValidFrom != "" {
		var t time.Time
		if t, err = parseDate("valid_from", req.ValidFrom); err != nil {
			return in, err
		}
		in.ValidFrom = &t
	}
	if req.ValidTo != "" {
		var t time.Time
		if t, err = parseDate("valid_to", req.ValidTo); err != nil {
			return in, err
		}
		in.ValidTo = &t
	}
	return in, nil
}

func (h *Handler) createPromotion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req promotionRequest
	if !h.decode(w, r, &req) {
		return
	}
	in, err := toPromotionInput(req)
	if err != nil {
		h.fail(w, err)
		return
	}
	p, err := h.catalog.CreatePromotion(r.Context(), mustPrincipal(r), id, in)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toPromotionResponse(p))
}

func (h *Handler) updatePromotion(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	pid, ok := h.pathID(w, r, "promo")
	if !ok {
		return
	}
	var req promotionRequest
	if !h.decode(w, r, &req) {
		return
	}
	in, err := toPromotionInput(req)
	if err != nil {
		h.fail(w, err)
		return
	}
	p, err := h.catalog.UpdatePromotion(r.Context(), mustPrincipal(r), id, pid, in)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPromotionResponse(p))
}

func pct(part, whole int) float64 {
	if whole == 0 {
		return 0
	}
	return math.Round(float64(part)/float64(whole)*1000) / 10
}

func (h *Handler) hotelReport(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	q := r.URL.Query()
	from, err := parseDate("from", q.Get("from"))
	if err != nil {
		h.fail(w, err)
		return
	}
	to, err := parseDate("to", q.Get("to"))
	if err != nil {
		h.fail(w, err)
		return
	}
	rep, err := h.catalog.Report(r.Context(), mustPrincipal(r), id, from, to)
	if err != nil {
		h.fail(w, err)
		return
	}
	var out reportResponse
	out.From, out.To = rep.From.Format(dateLayout), rep.To.Format(dateLayout)
	for _, d := range rep.Days {
		out.Occupancy.RoomNightsAvailable += d.Rooms
		out.Occupancy.RoomNightsSold += d.Reserved
		out.Occupancy.Days = append(out.Occupancy.Days, occupancyDay{Date: d.Date.Format(dateLayout), Rooms: d.Rooms, Reserved: d.Reserved, OccupancyPercent: pct(d.Reserved, d.Rooms)})
	}
	out.Occupancy.OccupancyPercent = pct(out.Occupancy.RoomNightsSold, out.Occupancy.RoomNightsAvailable)
	out.Revenue.Currency, out.Revenue.PaidCount, out.Revenue.PaidRevenue = rep.Currency, rep.PaidCount, rep.PaidRevenue
	out.Revenue.RefundedCount, out.Revenue.RefundedAmount, out.Revenue.FeesKept = rep.RefundedCount, rep.RefundedAmount, rep.FeesKept
	out.Revenue.NetRevenue, _ = domain.SumAmounts(rep.PaidRevenue, rep.FeesKept)
	out.Reservations = map[string]int{}
	for st, n := range rep.StatusCounts {
		out.Reservations[st.Name()] = n
	}
	out.NoShows = rep.NoShows
	h.write(w, http.StatusOK, out)
}

// ---- front desk and history

func (h *Handler) stayStep(step func(ctx context.Context, actor inbound.Principal, id int64) (domain.Reservation, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.pathID(w, r, "id")
		if !ok {
			return
		}
		x, err := step(r.Context(), mustPrincipal(r), id)
		if err != nil {
			h.fail(w, err)
			return
		}
		h.write(w, http.StatusOK, toReservationResponse(x))
	}
}

func (h *Handler) reservationHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	list, err := h.reservations.History(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]historyEntryResponse, len(list))
	for i, e := range list {
		out[i] = historyEntryResponse{At: e.At, By: e.By, Event: e.Event, From: e.From.Name(), To: e.To.Name(), Note: e.Note}
	}
	h.write(w, http.StatusOK, out)
}

// ---- wishlist

func (h *Handler) addFavourite(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.wishlist.Add(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeFavourite(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.wishlist.Remove(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listFavourites(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	after, err := decodeIDCursor(q.Get("cursor"))
	if err != nil {
		h.fail(w, err)
		return
	}
	pg, err := h.wishlist.List(r.Context(), mustPrincipal(r), after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(pg, toHotelResponse))
}

// ---- waiting list

func (h *Handler) joinWaitlist(w http.ResponseWriter, r *http.Request) {
	var req waitRequest
	if !h.decode(w, r, &req) {
		return
	}
	in, err := parseDate("check_in", req.CheckIn)
	if err != nil {
		h.fail(w, err)
		return
	}
	out, err := parseDate("check_out", req.CheckOut)
	if err != nil {
		h.fail(w, err)
		return
	}
	e, err := h.waitlist.Join(r.Context(), mustPrincipal(r), inbound.WaitCommand{HotelID: req.HotelID, RoomTypeID: req.RoomTypeID,
		Start: in, End: out, Rooms: req.Rooms, Adults: req.Adults, Children: req.Children})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toWaitResponse(e))
}

func (h *Handler) leaveWaitlist(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.waitlist.Leave(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listWaitlist(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	after, err := decodeIDCursor(q.Get("cursor"))
	if err != nil {
		h.fail(w, err)
		return
	}
	pg, err := h.waitlist.List(r.Context(), mustPrincipal(r), after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(pg, toWaitResponse))
}
