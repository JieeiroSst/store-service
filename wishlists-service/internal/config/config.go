package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

var Version = "dev"

type Config struct {
	HTTPPort    int
	LogLevel    string
	Postgres    PostgresConfig
	Redis       RedisConfig
	UserService UserServiceConfig
	RateLimit   RateLimitConfig
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

type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	Prefix      string
	WishlistTTL time.Duration
	UserTTL     time.Duration
	MissingTTL  time.Duration
	PopularTTL  time.Duration
	OpTimeout   time.Duration
}

func (r RedisConfig) Enabled() bool { return r.Addr != "" }

type UserServiceConfig struct {
	URL     string
	Timeout time.Duration
}

type RateLimitConfig struct {
	WritesPerMinute int
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort: 9000,
		LogLevel: str("LOG_LEVEL", "info"),
		Postgres: PostgresConfig{
			Host:         str("DB_HOST", "localhost"),
			Port:         5432,
			User:         str("DB_USER", "postgres"),
			Password:     os.Getenv("DB_PASSWORD"),
			Database:     str("DB_NAME", "wishlists_service"),
			SSLMode:      str("DB_SSLMODE", "disable"),
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Redis: RedisConfig{
			Addr:        os.Getenv("REDIS_ADDR"),
			Password:    os.Getenv("REDIS_PASSWORD"),
			Prefix:      str("REDIS_PREFIX", "wishlists:"),
			WishlistTTL: 5 * time.Minute,
			UserTTL:     10 * time.Minute,
			MissingTTL:  time.Minute,
			PopularTTL:  5 * time.Minute,
			OpTimeout:   300 * time.Millisecond,
		},
		UserService: UserServiceConfig{
			URL:     str("USER_SERVICE_URL", "http://user-api-svc"),
			Timeout: 2 * time.Second,
		},
		RateLimit: RateLimitConfig{WritesPerMinute: 120},
	}
	ints := []struct {
		key string
		dst *int
	}{
		{"HTTP_PORT", &cfg.HTTPPort},
		{"DB_PORT", &cfg.Postgres.Port},
		{"DB_MAX_OPEN_CONNS", &cfg.Postgres.MaxOpenConns},
		{"DB_MAX_IDLE_CONNS", &cfg.Postgres.MaxIdleConns},
		{"REDIS_DB", &cfg.Redis.DB},
		{"RATE_LIMIT_WRITES_PER_MINUTE", &cfg.RateLimit.WritesPerMinute},
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
		{"CACHE_WISHLIST_TTL", &cfg.Redis.WishlistTTL},
		{"CACHE_USER_TTL", &cfg.Redis.UserTTL},
		{"CACHE_MISSING_TTL", &cfg.Redis.MissingTTL},
		{"CACHE_POPULAR_TTL", &cfg.Redis.PopularTTL},
		{"REDIS_OP_TIMEOUT", &cfg.Redis.OpTimeout},
		{"USER_SERVICE_TIMEOUT", &cfg.UserService.Timeout},
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
