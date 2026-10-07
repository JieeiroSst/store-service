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

type Config struct {
	Server      ServerConfig   `json:"server"`
	Storage     StorageConfig  `json:"storage"`
	Postgres    PostgresConfig `json:"postgres"`
	UserService ServiceConfig  `json:"userService"`
	CardService ServiceConfig  `json:"cardService"`

	AuthorizeService AuthorizeServiceConfig `json:"authorizeService"`
	Challenge        ChallengeConfig        `json:"challenge"`
	OTPDelivery      OTPDeliveryConfig      `json:"otpDelivery"`
}

type ServerConfig struct {
	PortHttpServer  string   `json:"portHttpServer"`
	VerifyPerMinute int      `json:"verifyPerMinute"`
	AccessTokens    []string `json:"-"`
	InternalTokens  []string `json:"-"`
}

type AuthorizeServiceConfig struct {
	GRPCAddress string
	Timeout     time.Duration
}

func (a *AuthorizeServiceConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		GRPCAddress *string  `json:"grpcAddress"`
		Timeout     duration `json:"timeout"`
	}{GRPCAddress: &a.GRPCAddress, Timeout: duration{&a.Timeout}}
	return json.Unmarshal(b, &aux)
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

type ServiceConfig struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

type ChallengeConfig struct {
	OTPTTL       time.Duration
	MaxAttempts  int
	MaxResends   int
	ValidityTime time.Duration
	ExposeOTP    bool
}

type OTPDeliveryConfig struct {
	Backend      string
	WebhookURL   string
	WebhookToken string
	Timeout      time.Duration
}

func Defaults() *Config {
	return &Config{
		Server:  ServerConfig{PortHttpServer: "8080", VerifyPerMinute: 30},
		Storage: StorageConfig{Backend: "memory"},
		Postgres: PostgresConfig{
			Host:           "localhost",
			Port:           "5432",
			Database:       "payment_auth",
			SSLMode:        "disable",
			MaxConns:       10,
			ConnectTimeout: 30 * time.Second,
		},
		UserService: ServiceConfig{BaseURL: "http://localhost:1235", Timeout: 5 * time.Second},
		CardService: ServiceConfig{BaseURL: "http://localhost:8080", Timeout: 5 * time.Second},
		Challenge:   ChallengeConfig{OTPTTL: 5 * time.Minute, MaxAttempts: 3, MaxResends: 3, ValidityTime: 10 * time.Minute},

		AuthorizeService: AuthorizeServiceConfig{GRPCAddress: "localhost:9090", Timeout: 5 * time.Second},
		OTPDelivery:      OTPDeliveryConfig{Backend: "log", Timeout: 5 * time.Second},
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
		return fmt.Errorf("duration must be a string like \"5m\": %w", err)
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d.d = v
	return nil
}

func (s *ServiceConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		BaseURL *string  `json:"baseUrl"`
		Timeout duration `json:"timeout"`
	}{BaseURL: &s.BaseURL, Timeout: duration{&s.Timeout}}
	return json.Unmarshal(b, &aux)
}

func (c *ChallengeConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		OTPTTL       duration `json:"otpTtl"`
		MaxAttempts  *int     `json:"maxAttempts"`
		MaxResends   *int     `json:"maxResends"`
		ValidityTime duration `json:"validityTime"`
	}{OTPTTL: duration{&c.OTPTTL}, MaxAttempts: &c.MaxAttempts, MaxResends: &c.MaxResends, ValidityTime: duration{&c.ValidityTime}}
	return json.Unmarshal(b, &aux)
}

func (o *OTPDeliveryConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		Backend    *string  `json:"backend"`
		WebhookURL *string  `json:"webhookUrl"`
		Timeout    duration `json:"timeout"`
	}{Backend: &o.Backend, WebhookURL: &o.WebhookURL, Timeout: duration{&o.Timeout}}
	return json.Unmarshal(b, &aux)
}

func (c *Config) ApplyEnv() {
	setString(&c.Server.PortHttpServer, "PORT_HTTP_SERVER")
	if v, ok := os.LookupEnv("ACCESS_TOKENS"); ok {
		c.Server.AccessTokens = splitList(v)
	}
	if v, ok := os.LookupEnv("INTERNAL_TOKENS"); ok {
		c.Server.InternalTokens = splitList(v)
	}
	if v, err := strconv.Atoi(os.Getenv("VERIFY_PER_MINUTE")); err == nil && v > 0 {
		c.Server.VerifyPerMinute = v
	}
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
	setString(&c.AuthorizeService.GRPCAddress, "AUTHORIZE_SERVICE_GRPC_ADDR")
	setDuration(&c.AuthorizeService.Timeout, "AUTHORIZE_SERVICE_TIMEOUT")
	setString(&c.CardService.BaseURL, "CARD_SERVICE_URL")
	setString(&c.CardService.Token, "CARD_SERVICE_TOKEN")
	setDuration(&c.CardService.Timeout, "CARD_SERVICE_TIMEOUT")

	setDuration(&c.Challenge.OTPTTL, "OTP_TTL")
	setDuration(&c.Challenge.ValidityTime, "AUTHENTICATION_VALIDITY")
	if v, err := strconv.Atoi(os.Getenv("OTP_MAX_ATTEMPTS")); err == nil && v > 0 {
		c.Challenge.MaxAttempts = v
	}
	if v, err := strconv.Atoi(os.Getenv("OTP_MAX_RESENDS")); err == nil && v >= 0 {
		c.Challenge.MaxResends = v
	}
	if v, err := strconv.ParseBool(os.Getenv("EXPOSE_OTP")); err == nil {
		c.Challenge.ExposeOTP = v
	}

	setString(&c.OTPDelivery.Backend, "OTP_DELIVERY")
	setString(&c.OTPDelivery.WebhookURL, "OTP_WEBHOOK_URL")
	setString(&c.OTPDelivery.WebhookToken, "OTP_WEBHOOK_TOKEN")
	setDuration(&c.OTPDelivery.Timeout, "OTP_WEBHOOK_TIMEOUT")
}

func (c *Config) Validate() error {
	if c.Server.PortHttpServer == "" || c.Server.VerifyPerMinute <= 0 {
		return fmt.Errorf("config server.portHttpServer is empty or server.verifyPerMinute is not positive")
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
	for name, s := range map[string]*ServiceConfig{"userService": &c.UserService, "cardService": &c.CardService} {
		s.BaseURL = strings.TrimRight(s.BaseURL, "/")
		if u, err := url.Parse(s.BaseURL); err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("config %s.baseUrl must be an absolute URL, got %q", name, s.BaseURL)
		}
		if s.Timeout <= 0 {
			return fmt.Errorf("config %s.timeout must be positive", name)
		}
	}
	ch := c.Challenge
	if ch.OTPTTL <= 0 || ch.ValidityTime <= 0 || ch.MaxAttempts <= 0 || ch.MaxResends < 0 {
		return fmt.Errorf("config challenge.otpTtl, validityTime and maxAttempts must be positive and maxResends not negative")
	}
	if a := c.AuthorizeService; a.GRPCAddress == "" || a.Timeout <= 0 {
		return fmt.Errorf("config authorizeService.grpcAddress is required and authorizeService.timeout must be positive")
	}
	switch c.OTPDelivery.Backend {
	case "log":
	case "webhook":
		if u, err := url.Parse(c.OTPDelivery.WebhookURL); err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("config otpDelivery.webhookUrl must be an absolute URL for the webhook backend")
		}
		if c.OTPDelivery.Timeout <= 0 {
			return fmt.Errorf("config otpDelivery.timeout must be positive")
		}
	default:
		return fmt.Errorf("config otpDelivery.backend must be log or webhook, got %q", c.OTPDelivery.Backend)
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
