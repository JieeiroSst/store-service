package config

import (
	"encoding/json"
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
)

func loadFromConsul(host, key, service string) (*Config, error) {
	clientCfg := consulapi.DefaultConfig()
	clientCfg.Address = host
	client, err := consulapi.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to consul: %w", err)
	}

	if _, _, err := client.Catalog().Service(service, "", nil); err != nil {
		return nil, fmt.Errorf("lookup consul service %q: %w", service, err)
	}

	kvPair, _, err := client.KV().Get(key, nil)
	if err != nil {
		return nil, fmt.Errorf("get consul kv %q: %w", key, err)
	}
	if kvPair == nil {
		return nil, fmt.Errorf("consul kv %q not found", key)
	}

	var cfg Config
	if err := json.Unmarshal(kvPair.Value, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config from consul kv %q: %w", key, err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}
