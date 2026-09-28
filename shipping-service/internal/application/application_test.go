package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

var ctx = context.Background()

func TestQuoteRanksRoutesAcrossWarehouses(t *testing.T) {
	f := newFixture()
	in := validInput("")
	res, err := f.shipments.Quote(ctx, port.QuoteInput{Recipient: in.Recipient, Parcel: in.Parcel})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Options) != 2 || len(res.Unavailable) != 1 || res.Unavailable[0].ServiceID != 53322 {
		t.Fatalf("options=%+v unavailable=%+v", res.Options, res.Unavailable)
	}
	if res.Recommended[model.StrategyCheapest].WarehouseCode != "HCM" || res.Recommended[model.StrategyFastest].WarehouseCode != "HCM" {
		t.Errorf("recommended = %+v", res.Recommended)
	}
}

func TestQuoteRejectsAddressOutsideVietnamMasterData(t *testing.T) {
	f := newFixture()
	in := validInput("")
	in.Recipient.WardCode = "99999"
	_, err := f.shipments.Quote(ctx, port.QuoteInput{Recipient: in.Recipient, Parcel: in.Parcel})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("unknown ward: got %v", err)
	}
	in = validInput("")
	in.Recipient.ProvinceID = 201
	_, err = f.shipments.Quote(ctx, port.QuoteInput{Recipient: in.Recipient, Parcel: in.Parcel})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("district not in province: got %v", err)
	}
	in = validInput("")
	in.Recipient.Phone = "+1 415 555 0100"
	_, err = f.shipments.Quote(ctx, port.QuoteInput{Recipient: in.Recipient, Parcel: in.Parcel})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("foreign phone: got %v", err)
	}
}

func TestCreatePlacesCheapestRouteAndIsIdempotent(t *testing.T) {
	f := newFixture()
	in := validInput("DH-1")
	in.Strategy = model.StrategyCheapest

	s, created, err := f.shipments.Create(ctx, "order-service", in)
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	if s.Status != model.StatusCreated || s.WarehouseID != f.hcm.ID || s.ServiceID != 53321 {
		t.Fatalf("shipment = %+v", s)
	}
	if s.CarrierOrderCode != "GHNSHPTEST001" || s.ShippingFee != 31000 || s.QuotedFee != 22000 || s.Recipient.Phone != "0912345678" {
		t.Errorf("placement fields = %+v", s)
	}
	if req := f.carrier.created[0]; req.ShopID != 22 || req.ClientOrderCode != s.Code || req.CODAmount != 300000 {
		t.Errorf("carrier request = %+v", req)
	}

	again, created, err := f.shipments.Create(ctx, "order-service", in)
	if err != nil || created || again.ID != s.ID || len(f.carrier.created) != 1 {
		t.Errorf("second create: created=%v id=%d carrier calls=%d err=%v", created, again.ID, len(f.carrier.created), err)
	}

	other, created, err := f.shipments.Create(ctx, "shop-service", in)
	if err != nil || !created || other.ID == s.ID {
		t.Errorf("same client_order_code from another client must be separate: %v %v", created, err)
	}
}

func TestCreateWithExplicitRouteAndNoRoute(t *testing.T) {
	f := newFixture()
	in := validInput("DH-2")
	in.WarehouseID, in.ServiceID = f.hn.ID, 53320
	s, _, err := f.shipments.Create(ctx, "order-service", in)
	if err != nil || s.WarehouseID != f.hn.ID {
		t.Fatalf("explicit route: %+v %v", s, err)
	}

	in = validInput("DH-3")
	in.WarehouseID, in.ServiceID = f.hcm.ID, 53322
	_, _, err = f.shipments.Create(ctx, "order-service", in)
	if !errors.Is(err, port.ErrNoRoute) {
		t.Errorf("unserved route: got %v", err)
	}

	in = validInput("DH-4")
	in.CallbackURL = "http://evil.example"
	if _, _, err := f.shipments.Create(ctx, "order-service", in); !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("disallowed callback: got %v", err)
	}
}

func TestCarrierRejectionAndOutageThenRetryAdoptsExistingOrder(t *testing.T) {
	f := newFixture()
	f.carrier.createErr = fmt.Errorf("%w: ghn: Số điện thoại không hợp lệ", port.ErrCarrierRejected)
	s, _, err := f.shipments.Create(ctx, "order-service", validInput("DH-5"))
	if !errors.Is(err, port.ErrCarrierRejected) || s.Status != model.StatusRejected || s.PlaceAttempts != 1 {
		t.Fatalf("rejection: %+v %v", s, err)
	}

	f.carrier.createErr = fmt.Errorf("%w: timeout", port.ErrCarrierUnavailable)
	s, _, err = f.shipments.Create(ctx, "order-service", validInput("DH-6"))
	if !errors.Is(err, port.ErrCarrierUnavailable) || s.Status != model.StatusPending || s.LastError == "" {
		t.Fatalf("outage: %+v %v", s, err)
	}

	f.carrier.createErr = nil
	eta := baseTime.Add(30 * time.Hour)
	f.carrier.existing[s.Code] = &port.CarrierOrder{OrderCode: "GHN-LATE", Status: "ready_to_pick", ExpectedDeliveryAt: &eta}
	calls := len(f.carrier.created)
	placed, err := f.shipments.Place(ctx, "order-service", s.ID)
	if err != nil || placed.Status != model.StatusCreated || placed.CarrierOrderCode != "GHN-LATE" {
		t.Fatalf("retry: %+v %v", placed, err)
	}
	if len(f.carrier.created) != calls {
		t.Error("retry created a duplicate carrier order instead of adopting the existing one")
	}
	if placed.ShippingFee != placed.QuotedFee {
		t.Errorf("adopted order without fee should keep quoted fee, got %d", placed.ShippingFee)
	}
}

func TestPlacementLeasePreventsConcurrentPlacement(t *testing.T) {
	f := newFixture()
	f.carrier.createErr = fmt.Errorf("%w: timeout", port.ErrCarrierUnavailable)
	s, _, _ := f.shipments.Create(ctx, "order-service", validInput("DH-7"))
	until := baseTime.Add(time.Minute)
	sh := f.store.shipments[s.ID]
	sh.PlacingUntil = &until
	f.store.shipments[s.ID] = sh

	if _, err := f.shipments.Place(ctx, "order-service", s.ID); !errors.Is(err, port.ErrConflict) {
		t.Errorf("got %v, want ErrConflict", err)
	}
	if _, err := f.shipments.Cancel(ctx, "order-service", s.ID, ""); !errors.Is(err, port.ErrConflict) {
		t.Errorf("cancel during placement: got %v", err)
	}
}

func webhook(status, t string, code string) port.CarrierWebhook {
	return port.CarrierWebhook{OrderCode: code, Type: "switch_status", Status: status, Time: t}
}

func TestWebhookLifecycleOutOfOrderAndDuplicates(t *testing.T) {
	f := newFixture()
	s, _, _ := f.shipments.Create(ctx, "order-service", validInput("DH-8"))
	code := s.CarrierOrderCode

	if err := f.webhooks.HandleGHN(ctx, "wrong", webhook("picking", "", code)); !errors.Is(err, port.ErrUnauthorized) {
		t.Fatalf("bad token: %v", err)
	}

	steps := []struct{ status, at string }{
		{"picking", "2026-09-28T09:00:00Z"},
		{"sorting", "2026-09-28T12:00:00Z"},
		{"transporting", "2026-09-28T15:00:00Z"},
		{"delivering", "2026-09-29T08:00:00Z"},
		{"picked", "2026-09-28T10:00:00Z"},
		{"delivering", "2026-09-29T08:00:00Z"},
		{"delivery_fail", "2026-09-29T11:00:00Z"},
		{"delivering", "2026-09-30T08:00:00Z"},
		{"delivered", "2026-09-30T09:30:00Z"},
		{"returning", "2026-09-30T10:00:00Z"},
	}
	for _, st := range steps {
		if err := f.webhooks.HandleGHN(ctx, "hook-secret", webhook(st.status, st.at, code)); err != nil {
			t.Fatalf("%s: %v", st.status, err)
		}
	}

	d, _ := f.shipments.Get(ctx, "order-service", s.ID)
	if d.Shipment.Status != model.StatusDelivered || d.Shipment.DeliveredAt == nil || !d.Shipment.DeliveredAt.Equal(time.Date(2026, 9, 30, 9, 30, 0, 0, time.UTC)) {
		t.Fatalf("final = %s delivered_at=%v", d.Shipment.Status, d.Shipment.DeliveredAt)
	}
	var applied, ignored int
	for _, e := range d.Events {
		if e.Applied {
			applied++
		} else {
			ignored++
		}
	}
	if applied != 8 || ignored != 3 || len(d.Events) != 11 {
		t.Errorf("applied=%d ignored=%d events=%d, want 8 applied, 3 recorded but not applied (same-status hop, late picked, returning after delivered) and the duplicate delivering dropped", applied, ignored, len(d.Events))
	}

	if _, err := f.shipments.Get(ctx, "shop-service", s.ID); !errors.Is(err, port.ErrNotFound) {
		t.Error("another client could read the shipment")
	}
}

func TestWebhookForUnknownOrderIsAcknowledged(t *testing.T) {
	f := newFixture()
	if err := f.webhooks.HandleGHN(ctx, "hook-secret", webhook("delivered", "", "NOT-OURS")); err != nil {
		t.Errorf("got %v", err)
	}
}

func TestCancelRules(t *testing.T) {
	f := newFixture()
	s, _, _ := f.shipments.Create(ctx, "order-service", validInput("DH-9"))
	cancelled, err := f.shipments.Cancel(ctx, "order-service", s.ID, "khách đổi ý")
	if err != nil || cancelled.Status != model.StatusCancelled || len(f.carrier.cancelled) != 1 {
		t.Fatalf("cancel created: %+v %v", cancelled, err)
	}

	s2, _, _ := f.shipments.Create(ctx, "order-service", validInput("DH-10"))
	_ = f.webhooks.HandleGHN(ctx, "hook-secret", webhook("picked", "2026-09-28T10:00:00Z", s2.CarrierOrderCode))
	if _, err := f.shipments.Cancel(ctx, "order-service", s2.ID, ""); !errors.Is(err, port.ErrInvalidTransition) {
		t.Errorf("cancel after pickup: got %v", err)
	}

	s3, _, _ := f.shipments.Create(ctx, "order-service", validInput("DH-11"))
	f.carrier.cancelErr = fmt.Errorf("%w: đơn đã lấy", port.ErrCarrierRejected)
	if _, err := f.shipments.Cancel(ctx, "order-service", s3.ID, ""); !errors.Is(err, port.ErrCarrierRejected) {
		t.Errorf("carrier refusal: got %v", err)
	}
	if got, _ := f.shipments.Get(ctx, "order-service", s3.ID); got.Shipment.Status != model.StatusCreated {
		t.Errorf("status changed although carrier refused: %s", got.Shipment.Status)
	}
}

func TestSyncAppliesCarrierStatus(t *testing.T) {
	f := newFixture()
	s, _, _ := f.shipments.Create(ctx, "order-service", validInput("DH-12"))
	updated := baseTime.Add(5 * time.Hour)
	f.carrier.remote[s.CarrierOrderCode] = &port.CarrierOrder{OrderCode: s.CarrierOrderCode, Status: "delivering", UpdatedAt: &updated}
	got, err := f.shipments.Sync(ctx, "order-service", s.ID)
	if err != nil || got.Status != model.StatusDelivering {
		t.Errorf("sync: %+v %v", got, err)
	}
}

func TestOutboxDeliversNotificationsAndCallbacksWithRetry(t *testing.T) {
	f := newFixture()
	s, _, _ := f.shipments.Create(ctx, "order-service", validInput("DH-13"))
	_ = f.webhooks.HandleGHN(ctx, "hook-secret", webhook("delivering", "2026-09-29T08:00:00Z", s.CarrierOrderCode))

	f.notifier.err = errors.New("notification-service down")
	n, err := f.outbox.DispatchDue(ctx)
	if err != nil || n != 6 {
		t.Fatalf("dispatched %d, %v", n, err)
	}
	if len(f.callbacks.sent) != 2 || f.callbacks.sent[1].Status != model.StatusDelivering || f.callbacks.sent[1].PreviousStatus != model.StatusCreated {
		t.Fatalf("callbacks = %+v", f.callbacks.sent)
	}

	f.notifier.err = nil
	if n, _ := f.outbox.DispatchDue(ctx); n != 0 {
		t.Errorf("failed jobs retried before their backoff: %d", n)
	}
	f.clock.now = f.clock.now.Add(Backoff(1))
	if n, _ := f.outbox.DispatchDue(ctx); n != 4 {
		t.Errorf("after backoff dispatched %d, want 4", n)
	}
	if len(f.notifier.sent) != 4 || f.notifier.sent[0].Channel != "email" {
		t.Errorf("notifications = %+v", f.notifier.sent)
	}
}

func TestOutboxGivesUpAfterMaxAttempts(t *testing.T) {
	f := newFixture()
	_, _ = f.outbox.DispatchDue(ctx)
	f.notifier.err = errors.New("down")
	_, _, _ = f.shipments.Create(ctx, "order-service", port.CreateShipmentInput{
		ClientOrderCode: "DH-14", Recipient: validInput("").Recipient, Parcel: validInput("").Parcel,
		Customer: model.Customer{Email: "x@y.vn"},
	})
	for i := 0; i < DefaultSettings().OutboxMaxAttempts+2; i++ {
		_, _ = f.outbox.DispatchDue(ctx)
		f.clock.now = f.clock.now.Add(2 * time.Hour)
	}
	for _, j := range f.store.jobs {
		if j.State != model.JobFailed || j.Attempts != DefaultSettings().OutboxMaxAttempts {
			t.Errorf("job %+v", j)
		}
	}
}

func TestBackoff(t *testing.T) {
	if Backoff(1) != 5*time.Second || Backoff(3) != 20*time.Second || Backoff(30) != time.Hour {
		t.Errorf("backoff: %v %v %v", Backoff(1), Backoff(3), Backoff(30))
	}
}
