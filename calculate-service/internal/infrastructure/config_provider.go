package infrastructure

import (
	"encoding/json"
	"fmt"

	"github.com/JIeeiroSst/calculate-service/config"
	"github.com/JIeeiroSst/utils/consul"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	raw, err := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul).ConnectConfigConsul()
	if err != nil {
		return nil, fmt.Errorf("consul config unavailable: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("consul returned no config for key %q", dirEnv.KeyConsul)
	}

	var cfg config.Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse consul config: %w", err)
	}
	return &cfg, nil
}
