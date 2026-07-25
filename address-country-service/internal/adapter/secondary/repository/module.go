package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAddressCountryRepository),
	fx.Provide(NewOrderBillingaddressRepository),
	fx.Provide(NewOrderShippingaddressRepository),
	fx.Provide(NewPartneraddressRepository),
	fx.Provide(NewShippingOrderanditemchangesCountryRepository),
	fx.Provide(NewShippingWeightbasedCountryRepository),
	fx.Provide(NewUserAddressRepository),
)
