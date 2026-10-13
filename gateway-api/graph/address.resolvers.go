package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) AddressCreateCountry(ctx context.Context, input model.AddressCountryInput) (*model.AddressCountry, error) {
	return restCall[*model.AddressCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "POST", Path: "/address-countries", Body: input}, "data")
}

func (r *mutationResolver) AddressUpdateCountry(ctx context.Context, id int, input model.AddressCountryInput) (*model.AddressCountry, error) {
	return restCall[*model.AddressCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "PUT", Path: rest.P("/address-countries/%v", id), Body: input}, "data")
}

func (r *mutationResolver) AddressDeleteCountry(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "address-country", rest.Request{Method: "DELETE", Path: rest.P("/address-countries/%v", id)}, "")
}

func (r *mutationResolver) AddressCreateOrderBillingAddress(ctx context.Context, input model.AddressOrderBillingAddressInput) (*model.AddressOrderBillingAddress, error) {
	return restCall[*model.AddressOrderBillingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "POST", Path: "/order-billing-addresses", Body: input}, "data")
}

func (r *mutationResolver) AddressUpdateOrderBillingAddress(ctx context.Context, id int, input model.AddressOrderBillingAddressInput) (*model.AddressOrderBillingAddress, error) {
	return restCall[*model.AddressOrderBillingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "PUT", Path: rest.P("/order-billing-addresses/%v", id), Body: input}, "data")
}

func (r *mutationResolver) AddressDeleteOrderBillingAddress(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "address-country", rest.Request{Method: "DELETE", Path: rest.P("/order-billing-addresses/%v", id)}, "")
}

func (r *mutationResolver) AddressCreateOrderShippingAddress(ctx context.Context, input model.AddressOrderShippingAddressInput) (*model.AddressOrderShippingAddress, error) {
	return restCall[*model.AddressOrderShippingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "POST", Path: "/order-shipping-addresses", Body: input}, "data")
}

func (r *mutationResolver) AddressUpdateOrderShippingAddress(ctx context.Context, id int, input model.AddressOrderShippingAddressInput) (*model.AddressOrderShippingAddress, error) {
	return restCall[*model.AddressOrderShippingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "PUT", Path: rest.P("/order-shipping-addresses/%v", id), Body: input}, "data")
}

func (r *mutationResolver) AddressDeleteOrderShippingAddress(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "address-country", rest.Request{Method: "DELETE", Path: rest.P("/order-shipping-addresses/%v", id)}, "")
}

func (r *mutationResolver) AddressCreatePartnerAddress(ctx context.Context, input model.AddressPartnerAddressInput) (*model.AddressPartnerAddress, error) {
	return restCall[*model.AddressPartnerAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "POST", Path: "/partner-addresses", Body: input}, "data")
}

func (r *mutationResolver) AddressUpdatePartnerAddress(ctx context.Context, id int, input model.AddressPartnerAddressInput) (*model.AddressPartnerAddress, error) {
	return restCall[*model.AddressPartnerAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "PUT", Path: rest.P("/partner-addresses/%v", id), Body: input}, "data")
}

func (r *mutationResolver) AddressDeletePartnerAddress(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "address-country", rest.Request{Method: "DELETE", Path: rest.P("/partner-addresses/%v", id)}, "")
}

func (r *mutationResolver) AddressCreateShippingOrderAndItemChangesCountry(ctx context.Context, input model.AddressShippingOrderAndItemChangesCountryInput) (*model.AddressShippingOrderAndItemChangesCountry, error) {
	return restCall[*model.AddressShippingOrderAndItemChangesCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "POST", Path: "/shipping-orderanditemchanges-countries", Body: input}, "data")
}

func (r *mutationResolver) AddressUpdateShippingOrderAndItemChangesCountry(ctx context.Context, id int, input model.AddressShippingOrderAndItemChangesCountryInput) (*model.AddressShippingOrderAndItemChangesCountry, error) {
	return restCall[*model.AddressShippingOrderAndItemChangesCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "PUT", Path: rest.P("/shipping-orderanditemchanges-countries/%v", id), Body: input}, "data")
}

func (r *mutationResolver) AddressDeleteShippingOrderAndItemChangesCountry(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "address-country", rest.Request{Method: "DELETE", Path: rest.P("/shipping-orderanditemchanges-countries/%v", id)}, "")
}

func (r *mutationResolver) AddressCreateShippingWeightBasedCountry(ctx context.Context, input model.AddressShippingWeightBasedCountryInput) (*model.AddressShippingWeightBasedCountry, error) {
	return restCall[*model.AddressShippingWeightBasedCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "POST", Path: "/shipping-weightbased-countries", Body: input}, "data")
}

func (r *mutationResolver) AddressUpdateShippingWeightBasedCountry(ctx context.Context, id int, input model.AddressShippingWeightBasedCountryInput) (*model.AddressShippingWeightBasedCountry, error) {
	return restCall[*model.AddressShippingWeightBasedCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "PUT", Path: rest.P("/shipping-weightbased-countries/%v", id), Body: input}, "data")
}

func (r *mutationResolver) AddressDeleteShippingWeightBasedCountry(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "address-country", rest.Request{Method: "DELETE", Path: rest.P("/shipping-weightbased-countries/%v", id)}, "")
}

func (r *mutationResolver) AddressCreateUserAddress(ctx context.Context, input model.AddressAdminUserAddressInput) (*model.AddressUserAddress, error) {
	return restCall[*model.AddressUserAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "POST", Path: "/user-addresses", Body: input}, "data")
}

func (r *mutationResolver) AddressUpdateUserAddress(ctx context.Context, id int, input model.AddressAdminUserAddressInput) (*model.AddressUserAddress, error) {
	return restCall[*model.AddressUserAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "PUT", Path: rest.P("/user-addresses/%v", id), Body: input}, "data")
}

func (r *mutationResolver) AddressDeleteUserAddress(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "address-country", rest.Request{Method: "DELETE", Path: rest.P("/user-addresses/%v", id)}, "")
}

func (r *mutationResolver) CreateMyUserAddress(ctx context.Context, input model.AddressUserAddressInput) (*model.AddressUserAddress, error) {
	return r.createMyUserAddress(ctx, input)
}

func (r *mutationResolver) UpdateMyUserAddress(ctx context.Context, id int, input model.AddressUserAddressInput) (*model.AddressUserAddress, error) {
	return r.updateMyUserAddress(ctx, id, input)
}

func (r *mutationResolver) DeleteMyUserAddress(ctx context.Context, id int) (bool, error) {
	return r.deleteMyUserAddress(ctx, id)
}

func (r *queryResolver) AddressCountries(ctx context.Context) ([]*model.AddressCountry, error) {
	return restCall[[]*model.AddressCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: "/address-countries"}, "data")
}

func (r *queryResolver) AddressCountry(ctx context.Context, id int) (*model.AddressCountry, error) {
	return restCall[*model.AddressCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: rest.P("/address-countries/%v", id)}, "data")
}

func (r *queryResolver) AddressOrderBillingAddresses(ctx context.Context) ([]*model.AddressOrderBillingAddress, error) {
	return restCall[[]*model.AddressOrderBillingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: "/order-billing-addresses"}, "data")
}

func (r *queryResolver) AddressOrderBillingAddress(ctx context.Context, id int) (*model.AddressOrderBillingAddress, error) {
	return restCall[*model.AddressOrderBillingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: rest.P("/order-billing-addresses/%v", id)}, "data")
}

func (r *queryResolver) AddressOrderShippingAddresses(ctx context.Context) ([]*model.AddressOrderShippingAddress, error) {
	return restCall[[]*model.AddressOrderShippingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: "/order-shipping-addresses"}, "data")
}

func (r *queryResolver) AddressOrderShippingAddress(ctx context.Context, id int) (*model.AddressOrderShippingAddress, error) {
	return restCall[*model.AddressOrderShippingAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: rest.P("/order-shipping-addresses/%v", id)}, "data")
}

func (r *queryResolver) AddressPartnerAddresses(ctx context.Context) ([]*model.AddressPartnerAddress, error) {
	return restCall[[]*model.AddressPartnerAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: "/partner-addresses"}, "data")
}

func (r *queryResolver) AddressPartnerAddress(ctx context.Context, id int) (*model.AddressPartnerAddress, error) {
	return restCall[*model.AddressPartnerAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: rest.P("/partner-addresses/%v", id)}, "data")
}

func (r *queryResolver) AddressShippingOrderAndItemChangesCountries(ctx context.Context) ([]*model.AddressShippingOrderAndItemChangesCountry, error) {
	return restCall[[]*model.AddressShippingOrderAndItemChangesCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: "/shipping-orderanditemchanges-countries"}, "data")
}

func (r *queryResolver) AddressShippingOrderAndItemChangesCountry(ctx context.Context, id int) (*model.AddressShippingOrderAndItemChangesCountry, error) {
	return restCall[*model.AddressShippingOrderAndItemChangesCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: rest.P("/shipping-orderanditemchanges-countries/%v", id)}, "data")
}

func (r *queryResolver) AddressShippingWeightBasedCountries(ctx context.Context) ([]*model.AddressShippingWeightBasedCountry, error) {
	return restCall[[]*model.AddressShippingWeightBasedCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: "/shipping-weightbased-countries"}, "data")
}

func (r *queryResolver) AddressShippingWeightBasedCountry(ctx context.Context, id int) (*model.AddressShippingWeightBasedCountry, error) {
	return restCall[*model.AddressShippingWeightBasedCountry](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: rest.P("/shipping-weightbased-countries/%v", id)}, "data")
}

func (r *queryResolver) AddressUserAddresses(ctx context.Context) ([]*model.AddressUserAddress, error) {
	return restCall[[]*model.AddressUserAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: "/user-addresses"}, "data")
}

func (r *queryResolver) AddressUserAddress(ctx context.Context, id int) (*model.AddressUserAddress, error) {
	return restCall[*model.AddressUserAddress](ctx, r.Proxy, "address-country", rest.Request{Method: "GET", Path: rest.P("/user-addresses/%v", id)}, "data")
}

func (r *queryResolver) MyUserAddresses(ctx context.Context) ([]*model.AddressUserAddress, error) {
	return r.myUserAddresses(ctx)
}
