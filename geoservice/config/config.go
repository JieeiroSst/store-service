package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTPPort    string `mapstructure:"http_port"`
	DatabaseURL string `mapstructure:"database_url"`
	LogLevel    string `mapstructure:"log_level"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("GEO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("http_port", "8081")
	v.SetDefault("database_url", "postgres://geo:geo@localhost:5432/geo?sslmode=disable")
	v.SetDefault("log_level", "info")

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
