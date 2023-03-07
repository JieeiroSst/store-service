package consul

import (
	"log"
	"os"

	consulapi "github.com/hashicorp/consul/api"
)

type ConfigConsul struct {
	Host string
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func getConsul(address string) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = address
	consul, err := consulapi.NewClient(config)
	return consul, err

}

func getKvPair(client *consulapi.Client, key string) (*consulapi.KVPair, error) {
	kv := client.KV()
	keyPair, _, err := kv.Get(key, nil)
	return keyPair, err
}

func (c *ConfigConsul) ConnectConfigConsul(service, key string) (*consulapi.KVPair, error) {
	consul, err := getConsul(c.Host)
	if err != nil {
		log.Fatalf("Error connecting to Consul: %s", err)
	}

	cat := consul.Catalog()
	_, _, err = cat.Service(service, "", nil)
	if err != nil {
		return nil, err
	}

	redisPattern, err := getKvPair(consul, key)
	if err != nil || redisPattern == nil {
		log.Fatalf("Could not get REDISPATTERN: %s", err)
	}
	log.Printf("KV: %v %s\n", redisPattern.Key, redisPattern.Value)

	return redisPattern, nil
}
