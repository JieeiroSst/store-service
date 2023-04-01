package consul

import (
	"encoding/json"
	"log"
	"os"

	"github.com/JIeeiroSst/user-service/config"
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
	return consul, err

}

func (c *configConsul) getKvPair(client *consulapi.Client, key string) (*consulapi.KVPair, error) {
	kv := client.KV()
	keyPair, _, err := kv.Get(key, nil)
	return keyPair, err
}

func (c *configConsul) ConnectConfigConsul() (config *config.Config, err error) {
	consul, err := c.getConsul(c.Host)
	if err != nil {
		log.Fatalf("Error connecting to Consul: %s", err)
	}

	cat := consul.Catalog()
	_, _, err = cat.Service(c.Service, "", nil)
	if err != nil {
		return nil, err
	}

	redisPattern, err := c.getKvPair(consul, c.Key)
	if err != nil || redisPattern == nil {
		log.Fatalf("Could not get REDISPATTERN: %s", err)
	}
	log.Printf("KV: %v %s\n", redisPattern.Key, redisPattern.Value)

	if err := json.Unmarshal(redisPattern.Value, &config); err != nil {
		return nil, err
	}

	return config, nil
}
