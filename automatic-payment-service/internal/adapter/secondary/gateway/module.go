package gateway

import "go.uber.org/fx"

// Module wires port.PaymentGatewayPort to IntegratedPaymentGateway, which
// delegates charges to integrated-payment-service over HTTP. Swap in
// NewMockGateway for local development/tests that don't have that service
// running.
var Module = fx.Options(
	fx.Provide(NewIntegratedPaymentGateway),
)
