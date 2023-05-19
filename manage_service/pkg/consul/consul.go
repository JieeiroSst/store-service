package consul

import (
	"encoding/json"
	"os"

	"github.com/JIeeiroSst/manage-service/config"
	"github.com/JIeeiroSst/manage-service/pkg/log"
	consulapi "github.com/hashicorp/consul/api"
)

type configConsul struct {
	Host    string
	Key     string
	Service string
}

type Consul interface {
	getEnv(key, fallback string) string
	getConsul(address string) (*consulapi.Client, error)
	getKvPair(client *consulapi.Client, key string) (*consulapi.KVPair, error)
	ConnectConfigConsul() (*config.Config, error)
}

func NewConfigConsul(host, key, service string) Consul {
	return &configConsul{
		Host:    host,
		Key:     key,
		Service: service,
	}
}

func (c *configConsul) getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func (c *configConsul) getConsul(address string) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = address
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return consul, err

}

func (c *configConsul) getKvPair(client *consulapi.Client, key string) (*consulapi.KVPair, error) {
	kv := client.KV()
	keyPair, _, err := kv.Get(key, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return keyPair, err
}

func (c *configConsul) ConnectConfigConsul() (config *config.Config, err error) {
	consul, err := c.getConsul(c.Host)
	if err != nil {
		log.Error("Error connecting to Consul")
		return nil, err
	}

	cat := consul.Catalog()
	_, _, err = cat.Service(c.Service, "", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	redisPattern, err := c.getKvPair(consul, c.Key)
	if err != nil || redisPattern == nil {
		log.Error("Could not get REDISPATTERN")
		return nil, err
	}

	if err := json.Unmarshal(redisPattern.Value, &config); err != nil {
		log.Error(err)
		return nil, err
	}

	log.Info(config)
	return config, nil
}
