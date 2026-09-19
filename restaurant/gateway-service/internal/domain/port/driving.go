package port

import "net/http"

type GatewayUsecase interface {
	ConsumerHandler() http.Handler
	AccountingHandler() http.Handler
	DeliveryHandler() http.Handler
}
