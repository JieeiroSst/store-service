package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var Version = "dev"

type Config struct {
	HTTPPort        int
	LogLevel        string
	DefaultTimezone string
	Postgres        PostgresConfig
	RabbitMQ        RabbitMQConfig
	Relay           RelayConfig
	Redis           RedisConfig
	RateLimit       RateLimitConfig
	Fraud           FraudConfig
}

type FraudConfig struct {
	RequireDevice     bool
	MaxUsersPerDevice int
	MaxUsersPerIP     int
}

type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	Prefix      string
	CampaignTTL time.Duration
	WinnersSize int
	OpTimeout   time.Duration
}

func (r RedisConfig) Enabled() bool { return r.Addr != "" }

type RateLimitConfig struct {
	PlaysPerMinute int
}

type PostgresConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

func (p PostgresConfig) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		Path:     "/" + p.Database,
		RawQuery: url.Values{"sslmode": {p.SSLMode}}.Encode(),
	}
	return u.String()
}

type RabbitMQConfig struct {
	Enabled    bool
	Host       string
	Port       int
	User       string
	Password   string
	VHost      string
	Exchange   string
	RoutingKey string
}

func (r RabbitMQConfig) URL() string {
	u := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(r.User, r.Password),
		Host:   net.JoinHostPort(r.Host, strconv.Itoa(r.Port)),
		Path:   "/" + strings.TrimPrefix(r.VHost, "/"),
	}
	return u.String()
}

type RelayConfig struct {
	Interval time.Duration
	Grace    time.Duration
	Batch    int
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:        9000,
		LogLevel:        str("LOG_LEVEL", "info"),
		DefaultTimezone: str("DEFAULT_TIMEZONE", "Asia/Ho_Chi_Minh"),
		Postgres: PostgresConfig{
			Host:         str("DB_HOST", "localhost"),
			Port:         5432,
			User:         str("DB_USER", "postgres"),
			Password:     os.Getenv("DB_PASSWORD"),
			Database:     str("DB_NAME", "lucky_service"),
			SSLMode:      str("DB_SSLMODE", "disable"),
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		RabbitMQ: RabbitMQConfig{
			Host:       str("RABBITMQ_HOST", "localhost"),
			Port:       5672,
			User:       str("RABBITMQ_USER", "guest"),
			Password:   str("RABBITMQ_PASSWORD", "guest"),
			VHost:      os.Getenv("RABBITMQ_VHOST"),
			Exchange:   str("RABBITMQ_EXCHANGE", "lucky"),
			RoutingKey: str("RABBITMQ_ROUTING_KEY", "lucky.reward.granted"),
		},
		Redis: RedisConfig{
			Addr:        os.Getenv("REDIS_ADDR"),
			Password:    os.Getenv("REDIS_PASSWORD"),
			Prefix:      str("REDIS_PREFIX", "lucky:"),
			CampaignTTL: 15 * time.Second,
			WinnersSize: 50,
			OpTimeout:   300 * time.Millisecond,
		},
		RateLimit: RateLimitConfig{PlaysPerMinute: 30},
		Fraud:     FraudConfig{MaxUsersPerDevice: 3, MaxUsersPerIP: 50},
		Relay: RelayConfig{
			Interval: 10 * time.Second,
			Grace:    5 * time.Second,
			Batch:    100,
		},
	}
	if _, err := time.LoadLocation(cfg.DefaultTimezone); err != nil {
		return nil, fmt.Errorf("DEFAULT_TIMEZONE: %w", err)
	}
	enabled, err := boolean("RABBITMQ_ENABLED", false)
	if err != nil {
		return nil, err
	}
	cfg.RabbitMQ.Enabled = enabled
	if cfg.Fraud.RequireDevice, err = boolean("FRAUD_REQUIRE_DEVICE", false); err != nil {
		return nil, err
	}
	ints := []struct {
		key string
		dst *int
	}{
		{"HTTP_PORT", &cfg.HTTPPort},
		{"DB_PORT", &cfg.Postgres.Port},
		{"DB_MAX_OPEN_CONNS", &cfg.Postgres.MaxOpenConns},
		{"DB_MAX_IDLE_CONNS", &cfg.Postgres.MaxIdleConns},
		{"RABBITMQ_PORT", &cfg.RabbitMQ.Port},
		{"RELAY_BATCH", &cfg.Relay.Batch},
		{"REDIS_DB", &cfg.Redis.DB},
		{"WINNER_FEED_SIZE", &cfg.Redis.WinnersSize},
		{"RATE_LIMIT_PLAYS_PER_MINUTE", &cfg.RateLimit.PlaysPerMinute},
		{"FRAUD_MAX_USERS_PER_DEVICE", &cfg.Fraud.MaxUsersPerDevice},
		{"FRAUD_MAX_USERS_PER_IP", &cfg.Fraud.MaxUsersPerIP},
	}
	for _, v := range ints {
		if err := num(v.key, v.dst); err != nil {
			return nil, err
		}
	}
	durations := []struct {
		key string
		dst *time.Duration
	}{
		{"RELAY_INTERVAL", &cfg.Relay.Interval},
		{"RELAY_GRACE", &cfg.Relay.Grace},
		{"CACHE_CAMPAIGN_TTL", &cfg.Redis.CampaignTTL},
		{"REDIS_OP_TIMEOUT", &cfg.Redis.OpTimeout},
	}
	for _, v := range durations {
		if err := dur(v.key, v.dst); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func str(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func boolean(key string, def bool) (bool, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}
	return b, nil
}

func num(key string, dst *int) error {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	*dst = n
	return nil
}

func dur(key string, dst *time.Duration) error {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	if d <= 0 {
		return fmt.Errorf("%s: must be positive", key)
	}
	*dst = d
	return nil
}
