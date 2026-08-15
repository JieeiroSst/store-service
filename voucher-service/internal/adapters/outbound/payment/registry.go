package payment

import (
	"fmt"

	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
)

type Registry struct {
	byProvider map[string]paymentapp.PaymentGateway
}

func NewRegistry(gateways []paymentapp.PaymentGateway) paymentapp.PaymentGatewayRegistry {
	byProvider := make(map[string]paymentapp.PaymentGateway, len(gateways))
	for _, g := range gateways {
		byProvider[g.Provider()] = g
	}
	return &Registry{byProvider: byProvider}
}

func (r *Registry) Resolve(provider string) (paymentapp.PaymentGateway, error) {
	g, ok := r.byProvider[provider]
	if !ok {
		return nil, fmt.Errorf("no payment gateway registered for provider %q", provider)
	}
	return g, nil
}
