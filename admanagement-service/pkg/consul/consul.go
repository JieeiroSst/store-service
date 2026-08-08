package consul

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/JIeeiroSst/admanagement-service/config"
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

func (c *configConsul) ConnectConfigConsul() (config *config.Config, err error) {
	consul, err := c.getConsul(c.Host)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Consul: %w", err)
	}

	cat := consul.Catalog()
	_, _, err = cat.Service(c.Service, "", nil)
	if err != nil {
		return nil, err
	}

	kvPair, err := c.getKvPair(consul, c.Key)
	if err != nil {
		return nil, fmt.Errorf("could not get key %q from Consul: %w", c.Key, err)
	}
	if kvPair == nil {
		return nil, fmt.Errorf("key %q not found in Consul", c.Key)
	}
	log.Printf("KV: %v %s\n", kvPair.Key, kvPair.Value)

	if err := json.Unmarshal(kvPair.Value, &config); err != nil {
		return nil, err
	}

	return config, nil
}
