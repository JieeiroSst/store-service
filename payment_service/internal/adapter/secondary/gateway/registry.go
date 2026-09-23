package gateway

import (
	"fmt"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	Gateways []port.PaymentGateway `group:"payment_gateways"`
}

type resolver struct {
	byProvider map[model.Provider]port.PaymentGateway
}

func NewResolver(p Params) port.PaymentGatewayResolver {
	byProvider := make(map[model.Provider]port.PaymentGateway, len(p.Gateways))
	for _, g := range p.Gateways {
		byProvider[g.Provider()] = g
	}
	return &resolver{byProvider: byProvider}
}

func (r *resolver) Resolve(provider model.Provider) (port.PaymentGateway, error) {
	gateway, ok := r.byProvider[provider]
	if !ok {
		return nil, fmt.Errorf("%w: %s", port.ErrUnsupportedProvider, provider)
	}
	return gateway, nil
}

var Module = fx.Options(
	fx.Provide(NewResolver),
)
