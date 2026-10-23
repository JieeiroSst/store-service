package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type memBaskets struct{ byID map[int]*model.Basket }

func (m *memBaskets) Create(_ context.Context, b *model.Basket) error {
	b.ID = len(m.byID) + 1
	cp := *b
	m.byID[b.ID] = &cp
	return nil
}
func (m *memBaskets) GetByID(_ context.Context, id int) (*model.Basket, error) {
	b, ok := m.byID[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	cp := *b
	return &cp, nil
}
func (m *memBaskets) Update(_ context.Context, b *model.Basket) error {
	cp := *b
	m.byID[b.ID] = &cp
	return nil
}
func (m *memBaskets) Delete(context.Context, int) error            { return nil }
func (m *memBaskets) List(context.Context) ([]model.Basket, error) { return nil, nil }
func (m *memBaskets) ListByUser(context.Context, int) ([]model.Basket, error) {
	return nil, nil
}
func (m *memBaskets) ListAwaitingFulfillment(_ context.Context, limit int) ([]model.Basket, error) {
	var out []model.Basket
	for i := 1; i <= len(m.byID); i++ {
		if b, ok := m.byID[i]; ok && b.Status == model.BasketStatusSubmitted && b.FulfillmentID == "" && len(out) < limit {
			out = append(out, *b)
		}
	}
	return out, nil
}

type fakeFulfillment struct {
	orders []port.PaidOrder
	err    error
}

func (f *fakeFulfillment) StartPaidOrder(_ context.Context, o port.PaidOrder) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.orders = append(f.orders, o)
	return fmt.Sprintf("paid-order-basket-%d", o.BasketID), nil
}

type memLines struct{ byID map[int]*model.BasketLine }

func (m *memLines) Create(_ context.Context, l *model.BasketLine) error {
	l.ID = len(m.byID) + 1
	cp := *l
	m.byID[l.ID] = &cp
	return nil
}
func (m *memLines) GetByID(_ context.Context, id int) (*model.BasketLine, error) {
	l, ok := m.byID[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	cp := *l
	return &cp, nil
}
func (m *memLines) Update(_ context.Context, l *model.BasketLine) error {
	cp := *l
	m.byID[l.ID] = &cp
	return nil
}
func (m *memLines) Delete(_ context.Context, id int) error           { delete(m.byID, id); return nil }
func (m *memLines) List(context.Context) ([]model.BasketLine, error) { return nil, nil }
func (m *memLines) ListByBasket(_ context.Context, basketID int) ([]model.BasketLine, error) {
	var out []model.BasketLine
	for i := 1; i <= len(m.byID)+10; i++ {
		if l, ok := m.byID[i]; ok && l.BasketID == basketID {
			out = append(out, *l)
		}
	}
	return out, nil
}

type fakeCatalogue struct {
	records   map[int]port.StockRecord
	allocated map[int]int
	failOn    int
}

func newCatalogue() *fakeCatalogue {
	three, one := 3, 1
	return &fakeCatalogue{
		records: map[int]port.StockRecord{
			10: {ID: 10, ProductID: 100, PriceCurrency: "VND", Price: 150000, NumAvailable: &three},
			11: {ID: 11, ProductID: 101, PriceCurrency: "VND", Price: 20000, NumAvailable: &one},
			12: {ID: 12, ProductID: 102, PriceCurrency: "USD", Price: 3},
		},
		allocated: map[int]int{},
	}
}

func (f *fakeCatalogue) StockRecord(_ context.Context, id int) (*port.StockRecord, error) {
	r, ok := f.records[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	return &r, nil
}
func (f *fakeCatalogue) Allocate(_ context.Context, id, qty int) error {
	if id == f.failOn {
		return common.ErrOutOfStock
	}
	f.allocated[id] += qty
	return nil
}
func (f *fakeCatalogue) Release(_ context.Context, id, qty int) error {
	f.allocated[id] -= qty
	return nil
}

type fakeWallet struct {
	payments []port.WalletPayment
	err      error
}

func (f *fakeWallet) Pay(_ context.Context, p port.WalletPayment) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.payments = append(f.payments, p)
	return fmt.Sprintf("tx-%d", len(f.payments)), nil
}

type fakeAddresses struct {
	token string
	err   error
	plain bool
}

func (f *fakeAddresses) ShippingAddress(_ context.Context, token string, id int) (*port.ShippingAddress, error) {
	f.token = token
	if f.err != nil {
		return nil, f.err
	}
	if f.plain {
		return &port.ShippingAddress{Name: "Lưu Mạnh Quân", Phone: "0912345678", Street: "24 Lê Văn Hưu"}, nil
	}
	return &port.ShippingAddress{Name: "Lưu Mạnh Quân", Phone: "0912345678", Street: "24 Lê Văn Hưu", WardCode: "20109", DistrictID: 1442, ProvinceID: 202}, nil
}

type fakeRates struct {
	fee   int64
	err   error
	calls int
	items []port.ShippingItem
	to    port.ShippingAddress
}

func (f *fakeRates) Quote(_ context.Context, to port.ShippingAddress, items []port.ShippingItem, _ int64) (*port.ShippingQuote, error) {
	f.calls++
	f.to, f.items = to, items
	if f.err != nil {
		return nil, f.err
	}
	return &port.ShippingQuote{Fee: f.fee, ServiceName: "Chuẩn", WarehouseID: 2, ServiceID: 53320}, nil
}

type fixture struct {
	baskets   *memBaskets
	lines     *memLines
	catalogue *fakeCatalogue
	wallet    *fakeWallet
	orders    *fakeFulfillment
	addresses *fakeAddresses
	lineSvc   port.BasketLineUsecase
	checkout  port.CheckoutUsecase
	basketID  int
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		baskets:   &memBaskets{byID: map[int]*model.Basket{}},
		lines:     &memLines{byID: map[int]*model.BasketLine{}},
		catalogue: newCatalogue(),
		wallet:    &fakeWallet{},
		orders:    &fakeFulfillment{},
	}
	f.lineSvc = NewBasketLineService(f.lines, f.baskets, f.catalogue)
	f.addresses = &fakeAddresses{}
	f.checkout = NewCheckoutService(f.baskets, f.lines, f.catalogue, f.wallet, f.orders, f.addresses, nil)
	b, err := NewBasketService(f.baskets).CreateBasket(context.Background(), &model.Basket{UserID: 7, Status: "Submitted", TotalInclTax: 1})
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != model.BasketStatusOpen || b.TotalInclTax != 0 || b.DateCreated == "" {
		t.Fatalf("server must own basket lifecycle fields: %+v", b)
	}
	f.basketID = b.ID
	return f
}

func (f *fixture) add(t *testing.T, record, qty int) *model.BasketLine {
	t.Helper()
	l, err := f.lineSvc.CreateBasketLine(context.Background(), &model.BasketLine{BasketID: f.basketID, StockrecordID: record, Quantity: qty, PriceInclTax: 1, ProductID: 999})
	if err != nil {
		t.Fatalf("add line %d: %v", record, err)
	}
	return l
}

func TestLinesArePricedByTheCatalogue(t *testing.T) {
	f := newFixture(t)
	l := f.add(t, 10, 1)
	if l.PriceInclTax != 150000 || l.PriceCurrency != "VND" || l.ProductID != 100 {
		t.Fatalf("client price must be replaced by catalogue price: %+v", l)
	}
	merged := f.add(t, 10, 2)
	if merged.ID != l.ID || merged.Quantity != 3 {
		t.Fatalf("same stock record must merge into one line: %+v", merged)
	}
	if _, err := f.lineSvc.CreateBasketLine(context.Background(), &model.BasketLine{BasketID: f.basketID, StockrecordID: 10, Quantity: 1}); !errors.Is(err, common.ErrOutOfStock) {
		t.Fatalf("exceeding stock must fail, got %v", err)
	}
	for name, bad := range map[string]model.BasketLine{
		"no stock record": {BasketID: f.basketID, Quantity: 1},
		"zero quantity":   {BasketID: f.basketID, StockrecordID: 11},
	} {
		if _, err := f.lineSvc.CreateBasketLine(context.Background(), &bad); !errors.Is(err, common.ErrInvalidRequest) {
			t.Errorf("%s: got %v", name, err)
		}
	}
	updated, err := f.lineSvc.UpdateBasketLine(context.Background(), &model.BasketLine{ID: l.ID, Quantity: 1, BasketID: 99, StockrecordID: 11, PriceInclTax: 1})
	if err != nil {
		t.Fatal(err)
	}
	if updated.BasketID != f.basketID || updated.StockrecordID != 10 || updated.PriceInclTax != 150000 {
		t.Fatalf("update must keep basket, stock record and catalogue price: %+v", updated)
	}
}

func TestCheckoutChargesTheWalletAndSubmitsTheBasket(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 2)
	f.add(t, 11, 1)
	f.catalogue.records[10] = port.StockRecord{ID: 10, ProductID: 100, PriceCurrency: "VND", Price: 140000}

	b, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != model.BasketStatusSubmitted || b.TotalInclTax != 300000 || b.Currency != "VND" || b.PaymentReference != "tx-1" || b.ShippingAddressID != 4 {
		t.Fatalf("unexpected basket: %+v", b)
	}
	if _, err := time.Parse(time.RFC3339, b.DateSubmitted); err != nil {
		t.Fatalf("date_submitted: %v", err)
	}
	pay := f.wallet.payments[0]
	if pay.Amount != 300000 || pay.UserID != 7 || pay.WalletID != "w-1" || pay.Pin != "123456" || pay.IdempotencyKey == "" {
		t.Fatalf("unexpected payment: %+v", pay)
	}
	if f.catalogue.allocated[10] != 2 || f.catalogue.allocated[11] != 1 {
		t.Fatalf("stock not allocated: %v", f.catalogue.allocated)
	}
	if _, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "1"}); !errors.Is(err, common.ErrConflict) {
		t.Fatalf("second checkout must conflict, got %v", err)
	}
	if _, err := f.lineSvc.CreateBasketLine(context.Background(), &model.BasketLine{BasketID: f.basketID, StockrecordID: 11, Quantity: 1}); !errors.Is(err, common.ErrConflict) {
		t.Fatalf("submitted basket must be frozen, got %v", err)
	}
	if _, err := NewBasketService(f.baskets).UpdateBasket(context.Background(), &model.Basket{ID: f.basketID, UserID: 7, Status: "Open", TotalInclTax: 1}); err != nil {
		t.Fatal(err)
	}
	if stored := f.baskets.byID[f.basketID]; stored.Status != model.BasketStatusSubmitted || stored.TotalInclTax != 300000 {
		t.Fatalf("client must not reopen or reprice a submitted basket: %+v", stored)
	}
}

func TestCheckoutRollsBackStockWhenSomethingFails(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 2)
	f.add(t, 11, 1)
	f.catalogue.failOn = 11
	if _, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w", Pin: "1"}); !errors.Is(err, common.ErrOutOfStock) {
		t.Fatalf("got %v", err)
	}
	if f.catalogue.allocated[10] != 0 || len(f.wallet.payments) != 0 {
		t.Fatalf("partial allocation must be released and nothing charged: %v", f.catalogue.allocated)
	}

	f.catalogue.failOn = 0
	f.wallet.err = fmt.Errorf("%w: wrong pin", common.ErrPaymentFailed)
	if _, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w", Pin: "0"}); !errors.Is(err, common.ErrPaymentFailed) {
		t.Fatalf("got %v", err)
	}
	if f.catalogue.allocated[10] != 0 || f.catalogue.allocated[11] != 0 {
		t.Fatalf("failed payment must release stock: %v", f.catalogue.allocated)
	}
	if f.baskets.byID[f.basketID].Status != model.BasketStatusOpen {
		t.Fatal("basket must stay open after a failed payment")
	}
}

func TestCheckoutGuards(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if _, err := f.checkout.Checkout(ctx, port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w", Pin: "1"}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Fatalf("empty basket: %v", err)
	}
	f.add(t, 10, 1)
	f.add(t, 12, 1)
	if _, err := f.checkout.Checkout(ctx, port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w", Pin: "1"}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Fatalf("mixed currencies: %v", err)
	}
	if _, err := f.checkout.Checkout(ctx, port.CheckoutCommand{BasketID: f.basketID, UserID: 8, WalletID: "w", Pin: "1"}); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("someone else's basket: %v", err)
	}
	if _, err := f.checkout.Checkout(ctx, port.CheckoutCommand{BasketID: f.basketID, UserID: 7}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Fatalf("missing wallet and pin: %v", err)
	}
}

func TestCheckoutStartsTheOrderWorkflow(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 2)

	b, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w", Pin: "1", ShippingAddressID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if b.FulfillmentID != fmt.Sprintf("paid-order-basket-%d", f.basketID) {
		t.Fatalf("fulfillment id not recorded: %+v", b)
	}
	o := f.orders.orders[0]
	if o.UserID != 7 || o.Total != 300000 || o.PaymentReference != "tx-1" || o.ShippingAddressID != 4 || len(o.Lines) != 1 || o.Lines[0].Quantity != 2 || o.Lines[0].UnitPrice != 150000 {
		t.Fatalf("unexpected paid order: %+v", o)
	}
}

func TestFulfillmentOutageIsRetriedLater(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 1)
	f.orders.err = errors.New("temporal unavailable")

	b, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w", Pin: "1"})
	if err != nil {
		t.Fatalf("a paid basket must not fail checkout because of the order workflow: %v", err)
	}
	if b.Status != model.BasketStatusSubmitted || b.FulfillmentID != "" {
		t.Fatalf("unexpected basket: %+v", b)
	}

	f.orders.err = nil
	n, err := f.checkout.RetryFulfillment(context.Background(), 10)
	if err != nil || n != 1 || len(f.orders.orders) != 1 {
		t.Fatalf("retry: n=%d err=%v orders=%d", n, err, len(f.orders.orders))
	}
	if f.baskets.byID[f.basketID].FulfillmentID == "" {
		t.Fatal("retry must record the workflow id")
	}
	if n, _ := f.checkout.RetryFulfillment(context.Background(), 10); n != 0 {
		t.Fatalf("already fulfilled baskets must be skipped, started %d", n)
	}
}

func TestCheckoutSnapshotsTheGHNAddressForTheOrder(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 1)
	b, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4, AccessToken: "Bearer user-token"})
	if err != nil {
		t.Fatal(err)
	}
	if f.addresses.token != "Bearer user-token" {
		t.Fatalf("the buyer token must be forwarded, got %q", f.addresses.token)
	}
	if b.ShippingSnapshot == "" || len(f.orders.orders) != 1 {
		t.Fatalf("snapshot=%q orders=%d", b.ShippingSnapshot, len(f.orders.orders))
	}
	addr := f.orders.orders[0].ShippingAddress
	if addr == nil || addr.WardCode != "20109" || addr.DistrictID != 1442 || addr.Phone != "0912345678" {
		t.Fatalf("order address: %+v", addr)
	}
}

func TestCheckoutRejectsSomeoneElsesAddressBeforeCharging(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 1)
	f.addresses.err = common.ErrNotFound
	_, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 99})
	if !errors.Is(err, common.ErrInvalidRequest) {
		t.Fatalf("expected invalid request, got %v", err)
	}
	if len(f.wallet.payments) != 0 || f.catalogue.allocated[10] != 0 {
		t.Fatalf("nothing may be charged or reserved: payments=%d allocated=%v", len(f.wallet.payments), f.catalogue.allocated)
	}
}

func TestCheckoutStillSellsWhenTheAddressServiceIsDown(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 1)
	f.addresses.err = fmt.Errorf("%w: address service returned 503", common.ErrUpstream)
	b, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if b.ShippingSnapshot != "" || f.orders.orders[0].ShippingAddress != nil || f.orders.orders[0].ShippingAddressID != 4 {
		t.Fatalf("order must keep the address id without a snapshot: %+v", f.orders.orders[0])
	}
}

func TestRetriedFulfillmentKeepsTheAddressSnapshot(t *testing.T) {
	f := newFixture(t)
	f.add(t, 10, 1)
	f.orders.err = errors.New("temporal down")
	if _, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4}); err != nil {
		t.Fatal(err)
	}
	f.orders.err = nil
	if n, err := f.checkout.RetryFulfillment(context.Background(), 10); err != nil || n != 1 {
		t.Fatalf("retry: %d %v", n, err)
	}
	if addr := f.orders.orders[0].ShippingAddress; addr == nil || addr.WardCode != "20109" {
		t.Fatalf("retried order lost the address: %+v", addr)
	}
}

func withRates(t *testing.T, fee int64) (*fixture, *fakeRates) {
	t.Helper()
	f := newFixture(t)
	rates := &fakeRates{fee: fee}
	f.checkout = NewCheckoutService(f.baskets, f.lines, f.catalogue, f.wallet, f.orders, f.addresses, rates)
	f.add(t, 10, 2)
	return f, rates
}

func fee(v float64) *float64 { return &v }

func TestQuoteShippingAddsTheFeeToTheTotal(t *testing.T) {
	f, rates := withRates(t, 32000)
	q, err := f.checkout.QuoteShipping(context.Background(), port.ShippingQuoteCommand{BasketID: f.basketID, UserID: 7, ShippingAddressID: 4, AccessToken: "Bearer t"})
	if err != nil {
		t.Fatal(err)
	}
	if q.Subtotal != 300000 || q.ShippingFee != 32000 || q.Total != 332000 || !q.ShippingCharged || q.Currency != "VND" {
		t.Fatalf("quote %+v", q)
	}
	if rates.to.WardCode != "20109" || len(rates.items) != 1 || rates.items[0].Quantity != 2 || rates.items[0].Price != 150000 {
		t.Fatalf("rates got %+v %+v", rates.to, rates.items)
	}
	if len(f.wallet.payments) != 0 {
		t.Fatal("a quote must not charge")
	}
}

func TestCheckoutChargesTheConfirmedShippingFee(t *testing.T) {
	f, _ := withRates(t, 32000)
	b, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4, ShippingFee: fee(32000)})
	if err != nil {
		t.Fatal(err)
	}
	if f.wallet.payments[0].Amount != 332000 || f.wallet.payments[0].Metadata["shipping_fee"] != "32000" {
		t.Fatalf("payment %+v", f.wallet.payments[0])
	}
	if b.TotalInclTax != 332000 || b.ShippingInclTax != 32000 {
		t.Fatalf("basket %+v", b)
	}
	if o := f.orders.orders[0]; o.Total != 332000 || o.ShippingFee != 32000 || o.ShippingRoute == nil || o.ShippingRoute.WarehouseID != 2 || o.ShippingRoute.ServiceID != 53320 || o.ShippingRoute.Fee != 32000 {
		t.Fatalf("order %+v", o)
	}
}

func TestCheckoutNeverChargesMoreShippingThanConfirmed(t *testing.T) {
	for name, confirmed := range map[string]*float64{"unconfirmed": nil, "fee went up": fee(30000)} {
		f, _ := withRates(t, 32000)
		_, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4, ShippingFee: confirmed})
		if !errors.Is(err, common.ErrConflict) {
			t.Fatalf("%s: expected conflict, got %v", name, err)
		}
		if len(f.wallet.payments) != 0 || f.catalogue.allocated[10] != 0 {
			t.Fatalf("%s: nothing may be charged or reserved", name)
		}
	}
	f, _ := withRates(t, 28000)
	b, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4, ShippingFee: fee(32000)})
	if err != nil || b.ShippingInclTax != 28000 || f.wallet.payments[0].Amount != 328000 {
		t.Fatalf("a cheaper fee is charged as is: %+v %v", b, err)
	}
}

func TestCheckoutNeedsADeliverableGHNAddress(t *testing.T) {
	f, rates := withRates(t, 32000)
	f.addresses.plain = true
	_, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4, ShippingFee: fee(32000)})
	if !errors.Is(err, common.ErrInvalidRequest) || rates.calls != 0 || len(f.wallet.payments) != 0 {
		t.Fatalf("err=%v calls=%d", err, rates.calls)
	}
}

func TestCheckoutStopsWhenShippingCannotBePriced(t *testing.T) {
	for _, upstream := range []error{
		fmt.Errorf("%w: shipping service returned 503", common.ErrUpstream),
		fmt.Errorf("%w: không giao được tới địa chỉ này", common.ErrInvalidRequest),
	} {
		f, rates := withRates(t, 0)
		rates.err = upstream
		_, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4, ShippingFee: fee(50000)})
		if !errors.Is(err, upstream) || len(f.wallet.payments) != 0 {
			t.Fatalf("err=%v payments=%d", err, len(f.wallet.payments))
		}
	}
	f, _ := withRates(t, 32000)
	f.addresses.err = fmt.Errorf("%w: address service returned 503", common.ErrUpstream)
	if _, err := f.checkout.Checkout(context.Background(), port.CheckoutCommand{BasketID: f.basketID, UserID: 7, WalletID: "w-1", Pin: "123456", ShippingAddressID: 4, ShippingFee: fee(32000)}); !errors.Is(err, common.ErrUpstream) {
		t.Fatalf("without the address the fee is unknown, got %v", err)
	}
}
