package infrastructure

import (
	"encoding/json"

	"github.com/JIeeiroSst/delivery-service/config"
	consulapi "github.com/hashicorp/consul/api"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	data, err := readConsulKey(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func readConsulKey(host, key, service string) ([]byte, error) {
	cfg := consulapi.DefaultConfig()
	cfg.Address = host
	client, err := consulapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if _, _, err := client.Catalog().Service(service, "", nil); err != nil {
		return nil, err
	}
	pair, _, err := client.KV().Get(key, nil)
	if err != nil || pair == nil {
		return nil, err
	}
	return pair.Value, nil
}
