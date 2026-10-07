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
)

var Version = "1.0.0"

const minHashKeyLen = 32

type Config struct {
	Server      ServerConfig      `json:"server"`
	Security    SecurityConfig    `json:"-"`
	Storage     StorageConfig     `json:"storage"`
	Postgres    PostgresConfig    `json:"postgres"`
	UserService UserServiceConfig `json:"userService"`
	KYC         KYCConfig         `json:"kyc"`
	EkycService UserServiceConfig `json:"ekycService"`
}

type ServerConfig struct {
	PortHttpServer string   `json:"portHttpServer"`
	AccessTokens   []string `json:"-"`
}

type SecurityConfig struct {
	DocumentHashKey string
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

type UserServiceConfig struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

type KYCConfig struct {
	MinAge           int     `json:"minAge"`
	MinConfidence    float64 `json:"minConfidence"`
	RequireFaceMatch bool    `json:"requireFaceMatch"`
}

func Defaults() *Config {
	return &Config{
		Server:  ServerConfig{PortHttpServer: "8080"},
		Storage: StorageConfig{Backend: "memory"},
		Postgres: PostgresConfig{
			Host:           "localhost",
			Port:           "5432",
			Database:       "customer_info",
			SSLMode:        "disable",
			MaxConns:       10,
			ConnectTimeout: 30 * time.Second,
		},
		UserService: UserServiceConfig{BaseURL: "http://localhost:1235", Timeout: 5 * time.Second},
		KYC:         KYCConfig{MinAge: 18, MinConfidence: 0.5, RequireFaceMatch: true},
		EkycService: UserServiceConfig{BaseURL: "http://localhost:8087", Timeout: 60 * time.Second},
	}
}

func (c *Config) Merge(raw []byte) error {
	if err := json.Unmarshal(raw, c); err != nil {
		return fmt.Errorf("parse config json: %w", err)
	}
	return nil
}

type duration struct{ d *time.Duration }

func (d duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration must be a string like \"5s\": %w", err)
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d.d = v
	return nil
}

func (u *UserServiceConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		BaseURL *string  `json:"baseUrl"`
		Timeout duration `json:"timeout"`
	}{BaseURL: &u.BaseURL, Timeout: duration{&u.Timeout}}
	return json.Unmarshal(b, &aux)
}

func (c *Config) ApplyEnv() {
	setString(&c.Server.PortHttpServer, "PORT_HTTP_SERVER")
	if v, ok := os.LookupEnv("ACCESS_TOKENS"); ok {
		c.Server.AccessTokens = splitList(v)
	}
	setString(&c.Security.DocumentHashKey, "DOCUMENT_HASH_KEY")
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
	setDuration(&c.Postgres.ConnectTimeout, "POSTGRES_CONNECT_TIMEOUT")

	setString(&c.UserService.BaseURL, "USER_SERVICE_URL")
	setString(&c.UserService.Token, "USER_SERVICE_TOKEN")
	setDuration(&c.UserService.Timeout, "USER_SERVICE_TIMEOUT")

	if v, err := strconv.Atoi(os.Getenv("KYC_MIN_AGE")); err == nil && v > 0 {
		c.KYC.MinAge = v
	}
	if v, err := strconv.ParseFloat(os.Getenv("KYC_MIN_CONFIDENCE"), 64); err == nil && v >= 0 {
		c.KYC.MinConfidence = v
	}

	if v, err := strconv.ParseBool(os.Getenv("KYC_REQUIRE_FACE_MATCH")); err == nil {
		c.KYC.RequireFaceMatch = v
	}

	setString(&c.EkycService.BaseURL, "EKYC_SERVICE_URL")
	setString(&c.EkycService.Token, "EKYC_SERVICE_TOKEN")
	setDuration(&c.EkycService.Timeout, "EKYC_SERVICE_TIMEOUT")
}

func (c *Config) Validate() error {
	if len(c.Security.DocumentHashKey) < minHashKeyLen {
		return fmt.Errorf("DOCUMENT_HASH_KEY must be at least %d characters", minHashKeyLen)
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
	for name, svc := range map[string]*UserServiceConfig{"userService": &c.UserService, "ekycService": &c.EkycService} {
		svc.BaseURL = strings.TrimRight(svc.BaseURL, "/")
		if u, err := url.Parse(svc.BaseURL); err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("config %s.baseUrl must be an absolute URL, got %q", name, svc.BaseURL)
		}
		if svc.Timeout <= 0 {
			return fmt.Errorf("config %s.timeout must be positive", name)
		}
	}
	if c.KYC.MinAge <= 0 || c.KYC.MinConfidence < 0 || c.KYC.MinConfidence > 1 {
		return fmt.Errorf("config kyc.minAge must be positive and kyc.minConfidence within [0,1]")
	}
	return nil
}

func setString(dst *string, key string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		*dst = v
	}
}

func setDuration(dst *time.Duration, key string) {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil && v > 0 {
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
