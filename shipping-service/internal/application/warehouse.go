package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type warehouseService struct {
	warehouses port.WarehouseRepository
	directory  port.LocationDirectory
}

func NewWarehouseService(warehouses port.WarehouseRepository, directory port.LocationDirectory) port.WarehouseUsecase {
	return &warehouseService{warehouses: warehouses, directory: directory}
}

func (s *warehouseService) Create(ctx context.Context, in port.CreateWarehouseInput) (*model.Warehouse, error) {
	addr, err := model.Address{
		Name:       in.Name,
		Phone:      in.Phone,
		Street:     in.Street,
		WardCode:   in.WardCode,
		DistrictID: in.DistrictID,
		ProvinceID: in.ProvinceID,
	}.Normalize()
	if err != nil {
		return nil, invalid("%v", err)
	}
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	if code == "" {
		return nil, invalid("code is required")
	}
	if in.CarrierShopID <= 0 {
		return nil, invalid("carrier_shop_id is required")
	}
	if err := validateVNLocation(ctx, s.directory, addr.ProvinceID, addr.DistrictID, addr.WardCode); err != nil {
		return nil, err
	}

	w := &model.Warehouse{
		Code:          code,
		Name:          addr.Name,
		Phone:         addr.Phone,
		Street:        addr.Street,
		WardCode:      addr.WardCode,
		DistrictID:    addr.DistrictID,
		ProvinceID:    addr.ProvinceID,
		CarrierShopID: in.CarrierShopID,
		Active:        true,
	}
	if err := s.warehouses.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *warehouseService) List(ctx context.Context) ([]model.Warehouse, error) {
	return s.warehouses.List(ctx)
}

func (s *warehouseService) SetActive(ctx context.Context, id int64, active bool) (*model.Warehouse, error) {
	w, err := s.warehouses.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	w.Active = active
	if err := s.warehouses.Save(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}
