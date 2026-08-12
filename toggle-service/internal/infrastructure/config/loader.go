package config

import (
	"encoding/json"
	"os"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	_ = v.ReadInConfig() // .env is optional; real env vars still apply via AutomaticEnv

	host := firstNonEmpty(v.GetString("HostConsul"), os.Getenv("HostConsul"))
	key := firstNonEmpty(v.GetString("KeyConsul"), os.Getenv("KeyConsul"))
	service := firstNonEmpty(v.GetString("ServiceConsul"), os.Getenv("ServiceConsul"))

	if host != "" && key != "" && service != "" {
		if raw, err := consul.NewConfigConsul(host, key, service).ConnectConfigConsul(); err == nil {
			var cfg Config
			if err := json.Unmarshal(raw, &cfg); err == nil {
				applyDefaults(&cfg)
				return &cfg, nil
			}
		}
	}

	cfg := Config{
		Server: ServerConfig{
			Port: v.GetString("HTTP_PORT"),
			Env:  v.GetString("APP_ENV"),
		},
		Postgres: PostgresConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			DBName:   v.GetString("DB_NAME"),
			SSLMode:  v.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret:        v.GetString("JWT_SECRET"),
			ExpiryMinutes: v.GetInt("JWT_EXPIRY_MINUTES"),
		},
	}
	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.Env == "" {
		cfg.Server.Env = "development"
	}
	if cfg.Postgres.Host == "" {
		cfg.Postgres.Host = "localhost"
	}
	if cfg.Postgres.Port == "" {
		cfg.Postgres.Port = "5432"
	}
	if cfg.Postgres.User == "" {
		cfg.Postgres.User = "postgres"
	}
	if cfg.Postgres.DBName == "" {
		cfg.Postgres.DBName = "toggle_service"
	}
	if cfg.Postgres.SSLMode == "" {
		cfg.Postgres.SSLMode = "disable"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-me"
	}
	if cfg.JWT.ExpiryMinutes == 0 {
		cfg.JWT.ExpiryMinutes = 60
	}
}

func firstNonEmpty(vals ...string) string {
	for _, val := range vals {
		if val != "" {
			return val
		}
	}
	return ""
}

var Module = fx.Options(fx.Provide(NewConfig))
