package config

import (
	"os"
)

type Config struct {
	SolanaRPCURL   string
	Port           string
	CircleAPIKey   string
	CircleEntityID string
}

func Load() *Config {
	rpcURL := os.Getenv("SOLANA_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://api.devnet.solana.com"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	circleAPIKey := os.Getenv("CIRCLE_API_KEY")
	circleEntityID := os.Getenv("CIRCLE_ENTITY_ID")

	return &Config{
		SolanaRPCURL:   rpcURL,
		Port:           port,
		CircleAPIKey:   circleAPIKey,
		CircleEntityID: circleEntityID,
	}
}
