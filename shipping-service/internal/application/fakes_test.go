package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

var baseTime = time.Date(2026, 9, 28, 8, 0, 0, 0, time.UTC)

type memStore struct {
	mu         sync.Mutex
	shipments  map[int64]model.Shipment
	events     []model.ShipmentEvent
	warehouses map[int64]model.Warehouse
	jobs       map[int64]model.OutboxJob
	nextID     int64
}

func newMemStore() *memStore {
	return &memStore{shipments: map[int64]model.Shipment{}, warehouses: map[int64]model.Warehouse{}, jobs: map[int64]model.OutboxJob{}}
}

func (m *memStore) id() int64 { m.nextID++; return m.nextID }

type memShipments struct{ m *memStore }

func (r memShipments) Create(_ context.Context, s *model.Shipment) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, e := range r.m.shipments {
		if e.ClientService == s.ClientService && e.ClientOrderCode == s.ClientOrderCode {
			return port.ErrConflict
		}
	}
	s.ID = r.m.id()
	r.m.shipments[s.ID] = *s
	return nil
}

func (r memShipments) Save(_ context.Context, s *model.Shipment) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	cur := r.m.shipments[s.ID]
	s.PlacingUntil = cur.PlacingUntil
	r.m.shipments[s.ID] = *s
	return nil
}

func (r memShipments) find(match func(model.Shipment) bool) (*model.Shipment, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, s := range r.m.shipments {
		if match(s) {
			return &s, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r memShipments) GetByID(_ context.Context, id int64) (*model.Shipment, error) {
	return r.find(func(s model.Shipment) bool { return s.ID == id })
}

func (r memShipments) GetByIDForUpdate(ctx context.Context, id int64) (*model.Shipment, error) {
	return r.GetByID(ctx, id)
}

func (r memShipments) GetByClientOrderCode(_ context.Context, client, code string) (*model.Shipment, error) {
	return r.find(func(s model.Shipment) bool { return s.ClientService == client && s.ClientOrderCode == code })
}

func (r memShipments) GetByCode(_ context.Context, code string) (*model.Shipment, error) {
	return r.find(func(s model.Shipment) bool { return s.Code == code })
}

func (r memShipments) GetByCarrierOrderCode(_ context.Context, code string) (*model.Shipment, error) {
	return r.find(func(s model.Shipment) bool { return s.CarrierOrderCode == code })
}

func (r memShipments) List(_ context.Context, client string, f port.ShipmentFilter) ([]model.Shipment, int64, error) {
	var out []model.Shipment
	for _, s := range r.m.shipments {
		if s.ClientService == client && (f.Status == "" || s.Status == f.Status) {
			out = append(out, s)
		}
	}
	return out, int64(len(out)), nil
}

func (r memShipments) AcquirePlacement(_ context.Context, id int64, now, until time.Time) (bool, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	s := r.m.shipments[id]
	if s.PlacingUntil != nil && !s.PlacingUntil.Before(now) {
		return false, nil
	}
	s.PlacingUntil = &until
	r.m.shipments[id] = s
	return true, nil
}

func (r memShipments) ReleasePlacement(_ context.Context, id int64) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	s := r.m.shipments[id]
	s.PlacingUntil = nil
	r.m.shipments[id] = s
	return nil
}

type memEvents struct{ m *memStore }

func (r memEvents) Create(_ context.Context, e *model.ShipmentEvent) (bool, error) {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	for _, x := range r.m.events {
		if x.ShipmentID == e.ShipmentID && x.Source == e.Source && x.ToStatus == e.ToStatus &&
			x.CarrierStatus == e.CarrierStatus && x.OccurredAt.Equal(e.OccurredAt) {
			return false, nil
		}
	}
	e.ID = r.m.id()
	r.m.events = append(r.m.events, *e)
	return true, nil
}

func (r memEvents) ListByShipment(_ context.Context, id int64) ([]model.ShipmentEvent, error) {
	var out []model.ShipmentEvent
	for _, e := range r.m.events {
		if e.ShipmentID == id {
			out = append(out, e)
		}
	}
	return out, nil
}

type memWarehouses struct{ m *memStore }

func (r memWarehouses) Create(_ context.Context, w *model.Warehouse) error {
	w.ID = r.m.id()
	r.m.warehouses[w.ID] = *w
	return nil
}

func (r memWarehouses) Save(_ context.Context, w *model.Warehouse) error {
	r.m.warehouses[w.ID] = *w
	return nil
}

func (r memWarehouses) GetByID(_ context.Context, id int64) (*model.Warehouse, error) {
	w, ok := r.m.warehouses[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &w, nil
}

func (r memWarehouses) List(context.Context) ([]model.Warehouse, error) {
	return r.ListActive(nil, nil)
}

func (r memWarehouses) ListActive(_ context.Context, ids []int64) ([]model.Warehouse, error) {
	var out []model.Warehouse
	for _, w := range r.m.warehouses {
		if !w.Active {
			continue
		}
		if len(ids) > 0 {
			found := false
			for _, id := range ids {
				found = found || id == w.ID
			}
			if !found {
				continue
			}
		}
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

type memOutbox struct{ m *memStore }

func (r memOutbox) Enqueue(_ context.Context, jobs []model.OutboxJob) error {
	for _, j := range jobs {
		j.ID = r.m.id()
		r.m.jobs[j.ID] = j
	}
	return nil
}

func (r memOutbox) ClaimDue(_ context.Context, now time.Time, limit int, lease time.Duration) ([]model.OutboxJob, error) {
	var out []model.OutboxJob
	for id, j := range r.m.jobs {
		if j.State == model.JobPending && !j.NextAttemptAt.After(now) && len(out) < limit {
			j.Attempts++
			j.NextAttemptAt = now.Add(lease)
			r.m.jobs[id] = j
			out = append(out, j)
		}
	}
	sort.Slice(out, func(i, k int) bool { return out[i].ID < out[k].ID })
	return out, nil
}

func (r memOutbox) set(id int64, fn func(*model.OutboxJob)) error {
	j := r.m.jobs[id]
	fn(&j)
	r.m.jobs[id] = j
	return nil
}

func (r memOutbox) MarkDone(_ context.Context, id int64) error {
	return r.set(id, func(j *model.OutboxJob) { j.State = model.JobDone })
}

func (r memOutbox) MarkRetry(_ context.Context, id int64, next time.Time, e string) error {
	return r.set(id, func(j *model.OutboxJob) { j.NextAttemptAt, j.LastError = next, e })
}

func (r memOutbox) MarkFailed(_ context.Context, id int64, e string) error {
	return r.set(id, func(j *model.OutboxJob) { j.State, j.LastError = model.JobFailed, e })
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type seqCodes struct{ n int }

func (c *seqCodes) NewShipmentCode() string { c.n++; return fmt.Sprintf("SHPTEST%03d", c.n) }

type fakeDirectory struct{}

func (fakeDirectory) Provinces(context.Context) ([]port.Province, error) {
	return []port.Province{{ID: 201, Name: "Hà Nội"}, {ID: 202, Name: "Hồ Chí Minh"}}, nil
}

func (fakeDirectory) Districts(_ context.Context, provinceID int) ([]port.District, error) {
	return map[int][]port.District{
		201: {{ID: 1485, ProvinceID: 201, Name: "Ba Đình"}},
		202: {{ID: 1442, ProvinceID: 202, Name: "Quận 1"}, {ID: 1454, ProvinceID: 202, Name: "Quận 3"}},
	}[provinceID], nil
}

func (fakeDirectory) Wards(_ context.Context, districtID int) ([]port.Ward, error) {
	return map[int][]port.Ward{
		1485: {{Code: "1A0101", DistrictID: 1485}},
		1442: {{Code: "20109", DistrictID: 1442}},
		1454: {{Code: "21211", DistrictID: 1454}},
	}[districtID], nil
}

type quoteKey struct {
	shop    int64
	service int
}

type fakeCarrier struct {
	mu          sync.Mutex
	services    map[int64][]port.CarrierService
	quotes      map[quoteKey]port.CarrierQuote
	createErr   error
	created     []port.CarrierOrderRequest
	existing    map[string]*port.CarrierOrder
	remote      map[string]*port.CarrierOrder
	cancelErr   error
	cancelled   []string
	webhookAuth string
}

func newFakeCarrier() *fakeCarrier {
	return &fakeCarrier{
		services: map[int64][]port.CarrierService{},
		quotes:   map[quoteKey]port.CarrierQuote{},
		existing: map[string]*port.CarrierOrder{},
		remote:   map[string]*port.CarrierOrder{},
	}
}

func (c *fakeCarrier) AvailableServices(_ context.Context, shopID int64, _, _ int) ([]port.CarrierService, error) {
	return c.services[shopID], nil
}

func (c *fakeCarrier) Quote(_ context.Context, req port.CarrierQuoteRequest) (*port.CarrierQuote, error) {
	q, ok := c.quotes[quoteKey{req.ShopID, req.ServiceID}]
	if !ok {
		return nil, fmt.Errorf("%w: route not served", port.ErrCarrierRejected)
	}
	return &q, nil
}

func (c *fakeCarrier) CreateOrder(_ context.Context, req port.CarrierOrderRequest) (*port.CarrierOrder, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.created = append(c.created, req)
	if c.createErr != nil {
		return nil, c.createErr
	}
	eta := baseTime.Add(48 * time.Hour)
	o := &port.CarrierOrder{OrderCode: "GHN" + req.ClientOrderCode, Status: "ready_to_pick", Fee: 31000, ExpectedDeliveryAt: &eta}
	c.existing[req.ClientOrderCode] = o
	return o, nil
}

func (c *fakeCarrier) FindByClientOrderCode(_ context.Context, _ int64, code string) (*port.CarrierOrder, error) {
	if o, ok := c.existing[code]; ok {
		return o, nil
	}
	return nil, port.ErrNotFound
}

func (c *fakeCarrier) GetOrder(_ context.Context, _ int64, code string) (*port.CarrierOrder, error) {
	if o, ok := c.remote[code]; ok {
		return o, nil
	}
	return nil, errors.New("unexpected GetOrder")
}

func (c *fakeCarrier) CancelOrder(_ context.Context, _ int64, code string) error {
	c.cancelled = append(c.cancelled, code)
	return c.cancelErr
}

func (c *fakeCarrier) VerifyWebhookToken(token string) bool {
	return c.webhookAuth != "" && token == c.webhookAuth
}

type recordingNotifier struct {
	sent []model.Notification
	err  error
}

func (n *recordingNotifier) Notify(_ context.Context, x model.Notification) error {
	if n.err != nil {
		return n.err
	}
	n.sent = append(n.sent, x)
	return nil
}

type recordingCallbacks struct {
	sent []model.CallbackEvent
}

func (c *recordingCallbacks) Allowed(raw string) error {
	if raw == "http://evil.example" {
		return errors.New("host not allowed")
	}
	return nil
}

func (c *recordingCallbacks) Send(_ context.Context, _ string, e model.CallbackEvent) error {
	c.sent = append(c.sent, e)
	return nil
}

type fixture struct {
	store     *memStore
	carrier   *fakeCarrier
	notifier  *recordingNotifier
	callbacks *recordingCallbacks
	clock     *fakeClock
	shipments port.ShipmentUsecase
	webhooks  port.WebhookUsecase
	outbox    port.OutboxUsecase
	hn, hcm   model.Warehouse
}

func newFixture() *fixture {
	m := newMemStore()
	f := &fixture{
		store:     m,
		carrier:   newFakeCarrier(),
		notifier:  &recordingNotifier{},
		callbacks: &recordingCallbacks{},
		clock:     &fakeClock{now: baseTime},
	}
	f.carrier.webhookAuth = "hook-secret"
	settings := DefaultSettings()
	ships, events, whs, box := memShipments{m}, memEvents{m}, memWarehouses{m}, memOutbox{m}

	f.hn = model.Warehouse{Code: "HN", DistrictID: 1485, WardCode: "1A0101", CarrierShopID: 11, Active: true}
	f.hcm = model.Warehouse{Code: "HCM", DistrictID: 1454, WardCode: "21211", CarrierShopID: 22, Active: true}
	_ = whs.Create(context.Background(), &f.hn)
	_ = whs.Create(context.Background(), &f.hcm)

	f.carrier.services[11] = []port.CarrierService{{ID: 53320, TypeID: 2, Name: "Chuẩn"}}
	f.carrier.services[22] = []port.CarrierService{{ID: 53321, TypeID: 2, Name: "Chuẩn"}, {ID: 53322, TypeID: 5, Name: "Hàng nặng"}}
	f.carrier.quotes[quoteKey{11, 53320}] = port.CarrierQuote{Fee: 36000, ExpectedDeliveryAt: baseTime.Add(72 * time.Hour)}
	f.carrier.quotes[quoteKey{22, 53321}] = port.CarrierQuote{Fee: 22000, ExpectedDeliveryAt: baseTime.Add(24 * time.Hour)}

	f.shipments = NewShipmentService(ships, events, whs, box, f.carrier, fakeDirectory{}, f.callbacks, &seqCodes{}, memTx{}, f.clock, settings)
	f.webhooks = NewWebhookService(ships, events, box, f.carrier, memTx{}, f.clock)
	f.outbox = NewOutboxService(box, f.notifier, f.callbacks, f.clock, settings)
	return f
}

func validInput(code string) port.CreateShipmentInput {
	return port.CreateShipmentInput{
		ClientOrderCode: code,
		Recipient:       model.Address{Name: "Nguyễn Văn A", Phone: "+84 912 345 678", Street: "12 Lê Lợi", WardCode: "20109", DistrictID: 1442, ProvinceID: 202},
		Parcel: model.Parcel{WeightGram: 800, LengthCm: 20, WidthCm: 15, HeightCm: 10, Content: "Áo thun",
			Items: []model.Item{{Name: "Áo thun", Quantity: 2, WeightGram: 400, Price: 150000}}},
		CODAmount:   300000,
		Customer:    model.Customer{UserID: 42, Email: "a@example.vn"},
		CallbackURL: "http://order-service-svc.default.svc.cluster.local/hooks/shipping",
	}
}
