package activity

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/internal/proxy"
)

type memStore struct {
	shipments map[string]entity.ShipmentResult
}

func (m *memStore) CreatePaidOrder(context.Context, entity.PaidOrderInput) (*entity.PaidOrderRecord, error) {
	return nil, nil
}
func (m *memStore) SetStatus(context.Context, string, string) error { return nil }
func (m *memStore) SetShipment(_ context.Context, number string, s entity.ShipmentResult) error {
	if m.shipments == nil {
		m.shipments = map[string]entity.ShipmentResult{}
	}
	m.shipments[number] = s
	return nil
}

type fakeShipments struct {
	got      entity.ShipmentRequest
	sent     []entity.ShipmentRequest
	shipment *entity.Shipment
	err      error
	noRoute  bool
}

func (f *fakeShipments) CreateShipment(_ context.Context, req entity.ShipmentRequest) (*entity.Shipment, error) {
	f.got = req
	f.sent = append(f.sent, req)
	if f.noRoute && req.ServiceID != 0 {
		return nil, fmt.Errorf("%w: no route", proxy.ErrShipmentRejected)
	}
	return f.shipment, f.err
}

func order() entity.PaidOrderInput {
	return entity.PaidOrderInput{
		BasketID: 12, UserID: 7, Currency: "VND", Total: 12_000_000, PaymentReference: "tx-1",
		ShippingAddress: &entity.ShippingAddress{
			Name: "Lưu Mạnh Quân", Phone: "0912345678", Street: "24 Lê Văn Hưu",
			WardCode: "20109", DistrictID: 1442, ProvinceID: 202,
		},
		Lines: []entity.PaidOrderLine{
			{ProductID: 3, Title: "Áo thun", Quantity: 2, UnitPrice: 150000},
			{ProductID: 4, Quantity: 1, UnitPrice: 11_700_000},
		},
	}
}

func TestCreateShipmentBuildsGHNRequest(t *testing.T) {
	ships := &fakeShipments{shipment: &entity.Shipment{ID: 9, Code: "SHP-9", Status: "CREATED", CarrierOrderCode: "GHN1", QuotedFee: 32000}}
	store := &memStore{}
	a := &PaidOrderActivities{Store: store, Shipments: ships}

	res, err := a.CreateShipment(context.Background(), order(), "KS00000012")
	if err != nil || res.Outcome != entity.ShipmentOutcomeCreated || res.Code != "SHP-9" || res.Fee != 32000 {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	req := ships.got
	if req.ClientOrderCode != "KS00000012" || req.CODAmount != 0 || req.PaymentType != "SENDER" || req.Strategy != "CHEAPEST" {
		t.Fatalf("request=%+v", req)
	}
	if req.Recipient.WardCode != "20109" || req.Recipient.DistrictID != 1442 || req.Recipient.Phone != "0912345678" {
		t.Fatalf("recipient=%+v", req.Recipient)
	}
	if req.InsuranceValue != 5_000_000 {
		t.Fatalf("insurance must be capped, got %d", req.InsuranceValue)
	}
	p := req.Parcel
	if p.WeightGram != 1500 || p.HeightCm != 25 || len(p.Items) != 2 || p.Items[1].Name != "Sản phẩm #4" || p.Items[0].Price != 150000 {
		t.Fatalf("parcel=%+v", p)
	}
	if store.shipments["KS00000012"].Code != "SHP-9" {
		t.Fatalf("shipment not stored: %+v", store.shipments)
	}
}

func TestCreateShipmentSkipsOrdersWithoutGHNAddress(t *testing.T) {
	for _, in := range []entity.PaidOrderInput{
		{BasketID: 1},
		{BasketID: 2, ShippingAddress: &entity.ShippingAddress{Name: "A", Phone: "0912345678", Street: "x"}},
	} {
		ships := &fakeShipments{}
		store := &memStore{}
		res, err := (&PaidOrderActivities{Store: store, Shipments: ships}).CreateShipment(context.Background(), in, in.Number())
		if err != nil || res.Outcome != entity.ShipmentOutcomeSkipped || ships.got.ClientOrderCode != "" {
			t.Fatalf("res=%+v err=%v sent=%+v", res, err, ships.got)
		}
		if store.shipments[in.Number()].Outcome != entity.ShipmentOutcomeSkipped {
			t.Fatal("skip must be recorded so staff can ship by hand")
		}
	}
	res, _ := (&PaidOrderActivities{Store: &memStore{}}).CreateShipment(context.Background(), order(), "KS00000012")
	if res.Outcome != entity.ShipmentOutcomeSkipped {
		t.Fatalf("unconfigured shipping must skip, got %+v", res)
	}
}

func TestCreateShipmentRejectionsAreFinalButOutagesRetry(t *testing.T) {
	rejected := &fakeShipments{shipment: &entity.Shipment{ID: 3, Code: "SHP-3", Status: "REJECTED"}, err: fmt.Errorf("%w: no route", proxy.ErrShipmentRejected)}
	res, err := (&PaidOrderActivities{Store: &memStore{}, Shipments: rejected}).CreateShipment(context.Background(), order(), "KS00000012")
	if err != nil || res.Outcome != entity.ShipmentOutcomeRejected || res.Code != "SHP-3" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	invalid := &fakeShipments{err: fmt.Errorf("%w: 400 bad phone", proxy.ErrShipmentInvalid)}
	if res, err := (&PaidOrderActivities{Store: &memStore{}, Shipments: invalid}).CreateShipment(context.Background(), order(), "KS1"); err != nil || res.Outcome != entity.ShipmentOutcomeRejected {
		t.Fatalf("invalid: res=%+v err=%v", res, err)
	}
	down := &fakeShipments{err: errors.New("shipping-service: status 502")}
	store := &memStore{}
	if _, err := (&PaidOrderActivities{Store: store, Shipments: down}).CreateShipment(context.Background(), order(), "KS2"); err == nil {
		t.Fatal("an outage must be returned so Temporal retries")
	}
	if _, ok := store.shipments["KS2"]; ok {
		t.Fatal("nothing should be stored before the retry succeeds")
	}
}

func TestShippingFeeIsNotInsured(t *testing.T) {
	in := order()
	in.Total, in.ShippingFee = 332000, 32000
	ships := &fakeShipments{shipment: &entity.Shipment{ID: 1, Code: "SHP-1"}}
	if _, err := (&PaidOrderActivities{Store: &memStore{}, Shipments: ships}).CreateShipment(context.Background(), in, "KS1"); err != nil {
		t.Fatal(err)
	}
	if ships.got.InsuranceValue != 300000 || ships.got.CODAmount != 0 {
		t.Fatalf("insurance=%d cod=%d", ships.got.InsuranceValue, ships.got.CODAmount)
	}
	if msg := orderPlacedMessage(in, "KS1"); msg != "Đơn KS1 đã được thanh toán 332000 VND (gồm phí giao hàng 32000 VND) và đang được xử lý." {
		t.Fatalf("message %q", msg)
	}
}

func TestShipmentUsesTheRouteTheBuyerPaidFor(t *testing.T) {
	in := order()
	in.Total, in.ShippingFee = 332000, 32000
	in.ShippingRoute = &entity.ShippingRoute{WarehouseID: 2, ServiceID: 53320, Fee: 32000}
	ships := &fakeShipments{shipment: &entity.Shipment{ID: 1, Code: "SHP-1", QuotedFee: 32500}}
	res, err := (&PaidOrderActivities{Store: &memStore{}, Shipments: ships}).CreateShipment(context.Background(), in, "KS1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ships.sent) != 1 || ships.got.WarehouseID != 2 || ships.got.ServiceID != 53320 {
		t.Fatalf("sent %+v", ships.sent)
	}
	if res.Fee != 32500 || res.PaidFee != 32000 || res.RouteChanged {
		t.Fatalf("result %+v", res)
	}
}

func TestShipmentFallsBackWhenThePaidRouteIsGone(t *testing.T) {
	in := order()
	in.ShippingRoute = &entity.ShippingRoute{WarehouseID: 2, ServiceID: 53320}
	ships := &fakeShipments{noRoute: true, shipment: &entity.Shipment{ID: 1, Code: "SHP-1"}}
	res, err := (&PaidOrderActivities{Store: &memStore{}, Shipments: ships}).CreateShipment(context.Background(), in, "KS1")
	if err != nil || res.Outcome != entity.ShipmentOutcomeCreated || !res.RouteChanged {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if len(ships.sent) != 2 || ships.sent[1].ServiceID != 0 || ships.sent[1].Strategy != "CHEAPEST" {
		t.Fatalf("fallback must drop the locked route: %+v", ships.sent)
	}
}
