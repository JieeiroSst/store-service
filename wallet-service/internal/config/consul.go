package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
)

const envFile = ".env"

type ConsulSource struct {
	Host    string
	Key     string
	Service string
}

func consulSourceFromEnv() (*ConsulSource, error) {
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load %s: %w", envFile, err)
	}
	src := &ConsulSource{
		Host:    strings.TrimSpace(os.Getenv("HostConsul")),
		Key:     strings.TrimSpace(os.Getenv("KeyConsul")),
		Service: strings.TrimSpace(os.Getenv("ServiceConsul")),
	}
	if src.Host == "" {
		return nil, nil
	}
	if src.Key == "" {
		return nil, errors.New("KeyConsul is required when HostConsul is set")
	}
	return src, nil
}

func (s *ConsulSource) Fetch() ([]byte, error) {
	cfg := consulapi.DefaultConfig()
	cfg.Address = s.Host
	client, err := consulapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("consul client: %w", err)
	}
	if s.Service != "" {
		if _, _, err := client.Catalog().Service(s.Service, "", nil); err != nil {
			return nil, fmt.Errorf("consul %s unreachable: %w", s.Host, err)
		}
	}
	pair, _, err := client.KV().Get(s.Key, nil)
	if err != nil {
		return nil, fmt.Errorf("consul get %q: %w", s.Key, err)
	}
	if pair == nil || len(pair.Value) == 0 {
		return nil, fmt.Errorf("consul key %q not found or empty", s.Key)
	}
	return pair.Value, nil
}
