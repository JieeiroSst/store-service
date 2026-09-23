package paypal

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(fx.Annotate(NewGateway, fx.ResultTags(`group:"payment_gateways"`))),
)
