package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

var Version = "1.0.0"

const minMasterKeyLen = 32

type Config struct {
	Server   ServerConfig   `json:"server"`
	Security SecurityConfig `json:"-"`
	Storage  StorageConfig  `json:"storage"`
	Postgres PostgresConfig `json:"postgres"`
	Limits   LimitsConfig   `json:"limits"`

	CustomerInfo ServiceConfig `json:"customerInfo"`
	AuthService  ServiceConfig `json:"authService"`
}

type ServiceConfig struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

func (s *ServiceConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		BaseURL *string `json:"baseUrl"`
		Timeout *string `json:"timeout"`
	}{BaseURL: &s.BaseURL}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	if aux.Timeout != nil {
		d, err := time.ParseDuration(*aux.Timeout)
		if err != nil {
			return fmt.Errorf("timeout: %w", err)
		}
		s.Timeout = d
	}
	return nil
}

type ServerConfig struct {
	PortHttpServer string   `json:"portHttpServer"`
	AccessTokens   []string `json:"-"`
	InternalTokens []string `json:"-"`
}

type SecurityConfig struct {
	MasterKey string
}

type StorageConfig struct {
	Backend string `json:"backend"`
}

type PostgresConfig struct {
	Host           string        `json:"host"`
	Port           string        `json:"port"`
	User           string        `json:"user"`
	Password       string        `json:"password"`
	Database       string        `json:"database"`
	SSLMode        string        `json:"sslMode"`
	MaxConns       int32         `json:"maxConns"`
	ConnectTimeout time.Duration `json:"-"`
}

func (p PostgresConfig) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, p.Port),
		Path:     "/" + p.Database,
		RawQuery: url.Values{"sslmode": {p.SSLMode}}.Encode(),
	}
	return u.String()
}

type LimitsConfig struct {
	Timezone string         `json:"timezone"`
	Location *time.Location `json:"-"`
}

func Defaults() *Config {
	return &Config{
		Server:  ServerConfig{PortHttpServer: "8080"},
		Storage: StorageConfig{Backend: "memory"},
		Postgres: PostgresConfig{
			Host:           "localhost",
			Port:           "5432",
			Database:       "card",
			SSLMode:        "disable",
			MaxConns:       10,
			ConnectTimeout: 30 * time.Second,
		},
		Limits:       LimitsConfig{Timezone: "Asia/Ho_Chi_Minh"},
		CustomerInfo: ServiceConfig{Timeout: 5 * time.Second},
		AuthService:  ServiceConfig{Timeout: 5 * time.Second},
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
	if v, ok := os.LookupEnv("INTERNAL_TOKENS"); ok {
		c.Server.InternalTokens = splitList(v)
	}
	setString(&c.Security.MasterKey, "CARD_MASTER_KEY")
	setString(&c.CustomerInfo.BaseURL, "CUSTOMER_INFO_URL")
	setString(&c.CustomerInfo.Token, "CUSTOMER_INFO_TOKEN")
	setString(&c.AuthService.BaseURL, "AUTH_SERVICE_URL")
	setString(&c.AuthService.Token, "AUTH_SERVICE_TOKEN")
	setString(&c.Storage.Backend, "STORAGE_BACKEND")

	setString(&c.Postgres.Host, "POSTGRES_HOST")
	setString(&c.Postgres.Port, "POSTGRES_PORT")
	setString(&c.Postgres.User, "POSTGRES_USER")
	setString(&c.Postgres.Password, "POSTGRES_PASSWORD")
	setString(&c.Postgres.Database, "POSTGRES_DB")
	setString(&c.Postgres.SSLMode, "POSTGRES_SSLMODE")
	if v, err := strconv.ParseInt(os.Getenv("POSTGRES_MAX_CONNS"), 10, 32); err == nil && v > 0 {
		c.Postgres.MaxConns = int32(v)
	}
	if v, err := time.ParseDuration(os.Getenv("POSTGRES_CONNECT_TIMEOUT")); err == nil && v > 0 {
		c.Postgres.ConnectTimeout = v
	}

	setString(&c.Limits.Timezone, "LIMITS_TIMEZONE")
}

func (c *Config) Validate() error {
	if len(c.Security.MasterKey) < minMasterKeyLen {
		return fmt.Errorf("CARD_MASTER_KEY must be at least %d characters", minMasterKeyLen)
	}
	if c.Server.PortHttpServer == "" {
		return fmt.Errorf("config server.portHttpServer is empty")
	}
	switch c.Storage.Backend {
	case "memory":
	case "postgres":
		p := c.Postgres
		if p.Host == "" || p.Port == "" || p.User == "" || p.Database == "" {
			return fmt.Errorf("config postgres.host, port, user and database are required for the postgres backend")
		}
		if p.MaxConns <= 0 || p.ConnectTimeout <= 0 {
			return fmt.Errorf("config postgres.maxConns and connect timeout must be positive")
		}
	default:
		return fmt.Errorf("config storage.backend must be memory or postgres, got %q", c.Storage.Backend)
	}
	for name, svc := range map[string]*ServiceConfig{"customerInfo": &c.CustomerInfo, "authService": &c.AuthService} {
		svc.BaseURL = strings.TrimRight(svc.BaseURL, "/")
		if svc.BaseURL == "" {
			continue
		}
		if u, err := url.Parse(svc.BaseURL); err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("config %s.baseUrl must be an absolute URL, got %q", name, svc.BaseURL)
		}
		if svc.Timeout <= 0 {
			return fmt.Errorf("config %s.timeout must be positive", name)
		}
	}
	loc, err := time.LoadLocation(c.Limits.Timezone)
	if err != nil {
		return fmt.Errorf("config limits.timezone: %w", err)
	}
	c.Limits.Location = loc
	return nil
}

func setString(dst *string, key string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
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
