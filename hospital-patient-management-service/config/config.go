package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server       ServerConfig
	Postgres     PostgresConfig
	Auth         AuthConfig
	UserService  UpstreamConfig
	Billing      BillingConfig
	Wallet       WalletConfig
	Notification UpstreamConfig
	Ekyc         UpstreamConfig
	Upload       UpstreamConfig
}

// UpstreamConfig locates a sibling service. An empty BaseURL disables it.
type UpstreamConfig struct {
	BaseURL string
	Timeout time.Duration
}

const (
	AuthToken = "token" // every request must carry a valid user-service token
	AuthOff   = "off"   // no authentication (local development, or a trusted gateway in front)
)

type AuthConfig struct {
	Mode string
	// CacheTTL is how long a validated token is trusted without asking
	// user-service again, which is also how long a revoked token keeps
	// working. 0 checks every request.
	CacheTTL time.Duration
}

type BillingConfig struct {
	UpstreamConfig
	// PlanName is the billing-service plan every hospital invoice is filed under.
	PlanName string
}

type WalletConfig struct {
	UpstreamConfig
	// HospitalWalletID receives every patient payment.
	HospitalWalletID string
	// MinorUnit is how many minor units make one major unit of currency (100
	// for cents, 1 for VND); the wallet only deals in minor units.
	MinorUnit int64
	// UserPrefix + patient id is the patient's user_id in the wallet service.
	UserPrefix string
}

type ServerConfig struct {
	// HTTPPort serves the REST gateway and /health; GRPCPort serves gRPC.
	HTTPPort        string
	GRPCPort        string
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	return &Config{
		Server: ServerConfig{
			HTTPPort:        get("PORT", "8080"),
			GRPCPort:        get("GRPC_PORT", "9090"),
			ShutdownTimeout: durationOr(get("SHUTDOWN_TIMEOUT", "15s"), 15*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     get("POSTGRES_HOST", "localhost"),
			Port:     get("POSTGRES_PORT", "5432"),
			User:     get("POSTGRES_USER", "postgres"),
			Password: get("POSTGRES_PASSWORD", ""),
			DBName:   get("POSTGRES_DATABASE", "hospital_patient_management_service"),
			SSLMode:  get("POSTGRES_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			Mode:     strings.ToLower(get("AUTH_MODE", AuthToken)),
			CacheTTL: nonNegativeDuration(get("AUTH_CACHE_TTL", "5s"), 5*time.Second),
		},
		UserService: upstream("USER_SERVICE"),
		Billing: BillingConfig{
			UpstreamConfig: upstream("BILLING"),
			PlanName:       get("BILLING_PLAN_NAME", "Hospital services"),
		},
		Wallet: WalletConfig{
			UpstreamConfig:   upstream("WALLET"),
			HospitalWalletID: get("WALLET_HOSPITAL_ID", ""),
			MinorUnit:        int64Or(get("WALLET_MINOR_UNIT", "100"), 100),
			UserPrefix:       get("WALLET_USER_PREFIX", "patient-"),
		},
		Notification: upstream("NOTIFICATION"),
		Ekyc:         upstream("EKYC"),
		Upload:       upstream("UPLOAD"),
	}
}

// Validate rejects configurations that would fail on the first request.
func (c *Config) Validate() error {
	switch c.Auth.Mode {
	case AuthToken:
		if c.UserService.BaseURL == "" {
			return errors.New("AUTH_MODE=token needs USER_SERVICE_BASE_URL (or set AUTH_MODE=off)")
		}
	case AuthOff:
	default:
		return errors.New("AUTH_MODE must be 'token' or 'off'")
	}
	if c.Wallet.BaseURL != "" && c.Wallet.HospitalWalletID == "" {
		return errors.New("WALLET_BASE_URL needs WALLET_HOSPITAL_ID")
	}
	if c.Wallet.MinorUnit < 1 {
		return errors.New("WALLET_MINOR_UNIT must be at least 1")
	}
	return nil
}

func upstream(prefix string) UpstreamConfig {
	return UpstreamConfig{
		BaseURL: strings.TrimRight(get(prefix+"_BASE_URL", ""), "/"),
		Timeout: durationOr(get(prefix+"_TIMEOUT", "5s"), 5*time.Second),
	}
}

func int64Or(s string, def int64) int64 {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
		return n
	}
	return def
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func durationOr(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return def
}

func nonNegativeDuration(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d >= 0 {
		return d
	}
	if n, err := strconv.Atoi(s); err == nil && n >= 0 {
		return time.Duration(n) * time.Second
	}
	return def
}
