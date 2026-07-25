package http

import "gofr.dev/pkg/gofr"

func RegisterRoutes(app *gofr.App, h *Handler) {
	app.POST("/address-countries", h.CreateAddressCountry)
	app.GET("/address-countries", h.ListAddressCountries)
	app.GET("/address-countries/{id}", h.GetAddressCountry)
	app.PUT("/address-countries/{id}", h.UpdateAddressCountry)
	app.DELETE("/address-countries/{id}", h.DeleteAddressCountry)

	app.POST("/order-billing-addresses", h.CreateOrderBillingaddress)
	app.GET("/order-billing-addresses", h.ListOrderBillingaddresses)
	app.GET("/order-billing-addresses/{id}", h.GetOrderBillingaddress)
	app.PUT("/order-billing-addresses/{id}", h.UpdateOrderBillingaddress)
	app.DELETE("/order-billing-addresses/{id}", h.DeleteOrderBillingaddress)

	app.POST("/order-shipping-addresses", h.CreateOrderShippingaddress)
	app.GET("/order-shipping-addresses", h.ListOrderShippingaddresses)
	app.GET("/order-shipping-addresses/{id}", h.GetOrderShippingaddress)
	app.PUT("/order-shipping-addresses/{id}", h.UpdateOrderShippingaddress)
	app.DELETE("/order-shipping-addresses/{id}", h.DeleteOrderShippingaddress)

	app.POST("/partner-addresses", h.CreatePartneraddress)
	app.GET("/partner-addresses", h.ListPartneraddresses)
	app.GET("/partner-addresses/{id}", h.GetPartneraddress)
	app.PUT("/partner-addresses/{id}", h.UpdatePartneraddress)
	app.DELETE("/partner-addresses/{id}", h.DeletePartneraddress)

	app.POST("/shipping-orderanditemchanges-countries", h.CreateShippingOrderanditemchangesCountry)
	app.GET("/shipping-orderanditemchanges-countries", h.ListShippingOrderanditemchangesCountries)
	app.GET("/shipping-orderanditemchanges-countries/{id}", h.GetShippingOrderanditemchangesCountry)
	app.PUT("/shipping-orderanditemchanges-countries/{id}", h.UpdateShippingOrderanditemchangesCountry)
	app.DELETE("/shipping-orderanditemchanges-countries/{id}", h.DeleteShippingOrderanditemchangesCountry)

	app.POST("/shipping-weightbased-countries", h.CreateShippingWeightbasedCountry)
	app.GET("/shipping-weightbased-countries", h.ListShippingWeightbasedCountries)
	app.GET("/shipping-weightbased-countries/{id}", h.GetShippingWeightbasedCountry)
	app.PUT("/shipping-weightbased-countries/{id}", h.UpdateShippingWeightbasedCountry)
	app.DELETE("/shipping-weightbased-countries/{id}", h.DeleteShippingWeightbasedCountry)

	app.POST("/user-addresses", h.CreateUserAddress)
	app.GET("/user-addresses", h.ListUserAddresses)
	app.GET("/user-addresses/{id}", h.GetUserAddress)
	app.PUT("/user-addresses/{id}", h.UpdateUserAddress)
	app.DELETE("/user-addresses/{id}", h.DeleteUserAddress)
}
