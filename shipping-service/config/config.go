package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server       ServerConfig
	Mysql        MysqlConfig
	GHN          GHNConfig
	Notification NotificationConfig
	Callback     CallbackConfig
	Auth         AuthConfig
	Shipping     ShippingConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
}

type GHNConfig struct {
	BaseURL           string
	Token             string
	WebhookToken      string
	Timeout           string
	LocationCacheTTLs string
}

type NotificationConfig struct {
	BaseURL string
	Timeout string
}

type CallbackConfig struct {
	Secret       string
	AllowedHosts string
	Timeout      string
}

type AuthConfig struct {
	APIKeys string
}

type ShippingConfig struct {
	MaxCODAmount   string
	OutboxInterval string
}

func (c GHNConfig) TimeoutDuration() time.Duration { return duration(c.Timeout, 10*time.Second) }
func (c GHNConfig) LocationCacheTTL() time.Duration {
	return duration(c.LocationCacheTTLs, 24*time.Hour)
}
func (c NotificationConfig) TimeoutDuration() time.Duration {
	return duration(c.Timeout, 5*time.Second)
}
func (c CallbackConfig) TimeoutDuration() time.Duration { return duration(c.Timeout, 5*time.Second) }
func (c ShippingConfig) OutboxIntervalDuration() time.Duration {
	return duration(c.OutboxInterval, 2*time.Second)
}

func (c ShippingConfig) MaxCOD() int64 {
	if v, err := strconv.ParseInt(c.MaxCODAmount, 10, 64); err == nil && v > 0 {
		return v
	}
	return 50_000_000
}

func (c CallbackConfig) AllowedHostList() []string {
	var hosts []string
	for _, h := range strings.Split(c.AllowedHosts, ",") {
		if h = strings.ToLower(strings.TrimSpace(h)); h != "" {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

func (c AuthConfig) Keys() map[string]string {
	keys := map[string]string{}
	for _, pair := range strings.Split(c.APIKeys, ",") {
		name, key, ok := strings.Cut(strings.TrimSpace(pair), ":")
		name, key = strings.TrimSpace(name), strings.TrimSpace(key)
		if ok && name != "" && key != "" {
			keys[key] = name
		}
	}
	return keys
}

func duration(raw string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	return fallback
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{PortHttpServer: getEnv("PORT_HTTP_SERVER", "8094")},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "shipping_service"),
		},
		GHN: GHNConfig{
			BaseURL:           getEnv("GHN_BASE_URL", "https://dev-online-gateway.ghn.vn/shiip/public-api"),
			Token:             getEnv("GHN_TOKEN", ""),
			WebhookToken:      getEnv("GHN_WEBHOOK_TOKEN", ""),
			Timeout:           getEnv("GHN_TIMEOUT", "10s"),
			LocationCacheTTLs: getEnv("GHN_LOCATION_CACHE_TTL", "24h"),
		},
		Notification: NotificationConfig{
			BaseURL: getEnv("NOTIFICATION_BASE_URL", ""),
			Timeout: getEnv("NOTIFICATION_TIMEOUT", "5s"),
		},
		Callback: CallbackConfig{
			Secret:       getEnv("CALLBACK_SECRET", ""),
			AllowedHosts: getEnv("CALLBACK_ALLOWED_HOSTS", ""),
			Timeout:      getEnv("CALLBACK_TIMEOUT", "5s"),
		},
		Auth: AuthConfig{APIKeys: getEnv("API_KEYS", "")},
		Shipping: ShippingConfig{
			MaxCODAmount:   getEnv("MAX_COD_AMOUNT", "50000000"),
			OutboxInterval: getEnv("OUTBOX_INTERVAL", "2s"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
