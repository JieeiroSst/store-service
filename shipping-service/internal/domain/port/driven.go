package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
)

type Province struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type District struct {
	ID         int    `json:"id"`
	ProvinceID int    `json:"province_id"`
	Name       string `json:"name"`
}

type Ward struct {
	Code       string `json:"code"`
	DistrictID int    `json:"district_id"`
	Name       string `json:"name"`
}

type LocationDirectory interface {
	Provinces(ctx context.Context) ([]Province, error)
	Districts(ctx context.Context, provinceID int) ([]District, error)
	Wards(ctx context.Context, districtID int) ([]Ward, error)
}

type CarrierService struct {
	ID     int
	TypeID int
	Name   string
}

type CarrierQuoteRequest struct {
	ShopID         int64
	FromDistrictID int
	FromWardCode   string
	ToDistrictID   int
	ToWardCode     string
	ServiceID      int
	Parcel         model.Parcel
	CODAmount      int64
	InsuranceValue int64
}

type CarrierQuote struct {
	Fee                int64
	ExpectedDeliveryAt time.Time
}

type CarrierOrderRequest struct {
	ShopID          int64
	ClientOrderCode string
	ServiceID       int
	ServiceTypeID   int
	Recipient       model.Address
	Parcel          model.Parcel
	CODAmount       int64
	InsuranceValue  int64
	PaymentType     model.PaymentType
	RequiredNote    model.RequiredNote
	Note            string
}

type CarrierOrder struct {
	OrderCode          string
	Status             string
	Fee                int64
	ExpectedDeliveryAt *time.Time
	UpdatedAt          *time.Time
}

type Carrier interface {
	AvailableServices(ctx context.Context, shopID int64, fromDistrictID, toDistrictID int) ([]CarrierService, error)
	Quote(ctx context.Context, req CarrierQuoteRequest) (*CarrierQuote, error)
	CreateOrder(ctx context.Context, req CarrierOrderRequest) (*CarrierOrder, error)
	FindByClientOrderCode(ctx context.Context, shopID int64, clientOrderCode string) (*CarrierOrder, error)
	GetOrder(ctx context.Context, shopID int64, orderCode string) (*CarrierOrder, error)
	CancelOrder(ctx context.Context, shopID int64, orderCode string) error
	VerifyWebhookToken(token string) bool
}

type ShipmentRepository interface {
	Create(ctx context.Context, s *model.Shipment) error
	Save(ctx context.Context, s *model.Shipment) error
	GetByID(ctx context.Context, id int64) (*model.Shipment, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*model.Shipment, error)
	GetByClientOrderCode(ctx context.Context, client, clientOrderCode string) (*model.Shipment, error)
	GetByCode(ctx context.Context, code string) (*model.Shipment, error)
	GetByCarrierOrderCode(ctx context.Context, carrierOrderCode string) (*model.Shipment, error)
	List(ctx context.Context, client string, filter ShipmentFilter) ([]model.Shipment, int64, error)
	AcquirePlacement(ctx context.Context, id int64, now, until time.Time) (bool, error)
	ReleasePlacement(ctx context.Context, id int64) error
}

type EventRepository interface {
	Create(ctx context.Context, e *model.ShipmentEvent) (inserted bool, err error)
	ListByShipment(ctx context.Context, shipmentID int64) ([]model.ShipmentEvent, error)
}

type WarehouseRepository interface {
	Create(ctx context.Context, w *model.Warehouse) error
	Save(ctx context.Context, w *model.Warehouse) error
	GetByID(ctx context.Context, id int64) (*model.Warehouse, error)
	List(ctx context.Context) ([]model.Warehouse, error)
	ListActive(ctx context.Context, ids []int64) ([]model.Warehouse, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, jobs []model.OutboxJob) error
	ClaimDue(ctx context.Context, now time.Time, limit int, lease time.Duration) ([]model.OutboxJob, error)
	MarkDone(ctx context.Context, id int64) error
	MarkRetry(ctx context.Context, id int64, next time.Time, lastErr string) error
	MarkFailed(ctx context.Context, id int64, lastErr string) error
}

type Notifier interface {
	Notify(ctx context.Context, n model.Notification) error
}

type CallbackSender interface {
	Allowed(rawURL string) error
	Send(ctx context.Context, url string, event model.CallbackEvent) error
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type CodeGenerator interface {
	NewShipmentCode() string
}
