package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var Version = "1.0.0"

type Config struct {
	Server  ServerConfig  `json:"server"`
	SerpAPI SerpAPIConfig `json:"serpapi"`
	Search  SearchConfig  `json:"search"`
	Cache   CacheConfig   `json:"cache"`
}

type ServerConfig struct {
	PortHttpServer string   `json:"portHttpServer"`
	AccessTokens   []string `json:"-"`
}

type SerpAPIConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
}

type SearchConfig struct {
	StrictEngines bool `json:"strictEngines"`
	MaxConcurrent int  `json:"maxConcurrent"`
	MaxBatchSize  int  `json:"maxBatchSize"`
}

type CacheConfig struct {
	Backend       string
	TTL           time.Duration
	LocationsTTL  time.Duration
	MaxEntries    int
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisPrefix   string
}

func Defaults() *Config {
	return &Config{
		Server: ServerConfig{PortHttpServer: "8080"},
		SerpAPI: SerpAPIConfig{
			BaseURL:    "https://serpapi.com",
			Timeout:    90 * time.Second,
			MaxRetries: 2,
		},
		Search: SearchConfig{
			MaxConcurrent: 32,
			MaxBatchSize:  20,
		},
		Cache: CacheConfig{
			Backend:      "memory",
			TTL:          time.Hour,
			LocationsTTL: 24 * time.Hour,
			MaxEntries:   1000,
			RedisAddr:    "localhost:6379",
			RedisPrefix:  "serpapi:",
		},
	}
}

func (c *Config) Merge(raw []byte) error {
	if err := json.Unmarshal(raw, c); err != nil {
		return fmt.Errorf("parse config json: %w", err)
	}
	return nil
}

func (c *Config) ApplyEnv() {
	setString(&c.Server.PortHttpServer, "PORT_HTTP_SERVER")
	if v, ok := os.LookupEnv("ACCESS_TOKENS"); ok {
		c.Server.AccessTokens = splitList(v)
	}

	setString(&c.SerpAPI.BaseURL, "SERPAPI_BASE_URL")
	setString(&c.SerpAPI.APIKey, "SERPAPI_API_KEY")
	setDuration(&c.SerpAPI.Timeout, "SERPAPI_TIMEOUT")
	setNonNegInt(&c.SerpAPI.MaxRetries, "SERPAPI_MAX_RETRIES")

	setBool(&c.Search.StrictEngines, "STRICT_ENGINES")
	setInt(&c.Search.MaxConcurrent, "MAX_CONCURRENT_SEARCHES")
	setInt(&c.Search.MaxBatchSize, "MAX_BATCH_SIZE")

	setString(&c.Cache.Backend, "CACHE_BACKEND")
	setDuration(&c.Cache.TTL, "CACHE_TTL")
	setDuration(&c.Cache.LocationsTTL, "LOCATIONS_CACHE_TTL")
	setInt(&c.Cache.MaxEntries, "CACHE_MAX_ENTRIES")
	setString(&c.Cache.RedisAddr, "REDIS_ADDR")
	setString(&c.Cache.RedisPassword, "REDIS_PASSWORD")
	setNonNegInt(&c.Cache.RedisDB, "REDIS_DB")
	setString(&c.Cache.RedisPrefix, "REDIS_PREFIX")
}

func (c *Config) normalize() {
	c.SerpAPI.BaseURL = strings.TrimRight(c.SerpAPI.BaseURL, "/")
}

func (c *Config) Validate() error {
	if c.SerpAPI.APIKey == "" {
		return fmt.Errorf("SERPAPI_API_KEY is not set")
	}
	if c.SerpAPI.BaseURL == "" {
		return fmt.Errorf("config serpapi.baseUrl is empty")
	}
	switch c.Cache.Backend {
	case "memory", "redis", "none":
	default:
		return fmt.Errorf("config cache.backend must be memory, redis or none, got %q", c.Cache.Backend)
	}
	positive := map[string]int64{
		"serpapi.timeout":      int64(c.SerpAPI.Timeout),
		"search.maxConcurrent": int64(c.Search.MaxConcurrent),
		"search.maxBatchSize":  int64(c.Search.MaxBatchSize),
		"cache.maxEntries":     int64(c.Cache.MaxEntries),
	}
	for name, v := range positive {
		if v <= 0 {
			return fmt.Errorf("config %s must be positive", name)
		}
	}
	if c.SerpAPI.MaxRetries < 0 || c.Cache.TTL < 0 || c.Cache.LocationsTTL < 0 {
		return fmt.Errorf("config serpapi.maxRetries, cache.ttl and cache.locationsTtl must not be negative")
	}
	return nil
}

type duration struct{ d *time.Duration }

func (d duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration must be a string like \"5m\": %w", err)
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d.d = v
	return nil
}

func (s *SerpAPIConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		BaseURL    *string  `json:"baseUrl"`
		Timeout    duration `json:"timeout"`
		MaxRetries *int     `json:"maxRetries"`
	}{
		BaseURL:    &s.BaseURL,
		Timeout:    duration{&s.Timeout},
		MaxRetries: &s.MaxRetries,
	}
	return json.Unmarshal(b, &aux)
}

func (c *CacheConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		Backend      *string  `json:"backend"`
		TTL          duration `json:"ttl"`
		LocationsTTL duration `json:"locationsTtl"`
		MaxEntries   *int     `json:"maxEntries"`
		RedisAddr    *string  `json:"redisAddr"`
		RedisDB      *int     `json:"redisDb"`
		RedisPrefix  *string  `json:"redisPrefix"`
	}{
		Backend:      &c.Backend,
		TTL:          duration{&c.TTL},
		LocationsTTL: duration{&c.LocationsTTL},
		MaxEntries:   &c.MaxEntries,
		RedisAddr:    &c.RedisAddr,
		RedisDB:      &c.RedisDB,
		RedisPrefix:  &c.RedisPrefix,
	}
	return json.Unmarshal(b, &aux)
}

func setString(dst *string, key string) {
	if v, ok := os.LookupEnv(key); ok {
		*dst = v
	}
}

func setInt(dst *int, key string) {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		*dst = v
	}
}

func setNonNegInt(dst *int, key string) {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v >= 0 {
		*dst = v
	}
}

func setBool(dst *bool, key string) {
	if v, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		*dst = v
	}
}

func setDuration(dst *time.Duration, key string) {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil && v >= 0 {
		*dst = v
	}
}

func splitList(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
