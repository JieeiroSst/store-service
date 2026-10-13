package http

import "gofr.dev/pkg/gofr"

func RegisterRoutes(app *gofr.App, h *Handler) {
	app.POST("/address-countries", admin(h.CreateAddressCountry))
	app.GET("/address-countries", h.ListAddressCountries)
	app.GET("/address-countries/{id}", h.GetAddressCountry)
	app.PUT("/address-countries/{id}", admin(h.UpdateAddressCountry))
	app.DELETE("/address-countries/{id}", admin(h.DeleteAddressCountry))

	app.POST("/order-billing-addresses", admin(h.CreateOrderBillingaddress))
	app.GET("/order-billing-addresses", admin(h.ListOrderBillingaddresses))
	app.GET("/order-billing-addresses/{id}", admin(h.GetOrderBillingaddress))
	app.PUT("/order-billing-addresses/{id}", admin(h.UpdateOrderBillingaddress))
	app.DELETE("/order-billing-addresses/{id}", admin(h.DeleteOrderBillingaddress))

	app.POST("/order-shipping-addresses", admin(h.CreateOrderShippingaddress))
	app.GET("/order-shipping-addresses", admin(h.ListOrderShippingaddresses))
	app.GET("/order-shipping-addresses/{id}", admin(h.GetOrderShippingaddress))
	app.PUT("/order-shipping-addresses/{id}", admin(h.UpdateOrderShippingaddress))
	app.DELETE("/order-shipping-addresses/{id}", admin(h.DeleteOrderShippingaddress))

	app.POST("/partner-addresses", admin(h.CreatePartneraddress))
	app.GET("/partner-addresses", admin(h.ListPartneraddresses))
	app.GET("/partner-addresses/{id}", admin(h.GetPartneraddress))
	app.PUT("/partner-addresses/{id}", admin(h.UpdatePartneraddress))
	app.DELETE("/partner-addresses/{id}", admin(h.DeletePartneraddress))

	app.POST("/shipping-orderanditemchanges-countries", admin(h.CreateShippingOrderanditemchangesCountry))
	app.GET("/shipping-orderanditemchanges-countries", admin(h.ListShippingOrderanditemchangesCountries))
	app.GET("/shipping-orderanditemchanges-countries/{id}", admin(h.GetShippingOrderanditemchangesCountry))
	app.PUT("/shipping-orderanditemchanges-countries/{id}", admin(h.UpdateShippingOrderanditemchangesCountry))
	app.DELETE("/shipping-orderanditemchanges-countries/{id}", admin(h.DeleteShippingOrderanditemchangesCountry))

	app.POST("/shipping-weightbased-countries", admin(h.CreateShippingWeightbasedCountry))
	app.GET("/shipping-weightbased-countries", admin(h.ListShippingWeightbasedCountries))
	app.GET("/shipping-weightbased-countries/{id}", admin(h.GetShippingWeightbasedCountry))
	app.PUT("/shipping-weightbased-countries/{id}", admin(h.UpdateShippingWeightbasedCountry))
	app.DELETE("/shipping-weightbased-countries/{id}", admin(h.DeleteShippingWeightbasedCountry))

	app.POST("/user-addresses", h.CreateUserAddress)
	app.GET("/user-addresses", h.ListUserAddresses)
	app.GET("/user-addresses/{id}", h.GetUserAddress)
	app.PUT("/user-addresses/{id}", h.UpdateUserAddress)
	app.DELETE("/user-addresses/{id}", h.DeleteUserAddress)
}
