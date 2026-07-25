package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAddressCountryService),
	fx.Provide(NewOrderBillingaddressService),
	fx.Provide(NewOrderShippingaddressService),
	fx.Provide(NewPartneraddressService),
	fx.Provide(NewShippingOrderanditemchangesCountryService),
	fx.Provide(NewShippingWeightbasedCountryService),
	fx.Provide(NewUserAddressService),
)
