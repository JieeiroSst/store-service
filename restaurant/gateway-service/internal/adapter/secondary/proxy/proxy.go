package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/JIeeiroSst/gateway-service/config"
	"github.com/JIeeiroSst/gateway-service/internal/domain/port"
)

func newReverseProxy(baseURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(target), nil
}

func NewConsumerProxy(cfg *config.Config) (port.ConsumerProxy, error) {
	return newReverseProxy(cfg.Gateway.ConsumerURL)
}

func NewAccountingProxy(cfg *config.Config) (port.AccountingProxy, error) {
	return newReverseProxy(cfg.Gateway.AccountingURL)
}

func NewDeliveryProxy(cfg *config.Config) (port.DeliveryProxy, error) {
	return newReverseProxy(cfg.Gateway.DeliveryURL)
}
