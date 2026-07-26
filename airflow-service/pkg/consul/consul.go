package consul

import (
	"encoding/json"
	"log"

	"github.com/JIeeiroSst/airflow-service/config"
	consulapi "github.com/hashicorp/consul/api"
)

type configConsul struct {
	Host    string
	Key     string
	Service string
}

type Consul interface {
	ConnectConfigConsul() (*config.Config, error)
}

func NewConfigConsul(host, key, service string) Consul {
	return &configConsul{
		Host:    host,
		Key:     key,
		Service: service,
	}
}

func (c *configConsul) getConsul(address string) (*consulapi.Client, error) {
	cfg := consulapi.DefaultConfig()
	cfg.Address = address
	return consulapi.NewClient(cfg)
}

func (c *configConsul) getKvPair(client *consulapi.Client, key string) (*consulapi.KVPair, error) {
	kv := client.KV()
	keyPair, _, err := kv.Get(key, nil)
	return keyPair, err
}

func (c *configConsul) ConnectConfigConsul() (cfg *config.Config, err error) {
	consulClient, err := c.getConsul(c.Host)
	if err != nil {
		log.Fatalf("Error connecting to Consul: %s", err)
	}

	cat := consulClient.Catalog()
	if _, _, err := cat.Service(c.Service, "", nil); err != nil {
		return nil, err
	}

	pair, err := c.getKvPair(consulClient, c.Key)
	if err != nil || pair == nil {
		log.Fatalf("Could not get config KV: %s", err)
	}

	if err := json.Unmarshal(pair.Value, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
