package port

import (
	"context"

	"github.com/JIeeiroSst/address-country-service/internal/domain/model"
)

type AddressCountryRepository interface {
	Create(ctx context.Context, address *model.AddressCountry) error
	GetByID(ctx context.Context, id string) (*model.AddressCountry, error)
	Update(ctx context.Context, address *model.AddressCountry) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]model.AddressCountry, error)
}

type OrderBillingaddressRepository interface {
	Create(ctx context.Context, address *model.OrderBillingaddress) error
	GetByID(ctx context.Context, id int) (*model.OrderBillingaddress, error)
	Update(ctx context.Context, address *model.OrderBillingaddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.OrderBillingaddress, error)
}

type OrderShippingaddressRepository interface {
	Create(ctx context.Context, address *model.OrderShippingaddress) error
	GetByID(ctx context.Context, id int) (*model.OrderShippingaddress, error)
	Update(ctx context.Context, address *model.OrderShippingaddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.OrderShippingaddress, error)
}

type PartneraddressRepository interface {
	Create(ctx context.Context, address *model.Partneraddress) error
	GetByID(ctx context.Context, id int) (*model.Partneraddress, error)
	Update(ctx context.Context, address *model.Partneraddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Partneraddress, error)
}

type ShippingOrderanditemchangesCountryRepository interface {
	Create(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error
	GetByID(ctx context.Context, id int) (*model.ShippingOrderanditemchangesCountry, error)
	Update(ctx context.Context, entry *model.ShippingOrderanditemchangesCountry) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.ShippingOrderanditemchangesCountry, error)
}

type ShippingWeightbasedCountryRepository interface {
	Create(ctx context.Context, entry *model.ShippingWeightbasedCountry) error
	GetByID(ctx context.Context, id int) (*model.ShippingWeightbasedCountry, error)
	Update(ctx context.Context, entry *model.ShippingWeightbasedCountry) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.ShippingWeightbasedCountry, error)
}

type UserAddressRepository interface {
	Create(ctx context.Context, address *model.UserAddress) error
	GetByID(ctx context.Context, id int) (*model.UserAddress, error)
	Update(ctx context.Context, address *model.UserAddress) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.UserAddress, error)
}
