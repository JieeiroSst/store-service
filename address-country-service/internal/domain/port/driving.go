package port

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
)

type AddressCountryUsecase interface {
	Create(ctx context.Context, address *model.AddressCountry) error
	Get(ctx context.Context, id string) (*model.AddressCountry, error)
	Update(ctx context.Context, address *model.AddressCountry) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]model.AddressCountry, error)
}

type OrderBillingaddressUsecase interface {
	Create(ctx context.Context, address *model.OrderBillingaddress) error
	Get(ctx context.Context, id int) (*model.OrderBillingaddress, error)
	Update(ctx context.Context, address *model.OrderBillingaddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.OrderBillingaddress, error)
}

type OrderShippingaddressUsecase interface {
	Create(ctx context.Context, address *model.OrderShippingaddress) error
	Get(ctx context.Context, id int) (*model.OrderShippingaddress, error)
	Update(ctx context.Context, address *model.OrderShippingaddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.OrderShippingaddress, error)
}

type PartneraddressUsecase interface {
	Create(ctx context.Context, address *model.Partneraddress) error
	Get(ctx context.Context, id int) (*model.Partneraddress, error)
	Update(ctx context.Context, address *model.Partneraddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Partneraddress, error)
}

type ShippingOrderanditemchangesCountryUsecase interface {
	Create(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error
	Get(ctx context.Context, id int) (*model.ShippingOrderanditemchangesCountry, error)
	Update(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.ShippingOrderanditemchangesCountry, error)
}

type ShippingWeightbasedCountryUsecase interface {
	Create(ctx context.Context, entry *model.ShippingWeightbasedCountry) error
	Get(ctx context.Context, id int) (*model.ShippingWeightbasedCountry, error)
	Update(ctx context.Context, entry *model.ShippingWeightbasedCountry) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.ShippingWeightbasedCountry, error)
}

type UserAddressUsecase interface {
	Create(ctx context.Context, address *model.UserAddress) error
	Get(ctx context.Context, id int) (*model.UserAddress, error)
	Update(ctx context.Context, address *model.UserAddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.UserAddress, error)
}
