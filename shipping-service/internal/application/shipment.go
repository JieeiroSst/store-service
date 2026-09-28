package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type shipmentService struct {
	planner   *planner
	lifecycle *lifecycle
	shipments port.ShipmentRepository
	events    port.EventRepository
	carrier   port.Carrier
	callbacks port.CallbackSender
	codes     port.CodeGenerator
	clock     port.Clock
	settings  Settings
}

func NewShipmentService(
	shipments port.ShipmentRepository,
	events port.EventRepository,
	warehouses port.WarehouseRepository,
	outbox port.OutboxRepository,
	carrier port.Carrier,
	directory port.LocationDirectory,
	callbacks port.CallbackSender,
	codes port.CodeGenerator,
	tx port.TxManager,
	clock port.Clock,
	settings Settings,
) port.ShipmentUsecase {
	return &shipmentService{
		planner:   &planner{carrier: carrier, directory: directory, warehouses: warehouses, settings: settings},
		lifecycle: &lifecycle{shipments: shipments, events: events, outbox: outbox, tx: tx, clock: clock},
		shipments: shipments,
		events:    events,
		carrier:   carrier,
		callbacks: callbacks,
		codes:     codes,
		clock:     clock,
		settings:  settings,
	}
}

func (s *shipmentService) Quote(ctx context.Context, in port.QuoteInput) (*port.QuoteResult, error) {
	if err := s.planner.validate(ctx, &in); err != nil {
		return nil, err
	}
	options, unavailable, err := s.planner.plan(ctx, in)
	if err != nil {
		return nil, err
	}
	result := &port.QuoteResult{
		Options:     model.RankRoutes(options, model.DefaultRouteStrategy),
		Recommended: map[model.Strategy]*model.RouteOption{},
		Unavailable: unavailable,
	}
	for _, strategy := range []model.Strategy{model.StrategyCheapest, model.StrategyFastest, model.StrategyBalanced} {
		if ranked := model.RankRoutes(options, strategy); len(ranked) > 0 {
			best := ranked[0]
			result.Recommended[strategy] = &best
		}
	}
	return result, nil
}

func (s *shipmentService) Create(ctx context.Context, client string, in port.CreateShipmentInput) (*model.Shipment, bool, error) {
	in.ClientOrderCode = strings.TrimSpace(in.ClientOrderCode)
	if in.ClientOrderCode == "" {
		return nil, false, invalid("client_order_code is required")
	}

	existing, err := s.shipments.GetByClientOrderCode(ctx, client, in.ClientOrderCode)
	switch {
	case err == nil:
		if existing.Status == model.StatusPending || existing.Status == model.StatusRejected {
			placed, err := s.place(ctx, existing.ID)
			return placed, false, err
		}
		return existing, false, nil
	case !errors.Is(err, port.ErrNotFound):
		return nil, false, err
	}

	shipment, err := s.prepare(ctx, client, in)
	if err != nil {
		return nil, false, err
	}
	if err := s.shipments.Create(ctx, shipment); err != nil {
		if errors.Is(err, port.ErrConflict) {
			existing, getErr := s.shipments.GetByClientOrderCode(ctx, client, in.ClientOrderCode)
			if getErr == nil {
				return existing, false, nil
			}
		}
		return nil, false, err
	}
	if _, err := s.events.Create(ctx, &model.ShipmentEvent{
		ShipmentID: shipment.ID,
		Source:     model.SourceAPI,
		ToStatus:   model.StatusPending,
		Reason:     fmt.Sprintf("route %s via warehouse %d service %d", shipment.Strategy, shipment.WarehouseID, shipment.ServiceID),
		Applied:    true,
		OccurredAt: s.clock.Now().UTC(),
	}); err != nil {
		return nil, false, err
	}

	placed, err := s.place(ctx, shipment.ID)
	return placed, true, err
}

func (s *shipmentService) prepare(ctx context.Context, client string, in port.CreateShipmentInput) (*model.Shipment, error) {
	if in.PaymentType == "" {
		in.PaymentType = model.DefaultPaymentType
	}
	if in.RequiredNote == "" {
		in.RequiredNote = model.DefaultRequiredNote
	}
	if in.Strategy == "" {
		in.Strategy = model.DefaultRouteStrategy
	}
	switch {
	case !in.PaymentType.Valid():
		return nil, invalid("payment_type must be SENDER or RECIPIENT")
	case !in.RequiredNote.Valid():
		return nil, invalid("required_note must be CHOTHUHANG, CHOXEMHANGKHONGTHU or KHONGCHOXEMHANG")
	case !in.Strategy.Valid():
		return nil, invalid("strategy must be CHEAPEST, FASTEST or BALANCED")
	case (in.WarehouseID == 0) != (in.ServiceID == 0):
		return nil, invalid("warehouse_id and service_id must be given together")
	}
	customer, err := in.Customer.Normalize()
	if err != nil {
		return nil, invalid("%v", err)
	}
	in.CallbackURL = strings.TrimSpace(in.CallbackURL)
	if in.CallbackURL != "" {
		if err := s.callbacks.Allowed(in.CallbackURL); err != nil {
			return nil, invalid("callback_url: %v", err)
		}
	}

	quoteIn := port.QuoteInput{
		Recipient:      in.Recipient,
		Parcel:         in.Parcel,
		CODAmount:      in.CODAmount,
		InsuranceValue: in.InsuranceValue,
		WarehouseIDs:   in.WarehouseIDs,
	}
	if in.WarehouseID != 0 {
		quoteIn.WarehouseIDs = []int64{in.WarehouseID}
	}
	if err := s.planner.validate(ctx, &quoteIn); err != nil {
		return nil, err
	}
	options, unavailable, err := s.planner.plan(ctx, quoteIn)
	if err != nil {
		return nil, err
	}

	var chosen *model.RouteOption
	for _, o := range model.RankRoutes(options, in.Strategy) {
		if in.ServiceID == 0 || o.ServiceID == in.ServiceID {
			o := o
			chosen = &o
			break
		}
	}
	if chosen == nil {
		reasons := make([]string, 0, len(unavailable))
		for _, u := range unavailable {
			reasons = append(reasons, fmt.Sprintf("%s/%d: %s", u.WarehouseCode, u.ServiceID, u.Reason))
		}
		return nil, fmt.Errorf("%w: %s", port.ErrNoRoute, strings.Join(reasons, "; "))
	}

	eta := chosen.ExpectedDeliveryAt
	return &model.Shipment{
		Code:               s.codes.NewShipmentCode(),
		ClientService:      client,
		ClientOrderCode:    in.ClientOrderCode,
		Status:             model.StatusPending,
		WarehouseID:        chosen.WarehouseID,
		CarrierShopID:      chosen.CarrierShopID,
		ServiceID:          chosen.ServiceID,
		ServiceTypeID:      chosen.ServiceTypeID,
		ServiceName:        chosen.ServiceName,
		Strategy:           in.Strategy,
		Recipient:          quoteIn.Recipient,
		Parcel:             quoteIn.Parcel,
		Customer:           customer,
		CODAmount:          in.CODAmount,
		InsuranceValue:     in.InsuranceValue,
		PaymentType:        in.PaymentType,
		RequiredNote:       in.RequiredNote,
		Note:               strings.TrimSpace(in.Note),
		QuotedFee:          chosen.Fee,
		ExpectedDeliveryAt: &eta,
		CallbackURL:        in.CallbackURL,
	}, nil
}

func (s *shipmentService) Place(ctx context.Context, client string, id int64) (*model.Shipment, error) {
	shipment, err := s.owned(ctx, client, id)
	if err != nil {
		return nil, err
	}
	if shipment.Status != model.StatusPending && shipment.Status != model.StatusRejected {
		return shipment, nil
	}
	return s.place(ctx, shipment.ID)
}

func (s *shipmentService) place(ctx context.Context, id int64) (*model.Shipment, error) {
	now := s.clock.Now()
	acquired, err := s.shipments.AcquirePlacement(ctx, id, now, now.Add(s.settings.PlacementLease))
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, fmt.Errorf("%w: shipment %d is being placed with the carrier", port.ErrConflict, id)
	}
	defer s.shipments.ReleasePlacement(context.WithoutCancel(ctx), id)

	shipment, err := s.shipments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if shipment.Status != model.StatusPending && shipment.Status != model.StatusRejected {
		return shipment, nil
	}

	var order *port.CarrierOrder
	if shipment.PlaceAttempts > 0 {
		order, err = s.carrier.FindByClientOrderCode(ctx, shipment.CarrierShopID, shipment.Code)
		if err != nil && !errors.Is(err, port.ErrNotFound) {
			return s.recordPlacementFailure(ctx, shipment, err)
		}
	}
	if order == nil {
		order, err = s.carrier.CreateOrder(ctx, port.CarrierOrderRequest{
			ShopID:          shipment.CarrierShopID,
			ClientOrderCode: shipment.Code,
			ServiceID:       shipment.ServiceID,
			ServiceTypeID:   shipment.ServiceTypeID,
			Recipient:       shipment.Recipient,
			Parcel:          shipment.Parcel,
			CODAmount:       shipment.CODAmount,
			InsuranceValue:  shipment.InsuranceValue,
			PaymentType:     shipment.PaymentType,
			RequiredNote:    shipment.RequiredNote,
			Note:            shipment.Note,
		})
		if err != nil {
			return s.recordPlacementFailure(ctx, shipment, err)
		}
	}

	carrierStatus := order.Status
	if carrierStatus == "" {
		carrierStatus = "ready_to_pick"
	}
	placed, _, err := s.lifecycle.apply(ctx, shipment.ID, statusUpdate{
		To:            model.StatusCreated,
		CarrierStatus: carrierStatus,
		Source:        model.SourceCarrier,
		Mutate: func(sh *model.Shipment) {
			sh.PlaceAttempts++
			sh.CarrierOrderCode = order.OrderCode
			sh.ShippingFee = sh.QuotedFee
			if order.Fee > 0 {
				sh.ShippingFee = order.Fee
			}
			sh.LastError = ""
			if order.ExpectedDeliveryAt != nil {
				sh.ExpectedDeliveryAt = order.ExpectedDeliveryAt
			}
		},
	})
	if err != nil {
		return nil, err
	}
	return placed, nil
}

func (s *shipmentService) recordPlacementFailure(ctx context.Context, shipment *model.Shipment, cause error) (*model.Shipment, error) {
	if errors.Is(cause, port.ErrCarrierRejected) {
		updated, _, err := s.lifecycle.apply(ctx, shipment.ID, statusUpdate{
			To:     model.StatusRejected,
			Reason: cause.Error(),
			Source: model.SourceCarrier,
			Mutate: func(sh *model.Shipment) {
				sh.PlaceAttempts++
				sh.LastError = cause.Error()
			},
		})
		if err != nil {
			return nil, err
		}
		if updated.Status == model.StatusRejected {
			return updated, cause
		}
	}

	shipment.PlaceAttempts++
	shipment.LastError = cause.Error()
	if err := s.shipments.Save(ctx, shipment); err != nil {
		return nil, err
	}
	return shipment, cause
}

func (s *shipmentService) owned(ctx context.Context, client string, id int64) (*model.Shipment, error) {
	shipment, err := s.shipments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if shipment.ClientService != client {
		return nil, fmt.Errorf("%w: shipment %d", port.ErrNotFound, id)
	}
	return shipment, nil
}

func (s *shipmentService) detail(ctx context.Context, shipment *model.Shipment) (*port.ShipmentDetail, error) {
	events, err := s.events.ListByShipment(ctx, shipment.ID)
	if err != nil {
		return nil, err
	}
	return &port.ShipmentDetail{Shipment: *shipment, Events: events}, nil
}

func (s *shipmentService) Get(ctx context.Context, client string, id int64) (*port.ShipmentDetail, error) {
	shipment, err := s.owned(ctx, client, id)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, shipment)
}

func (s *shipmentService) GetByClientOrderCode(ctx context.Context, client, clientOrderCode string) (*port.ShipmentDetail, error) {
	shipment, err := s.shipments.GetByClientOrderCode(ctx, client, strings.TrimSpace(clientOrderCode))
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, shipment)
}

func (s *shipmentService) List(ctx context.Context, client string, filter port.ShipmentFilter) (*port.ShipmentList, error) {
	filter.Limit, filter.Offset = normalizePage(filter.Limit, filter.Offset)
	items, total, err := s.shipments.List(ctx, client, filter)
	if err != nil {
		return nil, err
	}
	return &port.ShipmentList{Items: items, Total: total}, nil
}

func (s *shipmentService) Cancel(ctx context.Context, client string, id int64, reason string) (*model.Shipment, error) {
	shipment, err := s.owned(ctx, client, id)
	if err != nil {
		return nil, err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "cancelled by " + client
	}

	switch {
	case shipment.Status.CancellableLocally():
		now := s.clock.Now()
		acquired, err := s.shipments.AcquirePlacement(ctx, id, now, now.Add(s.settings.PlacementLease))
		if err != nil {
			return nil, err
		}
		if !acquired {
			return nil, fmt.Errorf("%w: shipment %d is being placed with the carrier, retry shortly", port.ErrConflict, id)
		}
		defer s.shipments.ReleasePlacement(context.WithoutCancel(ctx), id)
	case shipment.Status.CancellableAtCarrier():
		if err := s.carrier.CancelOrder(ctx, shipment.CarrierShopID, shipment.CarrierOrderCode); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: a shipment in status %s can no longer be cancelled", port.ErrInvalidTransition, shipment.Status)
	}

	updated, applied, err := s.lifecycle.apply(ctx, id, statusUpdate{
		To:            model.StatusCancelled,
		CarrierStatus: carrierCancelStatus(shipment),
		Reason:        reason,
		Source:        model.SourceAPI,
	})
	if err != nil {
		return nil, err
	}
	if !applied && updated.Status != model.StatusCancelled {
		return nil, fmt.Errorf("%w: shipment moved to %s before it could be cancelled", port.ErrInvalidTransition, updated.Status)
	}
	return updated, nil
}

func carrierCancelStatus(s *model.Shipment) string {
	if s.CarrierOrderCode != "" {
		return "cancel"
	}
	return ""
}

func (s *shipmentService) Sync(ctx context.Context, client string, id int64) (*model.Shipment, error) {
	shipment, err := s.owned(ctx, client, id)
	if err != nil {
		return nil, err
	}
	if shipment.CarrierOrderCode == "" {
		return nil, invalid("shipment %d has not been placed with the carrier yet", id)
	}
	order, err := s.carrier.GetOrder(ctx, shipment.CarrierShopID, shipment.CarrierOrderCode)
	if err != nil {
		return nil, err
	}
	status, ok := model.MapGHNStatus(order.Status)
	if !ok {
		return shipment, nil
	}
	upd := statusUpdate{To: status, CarrierStatus: order.Status, Source: model.SourceSync}
	if order.UpdatedAt != nil {
		upd.OccurredAt = *order.UpdatedAt
	}
	updated, _, err := s.lifecycle.apply(ctx, id, upd)
	return updated, err
}
