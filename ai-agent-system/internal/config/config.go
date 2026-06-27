package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port                   string
	OllamaBaseURL          string
	OllamaModel            string
	OllamaEmbedModel       string
	FAQFile                string
	FAQSimilarityThreshold float64
	MySQLDSN               string
}

func Load() (Config, error) {
	cfg := Config{
		Port:             getEnv("PORT", "8080"),
		OllamaBaseURL:    getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:      os.Getenv("OLLAMA_MODEL"),
		OllamaEmbedModel: getEnv("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
		FAQFile:          getEnv("FAQ_FILE", "./configs/faq.yaml"),
		MySQLDSN:         os.Getenv("MYSQL_DSN"),
	}

	threshold, err := strconv.ParseFloat(getEnv("FAQ_SIMILARITY_THRESHOLD", "0.85"), 64)
	if err != nil {
		return Config{}, fmt.Errorf("config: invalid FAQ_SIMILARITY_THRESHOLD: %w", err)
	}
	cfg.FAQSimilarityThreshold = threshold

	if cfg.OllamaModel == "" {
		return Config{}, fmt.Errorf("config: OLLAMA_MODEL is required")
	}
	if cfg.MySQLDSN == "" {
		return Config{}, fmt.Errorf("config: MYSQL_DSN is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
