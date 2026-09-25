package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                    string
	PostgresHost            string
	PostgresPort            string
	PostgresDatabase        string
	PostgresUser            string
	PostgresPassword        string
	PostgresSSLMode         string
	UserServiceURL          string
	UserServiceValidatePath string
	UserServiceUserPath     string
	UserServiceTimeout      time.Duration
	AdminRole               string
	AuthCacheTTL            time.Duration
	WalletServiceURL        string
	PaymentServiceURL       string
	PaymentServiceTimeout   time.Duration
	RecompenseServiceURL    string
	RecompenseCacheTTL      time.Duration
	BookingHoldTTL          time.Duration
	SweepInterval           time.Duration
}

func Load() (Config, error) {
	timeout, err := strconv.Atoi(get("UserServiceTimeoutSeconds", "5"))
	if err != nil {
		return Config{}, fmt.Errorf("UserServiceTimeoutSeconds: %w", err)
	}
	cacheTTL, err := strconv.Atoi(get("AuthCacheTTLSeconds", "30"))
	if err != nil {
		return Config{}, fmt.Errorf("AuthCacheTTLSeconds: %w", err)
	}
	holdMin, err := strconv.Atoi(get("BookingHoldMinutes", "15"))
	if err != nil || holdMin <= 0 {
		return Config{}, fmt.Errorf("BookingHoldMinutes must be a positive integer")
	}
	sweepSec, err := strconv.Atoi(get("SweepIntervalSeconds", "60"))
	if err != nil || sweepSec <= 0 {
		return Config{}, fmt.Errorf("SweepIntervalSeconds must be a positive integer")
	}
	payTimeout, err := strconv.Atoi(get("PaymentServiceTimeoutSeconds", "15"))
	if err != nil {
		return Config{}, fmt.Errorf("PaymentServiceTimeoutSeconds: %w", err)
	}
	c := Config{
		Port:                    get("PORT", "8080"),
		PostgresHost:            get("HostPostgres", "localhost"),
		PostgresPort:            get("PortPostgres", "5432"),
		PostgresDatabase:        get("DatabasePostgres", "rent_house_service"),
		PostgresUser:            get("UserPostgres", "postgres"),
		PostgresPassword:        get("PasswordPostgres", ""),
		PostgresSSLMode:         get("SSLModePostgres", "disable"),
		UserServiceURL:          get("UserServiceURL", ""),
		UserServiceValidatePath: get("UserServiceValidatePath", "/api/v1/validate"),
		UserServiceUserPath:     get("UserServiceUserPath", "/user"),
		UserServiceTimeout:      time.Duration(timeout) * time.Second,
		AdminRole:               get("AdminRole", "admin"),
		AuthCacheTTL:            time.Duration(cacheTTL) * time.Second,
		WalletServiceURL:        get("WalletServiceURL", ""),
		PaymentServiceURL:       get("PaymentServiceURL", ""),
		PaymentServiceTimeout:   time.Duration(payTimeout) * time.Second,
		RecompenseServiceURL:    get("RecompenseServiceURL", ""),
		RecompenseCacheTTL:      5 * time.Minute,
		BookingHoldTTL:          time.Duration(holdMin) * time.Minute,
		SweepInterval:           time.Duration(sweepSec) * time.Second,
	}
	if c.UserServiceURL == "" {
		return Config{}, fmt.Errorf("UserServiceURL must be set")
	}
	if c.WalletServiceURL == "" && c.PaymentServiceURL == "" {
		return Config{}, fmt.Errorf("set WalletServiceURL and/or PaymentServiceURL: bookings need a way to be paid")
	}
	return c, nil
}

func (c Config) PostgresDSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:     net.JoinHostPort(c.PostgresHost, c.PostgresPort),
		Path:     c.PostgresDatabase,
		RawQuery: "sslmode=" + url.QueryEscape(c.PostgresSSLMode),
	}
	return u.String()
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
