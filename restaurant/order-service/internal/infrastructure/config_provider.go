package infrastructure

import (
	"encoding/json"

	"github.com/JIeeiroSst/order-service/config"
	"github.com/JieeiroSst/logger"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	consul := logger.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)

	data, err := consul.ConnectConfigConsul()
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
