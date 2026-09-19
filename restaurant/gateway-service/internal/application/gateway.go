package application

import (
	"net/http"

	"github.com/JIeeiroSst/gateway-service/internal/domain/port"
)

type gatewayService struct {
	consumer   port.ConsumerProxy
	accounting port.AccountingProxy
	delivery   port.DeliveryProxy
}

func NewGatewayService(consumer port.ConsumerProxy, accounting port.AccountingProxy, delivery port.DeliveryProxy) port.GatewayUsecase {
	return &gatewayService{
		consumer:   consumer,
		accounting: accounting,
		delivery:   delivery,
	}
}

func (s *gatewayService) ConsumerHandler() http.Handler   { return s.consumer }
func (s *gatewayService) AccountingHandler() http.Handler { return s.accounting }
func (s *gatewayService) DeliveryHandler() http.Handler   { return s.delivery }
