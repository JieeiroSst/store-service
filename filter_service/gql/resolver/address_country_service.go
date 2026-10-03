package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) AddrQuery() generated.AddrQueryResolver { return &addrQueryResolver{r} }

type addrQueryResolver struct{ *Resolver }

func (r *addrQueryResolver) AddressCountries(ctx context.Context, obj *model.AddrQuery) ([]*model.AddrAddressCountry, error) {
	return r.Clients.AddressCountryService.AddressCountries(ctx)
}

func (r *addrQueryResolver) AddressCountry(ctx context.Context, obj *model.AddrQuery, id string) (*model.AddrAddressCountry, error) {
	return r.Clients.AddressCountryService.AddressCountry(ctx, id)
}

func (r *addrQueryResolver) OrderBillingAddresses(ctx context.Context, obj *model.AddrQuery) ([]*model.AddrOrderBillingaddress, error) {
	return r.Clients.AddressCountryService.OrderBillingAddresses(ctx)
}

func (r *addrQueryResolver) OrderBillingAddress(ctx context.Context, obj *model.AddrQuery, id string) (*model.AddrOrderBillingaddress, error) {
	return r.Clients.AddressCountryService.OrderBillingAddress(ctx, id)
}

func (r *addrQueryResolver) OrderShippingAddresses(ctx context.Context, obj *model.AddrQuery) ([]*model.AddrOrderShippingaddress, error) {
	return r.Clients.AddressCountryService.OrderShippingAddresses(ctx)
}

func (r *addrQueryResolver) OrderShippingAddress(ctx context.Context, obj *model.AddrQuery, id string) (*model.AddrOrderShippingaddress, error) {
	return r.Clients.AddressCountryService.OrderShippingAddress(ctx, id)
}

func (r *addrQueryResolver) PartnerAddresses(ctx context.Context, obj *model.AddrQuery) ([]*model.AddrPartneraddress, error) {
	return r.Clients.AddressCountryService.PartnerAddresses(ctx)
}

func (r *addrQueryResolver) PartnerAddress(ctx context.Context, obj *model.AddrQuery, id string) (*model.AddrPartneraddress, error) {
	return r.Clients.AddressCountryService.PartnerAddress(ctx, id)
}

func (r *addrQueryResolver) ShippingOrderAndItemChangesCountries(ctx context.Context, obj *model.AddrQuery) ([]*model.AddrShippingOrderanditemchangesCountry, error) {
	return r.Clients.AddressCountryService.ShippingOrderAndItemChangesCountries(ctx)
}

func (r *addrQueryResolver) ShippingOrderAndItemChangesCountry(ctx context.Context, obj *model.AddrQuery, id string) (*model.AddrShippingOrderanditemchangesCountry, error) {
	return r.Clients.AddressCountryService.ShippingOrderAndItemChangesCountry(ctx, id)
}

func (r *addrQueryResolver) ShippingWeightBasedCountries(ctx context.Context, obj *model.AddrQuery) ([]*model.AddrShippingWeightbasedCountry, error) {
	return r.Clients.AddressCountryService.ShippingWeightBasedCountries(ctx)
}

func (r *addrQueryResolver) ShippingWeightBasedCountry(ctx context.Context, obj *model.AddrQuery, id string) (*model.AddrShippingWeightbasedCountry, error) {
	return r.Clients.AddressCountryService.ShippingWeightBasedCountry(ctx, id)
}

func (r *addrQueryResolver) UserAddresses(ctx context.Context, obj *model.AddrQuery) ([]*model.AddrUserAddress, error) {
	return r.Clients.AddressCountryService.UserAddresses(ctx)
}

func (r *addrQueryResolver) UserAddress(ctx context.Context, obj *model.AddrQuery, id string) (*model.AddrUserAddress, error) {
	return r.Clients.AddressCountryService.UserAddress(ctx, id)
}
