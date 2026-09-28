package module

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	App          AppConfig
	Database     DatabaseConfig
	Temporal     TemporalConfig
	AI           AIConfig
	Notification NotificationConfig
}

type AppConfig struct {
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
	// UserServiceURL is user-service's HTTP gateway; it issues and
	// validates every bearer token this service accepts.
	UserServiceURL     string        `mapstructure:"user_service_url"`
	UserServiceTimeout time.Duration `mapstructure:"user_service_timeout"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type TemporalConfig struct {
	HostPort  string `mapstructure:"host_port"`
	Namespace string `mapstructure:"namespace"`
}

type AIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
}

type NotificationConfig struct {
	APIKey    string `mapstructure:"api_key"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	v.SetDefault("app.env", "development")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.user_service_url", "http://user-service:1235")
	v.SetDefault("app.user_service_timeout", "3s")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("temporal.host_port", "localhost:7233")
	v.SetDefault("temporal.namespace", "recruitment")
	v.SetDefault("ai.base_url", "https://api.openai.com/v1")
	v.SetDefault("ai.model", "gpt-4o")
	v.SetDefault("notification.from_name", "Recruitment Platform")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := &Config{}
	return cfg, v.Unmarshal(cfg)
}
