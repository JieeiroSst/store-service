package http

import "github.com/JIeeiroSst/address-country-service/internal/domain/port"

type Handler struct {
	addressCountry                     port.AddressCountryUsecase
	orderBillingaddress                port.OrderBillingaddressUsecase
	orderShippingaddress               port.OrderShippingaddressUsecase
	partneraddress                     port.PartneraddressUsecase
	shippingOrderanditemchangesCountry port.ShippingOrderanditemchangesCountryUsecase
	shippingWeightbasedCountry         port.ShippingWeightbasedCountryUsecase
	userAddress                        port.UserAddressUsecase
}

func NewHandler(
	addressCountry port.AddressCountryUsecase,
	orderBillingaddress port.OrderBillingaddressUsecase,
	orderShippingaddress port.OrderShippingaddressUsecase,
	partneraddress port.PartneraddressUsecase,
	shippingOrderanditemchangesCountry port.ShippingOrderanditemchangesCountryUsecase,
	shippingWeightbasedCountry port.ShippingWeightbasedCountryUsecase,
	userAddress port.UserAddressUsecase,
) *Handler {
	return &Handler{
		addressCountry:                     addressCountry,
		orderBillingaddress:                orderBillingaddress,
		orderShippingaddress:               orderShippingaddress,
		partneraddress:                     partneraddress,
		shippingOrderanditemchangesCountry: shippingOrderanditemchangesCountry,
		shippingWeightbasedCountry:         shippingWeightbasedCountry,
		userAddress:                        userAddress,
	}
}
