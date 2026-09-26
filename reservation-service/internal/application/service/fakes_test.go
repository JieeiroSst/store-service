package service

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

var (
	adminP = inbound.Principal{UserID: 1, Admin: true}
	ownerP = inbound.Principal{UserID: 50} // manages hotel 1
	guestP = inbound.Principal{UserID: 7, Email: "g@x.com"}
	otherP = inbound.Principal{UserID: 60}
)

// ---- hotels

// fakeHotels is a one-hotel repository that records what the services store.
type fakeHotels struct {
	outbound.HotelRepository
	hotel      domain.Hotel
	saved      *domain.Hotel
	review     *reviewCall
	listed     *domain.HotelFilter
	list       []domain.Hotel
	services   []domain.HotelService
	inv        *outbound.SetInventoryParams
	rtypes     map[int64]domain.RoomType
	created    []domain.RoomType
	promos     []domain.Promotion
	report     *domain.HotelReport
	reportFrom time.Time
}

type reviewCall struct {
	status domain.HotelStatus
	note   string
}

func (f *fakeHotels) Get(context.Context, int64) (domain.Hotel, error) { return f.hotel, nil }
func (f *fakeHotels) Create(_ context.Context, h domain.Hotel) (domain.Hotel, error) {
	f.saved = &h
	return h, nil
}
func (f *fakeHotels) Update(_ context.Context, h domain.Hotel) (domain.Hotel, error) {
	f.saved = &h
	return h, nil
}
func (f *fakeHotels) SetStatus(_ context.Context, _ int64, s domain.HotelStatus) error {
	f.saved = &domain.Hotel{Status: s}
	return nil
}
func (f *fakeHotels) SetReview(_ context.Context, _ int64, s domain.HotelStatus, note string) error {
	f.review = &reviewCall{s, note}
	return nil
}
func (f *fakeHotels) List(_ context.Context, fl domain.HotelFilter) ([]domain.Hotel, error) {
	f.listed = &fl
	if f.list == nil {
		return nil, nil
	}
	if fl.Limit > 0 && fl.Limit < len(f.list) {
		return append([]domain.Hotel(nil), f.list[:fl.Limit]...), nil
	}
	return append([]domain.Hotel(nil), f.list...), nil
}
func (f *fakeHotels) Services(context.Context, int64, bool) ([]domain.HotelService, error) {
	return f.services, nil
}
func (f *fakeHotels) Promotions(context.Context, int64) ([]domain.Promotion, error) {
	return f.promos, nil
}
func (f *fakeHotels) PromotionByCode(_ context.Context, _ int64, code string) (domain.Promotion, error) {
	for _, p := range f.promos {
		if strings.EqualFold(p.Code, code) {
			return p, nil
		}
	}
	return domain.Promotion{}, domain.ErrNotFound
}
func (f *fakeHotels) SavePromotion(_ context.Context, p domain.Promotion) (domain.Promotion, error) {
	f.promos = append(f.promos, p)
	return p, nil
}
func (f *fakeHotels) Report(_ context.Context, _ int64, from, _ time.Time) (domain.HotelReport, error) {
	f.reportFrom = from
	if f.report != nil {
		return *f.report, nil
	}
	return domain.HotelReport{}, nil
}
func (f *fakeHotels) SaveService(_ context.Context, s domain.HotelService) (domain.HotelService, error) {
	return s, nil
}
func (f *fakeHotels) GetRoomType(_ context.Context, id int64) (domain.RoomType, error) {
	if t, ok := f.rtypes[id]; ok {
		return t, nil
	}
	return domain.RoomType{}, domain.ErrNotFound
}
func (f *fakeHotels) CreateRoomType(_ context.Context, t domain.RoomType) (domain.RoomType, error) {
	f.created = append(f.created, t)
	return t, nil
}
func (f *fakeHotels) SetInventory(_ context.Context, p outbound.SetInventoryParams) error {
	f.inv = &p
	return nil
}

// ---- wallets, gateway, loyalty

type fakeWallets struct {
	outbound.WalletGateway
	wallet      domain.Wallet
	noWallet    bool
	transfers   map[string]string // reference -> transfer id
	to          []string          // destination of each transfer
	reversed    []string
	transferErr error
	missing     map[string]bool
	status      map[string]string
	ownerID     int64   // user the wallet returned by Get belongs to
	feeErr      error   // returned for the cancellation-fee transfer only
	fees        []int64 // amounts of cancellation-fee transfers
}

func (f *fakeWallets) Get(_ context.Context, id string) (domain.Wallet, error) {
	if f.missing[id] {
		return domain.Wallet{}, domain.ErrNotFound
	}
	if st, ok := f.status[id]; ok {
		return domain.Wallet{ID: id, Status: st, UserID: f.ownerID}, nil
	}
	return domain.Wallet{ID: id, Status: "ACTIVE", UserID: f.ownerID}, nil
}
func (f *fakeWallets) GetByUser(context.Context, int64) (domain.Wallet, error) {
	if f.noWallet {
		return domain.Wallet{}, domain.ErrNotFound
	}
	return f.wallet, nil
}
func (f *fakeWallets) Transfer(_ context.Context, p outbound.TransferParams) (string, error) {
	if f.transferErr != nil {
		return "", f.transferErr
	}
	if f.feeErr != nil && strings.HasSuffix(p.ReferenceID, "-fee") {
		return "", f.feeErr
	}
	if strings.HasSuffix(p.ReferenceID, "-fee") {
		f.fees = append(f.fees, p.Amount)
	}
	f.to = append(f.to, p.ToWalletID)
	if f.transfers == nil {
		f.transfers = map[string]string{}
	}
	if id, ok := f.transfers[p.ReferenceID]; ok {
		return id, nil // idempotent
	}
	id := "t-" + p.ReferenceID
	f.transfers[p.ReferenceID] = id
	return id, nil
}
func (f *fakeWallets) ReverseTransfer(_ context.Context, id, _ string) error {
	f.reversed = append(f.reversed, id)
	return nil
}

type fakeGateway struct {
	outbound.PaymentGateway
	status   domain.GatewayStatus
	created  []outbound.CreateGatewayPayment
	refunded []int64
	partial  map[int64]int64 // payment id -> amount refunded of a partial refund
}

func (f *fakeGateway) Create(_ context.Context, in outbound.CreateGatewayPayment) (domain.GatewayPayment, error) {
	f.created = append(f.created, in)
	return domain.GatewayPayment{ID: 77, Status: f.status, Amount: in.Amount}, nil
}
func (f *fakeGateway) Get(context.Context, int64) (domain.GatewayPayment, error) {
	return domain.GatewayPayment{ID: 77, Status: f.status}, nil
}
func (f *fakeGateway) RefundPartial(_ context.Context, id, amount int64) error {
	if f.partial == nil {
		f.partial = map[int64]int64{}
	}
	f.partial[id] = amount
	return nil
}
func (f *fakeGateway) Refund(_ context.Context, id int64) error {
	f.refunded = append(f.refunded, id)
	return nil
}

type fakeLoyalty struct {
	earned map[string]int64
	tiers  map[int64]int
	calls  int
}

func (f *fakeLoyalty) Earn(_ context.Context, _ int64, key string, amount int64) error {
	if f.earned == nil {
		f.earned = map[string]int64{}
	}
	f.earned[key] = amount
	return nil
}
func (f *fakeLoyalty) Status(_ context.Context, id int64) (outbound.Member, error) {
	f.calls++
	return outbound.Member{Tier: f.tiers[id]}, nil
}

// ---- reservations: one reservation in memory with the same state rules as the SQL

type memReservations struct {
	outbound.ReservationRepository
	x           domain.Reservation
	reserved    *outbound.ReserveParams
	failPaid    error // returned once by MarkPaid to simulate a lost race
	transitions []domain.ReservationStatus
	byRequest   *domain.Reservation  // what FindByRequestID knows
	left        *int                 // what RoomsLeft answers (nil: plenty)
	due         []domain.Reservation // what DueForReminder returns
	reminded    map[int64]bool
	stayActor   int64
	events      []domain.ReservationEvent
	probes      int32 // RoomsLeft calls
	probeDelay  time.Duration
	reserveErr  error // returned by Reserve
}

func (m *memReservations) Reserve(_ context.Context, p outbound.ReserveParams) (domain.Reservation, error) {
	m.reserved = &p
	if m.byRequest != nil && m.byRequest.RequestID == p.RequestID && m.byRequest.GuestID != p.GuestID {
		return domain.Reservation{}, domain.ErrConflict // the request id belongs to another guest
	}
	if m.reserveErr != nil {
		return domain.Reservation{}, m.reserveErr
	}
	return domain.Reservation{ID: 1, GuestID: p.GuestID, Status: domain.ReservationPending}, nil
}
func (m *memReservations) Get(context.Context, int64) (domain.Reservation, error) { return m.x, nil }
func (m *memReservations) SetStay(_ context.Context, _ int64, from, to domain.StayStatus, actor int64) (domain.Reservation, error) {
	if m.x.Status != domain.ReservationPaid || m.x.Stay != from {
		return domain.Reservation{}, domain.ErrConflict
	}
	m.x.Stay = to
	m.stayActor = actor
	return m.x, nil
}
func (m *memReservations) Events(context.Context, int64) ([]domain.ReservationEvent, error) {
	return m.events, nil
}
func (m *memReservations) DueForReminder(context.Context, time.Time, time.Time, int) ([]domain.Reservation, error) {
	var out []domain.Reservation
	for _, x := range m.due {
		if !m.reminded[x.ID] {
			out = append(out, x)
		}
	}
	return out, nil
}
func (m *memReservations) MarkReminded(_ context.Context, id int64) (bool, error) {
	if m.reminded == nil {
		m.reminded = map[int64]bool{}
	}
	if m.reminded[id] {
		return false, nil
	}
	m.reminded[id] = true
	return true, nil
}
func (m *memReservations) RoomsLeft(context.Context, int64, int64, time.Time, time.Time) (int, error) {
	atomic.AddInt32(&m.probes, 1)
	time.Sleep(m.probeDelay)
	if m.left != nil {
		return *m.left, nil
	}
	return 100, nil
}
func (m *memReservations) RequestIDExists(_ context.Context, req string) (bool, error) {
	return m.byRequest != nil && m.byRequest.RequestID == req, nil
}
func (m *memReservations) FindByRequestID(_ context.Context, guest int64, req string) (domain.Reservation, bool, error) {
	if m.byRequest != nil && m.byRequest.RequestID == req && m.byRequest.GuestID == guest {
		return *m.byRequest, true, nil
	}
	return domain.Reservation{}, false, nil
}
func (m *memReservations) SetPaymentAttempt(_ context.Context, _ int64, method domain.PaymentMethod, ref string) (domain.Reservation, error) {
	m.x.PaymentMethod, m.x.PaymentRef, m.x.Version = method, ref, m.x.Version+1
	return m.x, nil
}
func (m *memReservations) MarkPaid(_ context.Context, _ int64, method domain.PaymentMethod, ref string) (domain.Reservation, error) {
	if m.failPaid != nil {
		err := m.failPaid
		m.failPaid = nil
		return domain.Reservation{}, err
	}
	if m.x.Status != domain.ReservationPending {
		if m.x.Status == domain.ReservationPaid && m.x.PaymentRef == ref {
			return m.x, nil
		}
		return domain.Reservation{}, domain.ErrConflict
	}
	m.x.Status, m.x.PaymentMethod, m.x.PaymentRef = domain.ReservationPaid, method, ref
	return m.x, nil
}
func (m *memReservations) Transition(_ context.Context, p outbound.TransitionParams) (domain.Reservation, error) {
	for _, f := range p.From {
		if m.x.Status == f {
			m.x.Status, m.x.StatusNote = p.To, p.Note
			m.x.FeeAmount, m.x.RefundAmount = p.Fee, p.Refund
			m.transitions = append(m.transitions, p.To)
			return m.x, nil
		}
	}
	return domain.Reservation{}, domain.ErrConflict
}
func (m *memReservations) ListExpired(context.Context, time.Time, int) ([]domain.Reservation, error) {
	if m.x.Status == domain.ReservationPending && m.x.ExpiresAt != nil && m.x.ExpiresAt.Before(time.Now()) {
		return []domain.Reservation{m.x}, nil
	}
	return nil, nil
}
