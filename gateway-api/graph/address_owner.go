package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

const addressService = "address-country"

func (r *Resolver) myUserAddresses(ctx context.Context) ([]*model.AddressUserAddress, error) {
	uid, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	all, err := restCall[[]*model.AddressUserAddress](ctx, r.Proxy, addressService, rest.Request{Method: "GET", Path: "/user-addresses"}, "data")
	if err != nil {
		return nil, err
	}
	out := []*model.AddressUserAddress{}
	for _, a := range all {
		if a != nil && a.UserID == uid {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *Resolver) ownedUserAddress(ctx context.Context, id int) (int, error) {
	uid, err := currentUserID(ctx)
	if err != nil {
		return 0, err
	}
	existing, err := restCall[*model.AddressUserAddress](ctx, r.Proxy, addressService, rest.Request{Method: "GET", Path: rest.P("/user-addresses/%v", id)}, "data")
	if err != nil {
		return 0, err
	}
	if existing == nil || existing.UserID != uid {
		return 0, notOwned(ctx)
	}
	return uid, nil
}

func (r *Resolver) createMyUserAddress(ctx context.Context, input model.AddressUserAddressInput) (*model.AddressUserAddress, error) {
	uid, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	return restCall[*model.AddressUserAddress](ctx, r.Proxy, addressService, rest.Request{Method: "POST", Path: "/user-addresses", Body: withField(input, "user_id", uid)}, "data")
}

func (r *Resolver) updateMyUserAddress(ctx context.Context, id int, input model.AddressUserAddressInput) (*model.AddressUserAddress, error) {
	uid, err := r.ownedUserAddress(ctx, id)
	if err != nil {
		return nil, err
	}
	return restCall[*model.AddressUserAddress](ctx, r.Proxy, addressService, rest.Request{Method: "PUT", Path: rest.P("/user-addresses/%v", id), Body: withField(input, "user_id", uid)}, "data")
}

func (r *Resolver) deleteMyUserAddress(ctx context.Context, id int) (bool, error) {
	if _, err := r.ownedUserAddress(ctx, id); err != nil {
		return false, err
	}
	return restCall[bool](ctx, r.Proxy, addressService, rest.Request{Method: "DELETE", Path: rest.P("/user-addresses/%v", id)}, "")
}
