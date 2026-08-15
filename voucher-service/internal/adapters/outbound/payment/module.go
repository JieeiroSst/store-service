package payment

import (
	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	"go.uber.org/fx"
)

var Module = fx.Module("payment-gateways",
	fx.Provide(
		fx.Annotate(NewVNPayGateway, fx.As(new(paymentapp.PaymentGateway)), fx.ResultTags(`group:"gateways"`)),
		fx.Annotate(NewMomoGateway, fx.As(new(paymentapp.PaymentGateway)), fx.ResultTags(`group:"gateways"`)),
		fx.Annotate(NewRegistry, fx.ParamTags(`group:"gateways"`), fx.As(new(paymentapp.PaymentGatewayRegistry))),
	),
)
