package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JIeeiroSst/bot-service/config"
	"github.com/JIeeiroSst/utils/consul"
)

func newConfig() (*config.Config, error) {
	host := os.Getenv("HostConsul")
	key := os.Getenv("KeyConsul")
	service := os.Getenv("ServiceConsul")

	if host == "" || key == "" || service == "" {
		return nil, fmt.Errorf("HostConsul/KeyConsul/ServiceConsul must be set")
	}

	raw, err := consul.NewConfigConsul(host, key, service).ConnectConfigConsul()
	if err != nil || raw == nil {
		return nil, fmt.Errorf("consul config unavailable: %w", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse consul config: %w", err)
	}
	return &cfg, nil
}
