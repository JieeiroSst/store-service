package infrastructure

import (
	"os"

	"github.com/JIeeiroSst/basket-service/config"
	"github.com/JIeeiroSst/basket-service/pkg/consul"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	cfg, err := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul).ConnectConfigConsul()
	if err != nil {
		return nil, err
	}
	if key := os.Getenv("SHIPPING_SERVICE_API_KEY"); key != "" {
		cfg.ShippingService.APIKey = key
	}
	return cfg, nil
}
