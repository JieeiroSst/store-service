package port

import (
	"context"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
)

type QuoteInput struct {
	Recipient      model.Address
	Parcel         model.Parcel
	CODAmount      int64
	InsuranceValue int64
	WarehouseIDs   []int64
}

type QuoteResult struct {
	Options     []model.RouteOption                   `json:"options"`
	Recommended map[model.Strategy]*model.RouteOption `json:"recommended"`
	Unavailable []model.UnavailableRoute              `json:"unavailable"`
}

type CreateShipmentInput struct {
	ClientOrderCode string
	Recipient       model.Address
	Parcel          model.Parcel
	CODAmount       int64
	InsuranceValue  int64
	PaymentType     model.PaymentType
	RequiredNote    model.RequiredNote
	Note            string
	Strategy        model.Strategy
	WarehouseIDs    []int64
	WarehouseID     int64
	ServiceID       int
	Customer        model.Customer
	CallbackURL     string
}

type ShipmentFilter struct {
	Status model.Status
	Limit  int
	Offset int
}

type ShipmentList struct {
	Items []model.Shipment `json:"items"`
	Total int64            `json:"total"`
}

type ShipmentDetail struct {
	Shipment model.Shipment        `json:"shipment"`
	Events   []model.ShipmentEvent `json:"events"`
}

type ShipmentUsecase interface {
	Quote(ctx context.Context, in QuoteInput) (*QuoteResult, error)
	Create(ctx context.Context, client string, in CreateShipmentInput) (shipment *model.Shipment, created bool, err error)
	Place(ctx context.Context, client string, id int64) (*model.Shipment, error)
	Get(ctx context.Context, client string, id int64) (*ShipmentDetail, error)
	GetByClientOrderCode(ctx context.Context, client, clientOrderCode string) (*ShipmentDetail, error)
	List(ctx context.Context, client string, filter ShipmentFilter) (*ShipmentList, error)
	Cancel(ctx context.Context, client string, id int64, reason string) (*model.Shipment, error)
	Sync(ctx context.Context, client string, id int64) (*model.Shipment, error)
}

type CarrierWebhook struct {
	OrderCode       string
	ClientOrderCode string
	Type            string
	Status          string
	Reason          string
	Time            string
}

type WebhookUsecase interface {
	HandleGHN(ctx context.Context, token string, event CarrierWebhook) error
}

type CreateWarehouseInput struct {
	Code          string
	Name          string
	Phone         string
	Street        string
	WardCode      string
	DistrictID    int
	ProvinceID    int
	CarrierShopID int64
}

type WarehouseUsecase interface {
	Create(ctx context.Context, in CreateWarehouseInput) (*model.Warehouse, error)
	List(ctx context.Context) ([]model.Warehouse, error)
	SetActive(ctx context.Context, id int64, active bool) (*model.Warehouse, error)
}

type LocationUsecase interface {
	Provinces(ctx context.Context) ([]Province, error)
	Districts(ctx context.Context, provinceID int) ([]District, error)
	Wards(ctx context.Context, districtID int) ([]Ward, error)
}

type OutboxUsecase interface {
	DispatchDue(ctx context.Context) (int, error)
}
