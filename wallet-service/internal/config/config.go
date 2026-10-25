package config

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server          ServerConfig          `mapstructure:"server"`
	Database        DatabaseConfig        `mapstructure:"database"`
	Redis           RedisConfig           `mapstructure:"redis"`
	Log             LogConfig             `mapstructure:"log"`
	Fees            FeesConfig            `mapstructure:"fees"`
	Security        SecurityConfig        `mapstructure:"security"`
	PaymentRequests PaymentRequestsConfig `mapstructure:"payment_requests"`
	BankGateway     BankGatewayConfig     `mapstructure:"bank_gateway"`
	Source          string                `mapstructure:"-"`
}

type ServerConfig struct {
	Port              string  `mapstructure:"port"`
	ReadTimeout       int     `mapstructure:"read_timeout"`
	WriteTimeout      int     `mapstructure:"write_timeout"`
	ShutdownTimeout   int     `mapstructure:"shutdown_timeout"`
	RateLimitCapacity float64 `mapstructure:"rate_limit_capacity"`
	RateLimitPerSec   float64 `mapstructure:"rate_limit_per_sec"`
}

type DatabaseConfig struct {
	DSN             string `mapstructure:"dsn"`
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	AutoMigrate     bool   `mapstructure:"auto_migrate"`
	SeedDevData     bool   `mapstructure:"seed_dev_data"`
}

func (d DatabaseConfig) ConnString() (string, error) {
	if d.DSN != "" {
		return d.DSN, nil
	}
	if d.Host == "" || d.Name == "" {
		return "", errors.New("database.dsn or database.host and database.name are required")
	}
	u := url.URL{
		Scheme:   "postgres",
		Host:     net.JoinHostPort(d.Host, d.Port),
		Path:     "/" + d.Name,
		RawQuery: url.Values{"sslmode": {d.SSLMode}}.Encode(),
	}
	if d.User != "" {
		u.User = url.UserPassword(d.User, d.Password)
	}
	return u.String(), nil
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type FeesConfig struct {
	WithdrawalRate   string `mapstructure:"withdrawal_rate"`
	WithdrawalMinUSD string `mapstructure:"withdrawal_min_usd"`
	TransferRate     string `mapstructure:"transfer_rate"`
}

type SecurityConfig struct {
	PinBcryptCost int    `mapstructure:"pin_bcrypt_cost"`
	CardHashKey   string `mapstructure:"card_hash_key"`
}

type PaymentRequestsConfig struct {
	TTLHours int `mapstructure:"ttl_hours"`
}

type BankGatewayConfig struct {
	DeclineAbove string `mapstructure:"decline_above"`
}

func (c PaymentRequestsConfig) TTL() time.Duration {
	return time.Duration(c.TTLHours) * time.Hour
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("server.port", ":8080")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.shutdown_timeout", 15)
	v.SetDefault("server.rate_limit_capacity", 100)
	v.SetDefault("server.rate_limit_per_sec", 50)
	v.SetDefault("database.dsn", "")
	v.SetDefault("database.host", "")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.seed_dev_data", false)
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 300)
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("log.level", "info")
	v.SetDefault("fees.withdrawal_rate", "0.005")
	v.SetDefault("fees.withdrawal_min_usd", "0.5")
	v.SetDefault("fees.transfer_rate", "0")
	v.SetDefault("security.pin_bcrypt_cost", 10)
	v.SetDefault("security.card_hash_key", "")
	v.SetDefault("payment_requests.ttl_hours", 168)
	v.SetDefault("bank_gateway.decline_above", "0")

	source, err := readSource(v)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	cfg.Source = source
	if cfg.Server.Port != "" && !strings.Contains(cfg.Server.Port, ":") {
		cfg.Server.Port = ":" + cfg.Server.Port
	}
	return &cfg, nil
}

func readSource(v *viper.Viper) (string, error) {
	consul, err := consulSourceFromEnv()
	if err != nil {
		return "", err
	}
	if consul != nil {
		raw, err := consul.Fetch()
		if err != nil {
			return "", err
		}
		v.SetConfigType("json")
		if err := v.ReadConfig(bytes.NewReader(raw)); err != nil {
			return "", fmt.Errorf("parse consul key %q: %w", consul.Key, err)
		}
		return "consul:" + consul.Key, nil
	}
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return "defaults", nil
		}
		return "", err
	}
	return "file:" + v.ConfigFileUsed(), nil
}
