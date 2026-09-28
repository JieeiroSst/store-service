package application

import (
	"context"
	"errors"
	"sync"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type planner struct {
	carrier    port.Carrier
	directory  port.LocationDirectory
	warehouses port.WarehouseRepository
	settings   Settings
}

type quoteTask struct {
	warehouse model.Warehouse
	service   port.CarrierService
}

func (p *planner) validate(ctx context.Context, in *port.QuoteInput) error {
	recipient, err := in.Recipient.Normalize()
	if err != nil {
		return invalid("recipient: %v", err)
	}
	in.Recipient = recipient
	if err := in.Parcel.Validate(); err != nil {
		return invalid("%v", err)
	}
	if in.CODAmount < 0 || in.InsuranceValue < 0 {
		return invalid("cod_amount and insurance_value must not be negative")
	}
	if in.CODAmount > p.settings.MaxCODAmount {
		return invalid("cod_amount must not exceed %d VND", p.settings.MaxCODAmount)
	}
	return validateVNLocation(ctx, p.directory, in.Recipient.ProvinceID, in.Recipient.DistrictID, in.Recipient.WardCode)
}

func (p *planner) plan(ctx context.Context, in port.QuoteInput) ([]model.RouteOption, []model.UnavailableRoute, error) {
	warehouses, err := p.warehouses.ListActive(ctx, in.WarehouseIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(warehouses) == 0 {
		return nil, nil, invalid("no active warehouse to ship from")
	}

	var (
		tasks       []quoteTask
		unavailable []model.UnavailableRoute
	)
	for _, w := range warehouses {
		services, err := p.carrier.AvailableServices(ctx, w.CarrierShopID, w.DistrictID, in.Recipient.DistrictID)
		if err != nil {
			if !isCarrierError(err) {
				return nil, nil, err
			}
			unavailable = append(unavailable, model.UnavailableRoute{WarehouseCode: w.Code, Reason: err.Error()})
			continue
		}
		if len(services) == 0 {
			unavailable = append(unavailable, model.UnavailableRoute{WarehouseCode: w.Code, Reason: "carrier has no service for this route"})
		}
		for _, svc := range services {
			tasks = append(tasks, quoteTask{warehouse: w, service: svc})
		}
	}

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		options []model.RouteOption
		fatal   error
		sem     = make(chan struct{}, max(1, p.settings.QuoteConcurrency))
	)
	for _, t := range tasks {
		wg.Add(1)
		sem <- struct{}{}
		go func(t quoteTask) {
			defer wg.Done()
			defer func() { <-sem }()
			q, err := p.carrier.Quote(ctx, port.CarrierQuoteRequest{
				ShopID:         t.warehouse.CarrierShopID,
				FromDistrictID: t.warehouse.DistrictID,
				FromWardCode:   t.warehouse.WardCode,
				ToDistrictID:   in.Recipient.DistrictID,
				ToWardCode:     in.Recipient.WardCode,
				ServiceID:      t.service.ID,
				Parcel:         in.Parcel,
				CODAmount:      in.CODAmount,
				InsuranceValue: in.InsuranceValue,
			})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				options = append(options, model.RouteOption{
					WarehouseID:        t.warehouse.ID,
					WarehouseCode:      t.warehouse.Code,
					CarrierShopID:      t.warehouse.CarrierShopID,
					ServiceID:          t.service.ID,
					ServiceTypeID:      t.service.TypeID,
					ServiceName:        t.service.Name,
					Fee:                q.Fee,
					ExpectedDeliveryAt: q.ExpectedDeliveryAt,
				})
			case isCarrierError(err):
				unavailable = append(unavailable, model.UnavailableRoute{WarehouseCode: t.warehouse.Code, ServiceID: t.service.ID, Reason: err.Error()})
			default:
				fatal = err
			}
		}(t)
	}
	wg.Wait()
	if fatal != nil {
		return nil, nil, fatal
	}
	return options, unavailable, nil
}

func isCarrierError(err error) bool {
	return errors.Is(err, port.ErrCarrierRejected) || errors.Is(err, port.ErrCarrierUnavailable)
}
