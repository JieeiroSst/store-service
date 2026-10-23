package activity

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/repository"
	"github.com/JIeeiroSst/order-processing-service/internal/proxy"
	"go.temporal.io/sdk/temporal"
)

const (
	maxInsuranceValue = 5_000_000
	maxParcelItems    = 6
)

type ParcelDefaults struct {
	ItemWeightGram int
	LengthCm       int
	WidthCm        int
	BaseHeightCm   int
	ItemHeightCm   int
}

func DefaultParcel() ParcelDefaults {
	return ParcelDefaults{ItemWeightGram: 500, LengthCm: 30, WidthCm: 20, BaseHeightCm: 10, ItemHeightCm: 5}
}

type PaidOrderActivities struct {
	Store     repository.PaidOrderStore
	Notifier  proxy.NotificationProxy
	Shipments proxy.ShipmentProxy
	Parcel    ParcelDefaults
}

func (a *PaidOrderActivities) RecordPaidOrder(ctx context.Context, in entity.PaidOrderInput) (*entity.PaidOrderRecord, error) {
	if err := in.Validate(); err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "InvalidPaidOrder", err)
	}
	return a.Store.CreatePaidOrder(ctx, in)
}

func (a *PaidOrderActivities) NotifyOrderPlaced(ctx context.Context, in entity.PaidOrderInput, number string) (bool, error) {
	resp, err := a.Notifier.Send(ctx, entity.NotificationRequest{
		CustomerID: strconv.Itoa(in.UserID),
		OrderID:    number,
		Type:       "push",
		Message:    orderPlacedMessage(in, number),
	})
	if err != nil {
		return false, err
	}
	return resp.Sent, nil
}

func (a *PaidOrderActivities) SetPaidOrderStatus(ctx context.Context, number, status string) error {
	return a.Store.SetStatus(ctx, number, status)
}

func (a *PaidOrderActivities) CreateShipment(ctx context.Context, in entity.PaidOrderInput, number string) (*entity.ShipmentResult, error) {
	result, err := a.requestShipment(ctx, in, number)
	if err != nil {
		return nil, err
	}
	if err := a.Store.SetShipment(ctx, number, *result); err != nil {
		return nil, err
	}
	return result, nil
}

func (a *PaidOrderActivities) requestShipment(ctx context.Context, in entity.PaidOrderInput, number string) (*entity.ShipmentResult, error) {
	if a.Shipments == nil {
		return &entity.ShipmentResult{Outcome: entity.ShipmentOutcomeSkipped, Reason: "shipping-service is not configured"}, nil
	}
	if !in.ShippingAddress.Deliverable() {
		return &entity.ShipmentResult{Outcome: entity.ShipmentOutcomeSkipped, Reason: "order has no delivery address with GHN ward and district"}, nil
	}
	req := a.shipmentRequest(in, number)
	routeChanged := false
	if r := in.ShippingRoute; r != nil && r.WarehouseID > 0 && r.ServiceID > 0 {
		req.WarehouseID, req.ServiceID = r.WarehouseID, r.ServiceID
	}
	shipment, err := a.Shipments.CreateShipment(ctx, req)
	if errors.Is(err, proxy.ErrShipmentRejected) && shipment == nil && req.ServiceID != 0 {
		req.WarehouseID, req.ServiceID = 0, 0
		routeChanged = true
		shipment, err = a.Shipments.CreateShipment(ctx, req)
	}
	switch {
	case err == nil:
		r := created(shipment)
		r.RouteChanged = routeChanged
		if in.ShippingFee > 0 {
			r.PaidFee = int64(math.Round(in.ShippingFee))
		}
		return r, nil
	case errors.Is(err, proxy.ErrShipmentRejected):
		r := &entity.ShipmentResult{Outcome: entity.ShipmentOutcomeRejected, Reason: err.Error()}
		if shipment != nil {
			r.ShipmentID, r.Code, r.Status = shipment.ID, shipment.Code, shipment.Status
		}
		return r, nil
	case errors.Is(err, proxy.ErrShipmentInvalid), errors.Is(err, proxy.ErrShipmentAuth):
		return &entity.ShipmentResult{Outcome: entity.ShipmentOutcomeRejected, Reason: err.Error()}, nil
	default:
		return nil, err
	}
}

func created(s *entity.Shipment) *entity.ShipmentResult {
	return &entity.ShipmentResult{
		Outcome:          entity.ShipmentOutcomeCreated,
		ShipmentID:       s.ID,
		Code:             s.Code,
		Status:           s.Status,
		CarrierOrderCode: s.CarrierOrderCode,
		Fee:              s.QuotedFee,
	}
}

func (a *PaidOrderActivities) shipmentRequest(in entity.PaidOrderInput, number string) entity.ShipmentRequest {
	d := a.Parcel
	if d.ItemWeightGram <= 0 {
		d = DefaultParcel()
	}
	addr := in.ShippingAddress
	items := make([]entity.ShipmentItem, 0, len(in.Lines))
	units := 0
	for _, l := range in.Lines {
		name := l.Title
		if name == "" {
			name = fmt.Sprintf("Sản phẩm #%d", l.ProductID)
		}
		items = append(items, entity.ShipmentItem{
			Name:       name,
			Code:       strconv.Itoa(l.ProductID),
			Quantity:   l.Quantity,
			WeightGram: d.ItemWeightGram,
			Price:      int64(math.Round(l.UnitPrice)),
		})
		units += l.Quantity
	}
	return entity.ShipmentRequest{
		ClientOrderCode: number,
		Recipient: entity.ShipmentRecipient{
			Name: addr.Name, Phone: addr.Phone, Street: addr.Street,
			WardCode: addr.WardCode, DistrictID: addr.DistrictID, ProvinceID: addr.ProvinceID,
		},
		Parcel: entity.ShipmentParcel{
			WeightGram: max(units, 1) * d.ItemWeightGram,
			LengthCm:   d.LengthCm,
			WidthCm:    d.WidthCm,
			HeightCm:   d.BaseHeightCm + d.ItemHeightCm*min(max(units, 1), maxParcelItems),
			Content:    "Đơn " + number,
			Items:      items,
		},
		InsuranceValue: min(int64(math.Round(in.GoodsTotal())), maxInsuranceValue),
		PaymentType:    "SENDER",
		RequiredNote:   "CHOXEMHANGKHONGTHU",
		Note:           "Đơn đã thanh toán qua Ví Keiki, không thu tiền người nhận",
		Strategy:       "CHEAPEST",
		Customer:       entity.ShipmentCustomer{UserID: uint(in.UserID)},
	}
}

func orderPlacedMessage(in entity.PaidOrderInput, number string) string {
	if in.ShippingFee > 0 {
		return fmt.Sprintf("Đơn %s đã được thanh toán %.0f %s (gồm phí giao hàng %.0f %s) và đang được xử lý.", number, in.Total, in.Currency, in.ShippingFee, in.Currency)
	}
	return fmt.Sprintf("Đơn %s đã được thanh toán %.0f %s và đang được xử lý.", number, in.Total, in.Currency)
}
