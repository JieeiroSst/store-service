package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

const maxBodyBytes = 1 << 20

type Handler struct {
	auth      inbound.AuthUseCase
	homestays inbound.HomestayUseCase
	bookings  inbound.BookingUseCase
	wallets   inbound.WalletUseCase
	leases    inbound.LeaseUseCase
	reviews   inbound.ReviewUseCase
	wishlist  inbound.WishlistUseCase
	loyalty   inbound.LoyaltyUseCase
	log       *slog.Logger
}

func NewHandler(a inbound.AuthUseCase, h inbound.HomestayUseCase, b inbound.BookingUseCase, w inbound.WalletUseCase,
	l inbound.LeaseUseCase, rv inbound.ReviewUseCase, wl inbound.WishlistUseCase, lo inbound.LoyaltyUseCase, log *slog.Logger) *Handler {
	return &Handler{auth: a, homestays: h, bookings: b, wallets: w, leases: l, reviews: rv, wishlist: wl, loyalty: lo, log: log}
}

// ---- homestays (public reads)

func (h *Handler) listHomestays(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := domain.HomestayFilter{
		ProvinceID: atoi(q.Get("province_id")), DistrictID: atoi(q.Get("district_id")),
		WardID: atoi(q.Get("ward_id")), Type: atoi(q.Get("type")), MinGuests: atoi(q.Get("guests")),
		Query: strings.TrimSpace(q.Get("q")), Model: domain.RentalModel(q.Get("model")),
		MinPrice: q.Get("min_price"), MaxPrice: q.Get("max_price"), Sort: q.Get("sort"),
		Limit: atoi(q.Get("limit")),
	}
	if c := q.Get("cursor"); c != "" {
		var err error
		if f.After, err = decodeCursor(c); err != nil {
			h.fail(w, err)
			return
		}
	}
	for _, v := range strings.Split(q.Get("amenity_ids"), ",") {
		if id := atoi(strings.TrimSpace(v)); id > 0 {
			f.AmenityIDs = append(f.AmenityIDs, id)
		}
	}
	if q.Get("checkin") != "" || q.Get("checkout") != "" {
		var err error
		if f.CheckIn, err = parseDate("checkin", q.Get("checkin")); err != nil {
			h.fail(w, err)
			return
		}
		if f.CheckOut, err = parseDate("checkout", q.Get("checkout")); err != nil {
			h.fail(w, err)
			return
		}
	}
	pg, err := h.homestays.List(r.Context(), f)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]homestayResponse, len(pg.Items))
	for i, x := range pg.Items {
		out[i] = toHomestayResponse(x)
	}
	h.write(w, http.StatusOK, homestayPageResponse{Items: out, HasMore: pg.Next != nil, NextCursor: encodeCursor(pg.Next)})
}

func (h *Handler) getHomestay(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var viewer *inbound.Principal
	if p, ok := principalFrom(r.Context()); ok {
		viewer = &p
	}
	x, err := h.homestays.View(r.Context(), viewer, id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHomestayResponse(x))
}

func (h *Handler) availability(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
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
	slots, err := h.homestays.Availability(r.Context(), id, from, to)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]slotResponse, len(slots))
	for i, s := range slots {
		out[i] = slotResponse{Date: s.Date.Format(dateLayout), Price: s.Price, Status: int16(s.Status)}
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) listAmenities(w http.ResponseWriter, r *http.Request) {
	list, err := h.homestays.ListAmenities(r.Context())
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]amenityResponse, len(list))
	for i, a := range list {
		out[i] = amenityResponse{ID: a.ID, Name: a.Name, Icon: a.Icon}
	}
	h.write(w, http.StatusOK, out)
}

// ---- homestays (admin)

func (h *Handler) createHomestay(w http.ResponseWriter, r *http.Request) {
	var req homestayRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.homestays.Create(r.Context(), mustPrincipal(r), toHomestayInput(req))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toHomestayResponse(x))
}

func (h *Handler) updateHomestay(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var req homestayRequest
	if !h.decode(w, r, &req) {
		return
	}
	x, err := h.homestays.Update(r.Context(), mustPrincipal(r), id, toHomestayInput(req))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toHomestayResponse(x))
}

func (h *Handler) deactivateHomestay(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	if err := h.homestays.Deactivate(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setAvailability(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var req availabilityRequest
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
	err = h.homestays.SetAvailability(r.Context(), mustPrincipal(r), inbound.SetAvailabilityCommand{
		HomestayID: id, From: from, To: to, Price: req.Price.String(), Status: domain.AvailabilityStatus(req.Status),
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createAmenity(w http.ResponseWriter, r *http.Request) {
	var req amenityRequest
	if !h.decode(w, r, &req) {
		return
	}
	a, err := h.homestays.CreateAmenity(r.Context(), mustPrincipal(r), req.Name, req.Icon)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, amenityResponse{ID: a.ID, Name: a.Name, Icon: a.Icon})
}

// ---- bookings

func (h *Handler) book(w http.ResponseWriter, r *http.Request) {
	var req bookingRequest
	if !h.decode(w, r, &req) {
		return
	}
	in, err := parseDate("checkin_date", req.CheckIn)
	if err != nil {
		h.fail(w, err)
		return
	}
	out, err := parseDate("checkout_date", req.CheckOut)
	if err != nil {
		h.fail(w, err)
		return
	}
	// Idempotency-Key header wins over the body field.
	if k := r.Header.Get("Idempotency-Key"); k != "" {
		req.RequestID = k
	}
	b, err := h.bookings.Book(r.Context(), mustPrincipal(r), inbound.BookCommand{
		HomestayID: req.HomestayID, CheckIn: in, CheckOut: out, Guests: req.Guests,
		Currency: req.Currency, Note: req.Note, RequestID: req.RequestID,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toBookingResponse(b))
}

func (h *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	b, err := h.bookings.Get(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toBookingResponse(b))
}

func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	b, err := h.bookings.Cancel(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toBookingResponse(b))
}

func (h *Handler) listBookings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.bookings.List(r.Context(), mustPrincipal(r), int64(atoi(q.Get("user_id"))), q.Get("as") == "host", atoi(q.Get("limit")), atoi(q.Get("offset")))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]bookingResponse, len(list))
	for i, b := range list {
		out[i] = toBookingResponse(b)
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) payBooking(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var req payRequest
	if !h.decode(w, r, &req) {
		return
	}
	b, err := h.bookings.Pay(r.Context(), mustPrincipal(r), id, inbound.PayCommand{
		Method: domain.PaymentMethod(strings.ToLower(req.Method)), Provider: req.Provider,
	})
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toBookingResponse(b))
}

// ---- wallet (the caller's own, held by payment-wallet-service)

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

func toHomestayInput(r homestayRequest) inbound.HomestayInput {
	return inbound.HomestayInput{
		Name: r.Name, Description: r.Description, Type: r.Type, Status: domain.HomestayStatus(r.Status),
		PhoneNumber: r.PhoneNumber, Address: r.Address, WardID: r.WardID, DistrictID: r.DistrictID,
		ProvinceID: r.ProvinceID, Images: r.Images, Guests: r.Guests, Bedrooms: r.Bedrooms,
		Bathrooms: r.Bathrooms, AmenityIDs: r.AmenityIDs, WalletID: r.WalletID,
	}
}

func (h *Handler) decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		h.fail(w, errors.Join(domain.ErrInvalid, err))
		return false
	}
	return true
}

func (h *Handler) pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
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
		status, msg = http.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		status, msg = http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrDatesUnavailable):
		status, msg = http.StatusConflict, cleanMsg(err)
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

// cleanMsg drops the sentinel prefix ("invalid input: ", "conflict: ") from wrapped messages.
func cleanMsg(err error) string {
	m := err.Error()
	for _, p := range []string{"invalid input: ", "conflict: "} {
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
