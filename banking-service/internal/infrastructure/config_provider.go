package infrastructure

import (
	"log"
	"os"

	"github.com/JieeiroSst/banking-service/config"
	"github.com/JieeiroSst/banking-service/pkg/consul"
	"github.com/joho/godotenv"
)

// newConfig loads configuration from Consul when HostConsul/KeyConsul/
// ServiceConsul are set, falling back to plain environment variables
// otherwise (e.g. local dev, CI, or a Consul-less deployment).
func newConfig() (*config.Config, error) {
	_ = godotenv.Load(".env")

	host := os.Getenv("HostConsul")
	key := os.Getenv("KeyConsul")
	service := os.Getenv("ServiceConsul")

	if host == "" || key == "" || service == "" {
		return config.FromEnv(), nil
	}

	cfg, err := consul.NewConfigConsul(host, key, service).ConnectConfigConsul()
	if err != nil || cfg == nil {
		log.Printf("consul config unavailable, falling back to env: %v", err)
		return config.FromEnv(), nil
	}
	return cfg, nil
}
