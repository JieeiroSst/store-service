package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

const checkoutUpstreamCalls = 8

type Config struct {
	Server   ServerConfig   `json:"server"`
	Postgres PostgresConfig `json:"postgres"`
	Vending  VendingConfig  `json:"vending"`
	Upstream UpstreamConfig `json:"upstream"`
}

type ServerConfig struct {
	PortHttpServer string `json:"portHttpServer"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
	MaxConns int    `json:"maxConns"`
}

type VendingConfig struct {
	SessionTTL        time.Duration
	ReservationTTL    time.Duration
	SweepInterval     time.Duration
	AlertInterval     time.Duration
	ReconcileInterval time.Duration
	ReconcileAfter    time.Duration
	Currency          string
}

type UpstreamConfig struct {
	ServiceName      string
	Timeout          time.Duration
	WalletServiceURL string
	CouponServiceURL string
	NotificationURL  string
}

func Defaults() *Config {
	return &Config{
		Server: ServerConfig{PortHttpServer: "8080"},
		Postgres: PostgresConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Dbname:   "vending_machine",
			SSLMode:  "disable",
			MaxConns: 10,
		},
		Vending: VendingConfig{
			SessionTTL:        5 * time.Minute,
			ReservationTTL:    2 * time.Minute,
			SweepInterval:     30 * time.Second,
			AlertInterval:     15 * time.Second,
			ReconcileInterval: time.Minute,
			ReconcileAfter:    2 * time.Minute,
			Currency:          "USD",
		},
		Upstream: UpstreamConfig{
			ServiceName: "vending-machine-service",
			Timeout:     5 * time.Second,
		},
	}
}

func FromEnv() *Config {
	cfg := Defaults()
	cfg.ApplyEnv()
	return cfg
}

func (c *Config) Merge(raw []byte) error {
	if err := json.Unmarshal(raw, c); err != nil {
		return fmt.Errorf("parse config json: %w", err)
	}
	return nil
}

func (c *Config) ApplyEnv() {
	setString(&c.Server.PortHttpServer, "PORT_HTTP_SERVER")

	setString(&c.Postgres.Host, "POSTGRES_HOST")
	setString(&c.Postgres.Port, "POSTGRES_PORT")
	setString(&c.Postgres.User, "POSTGRES_USER")
	setString(&c.Postgres.Password, "POSTGRES_PASSWORD")
	setString(&c.Postgres.Dbname, "POSTGRES_DBNAME")
	setString(&c.Postgres.SSLMode, "POSTGRES_SSLMODE")
	setInt(&c.Postgres.MaxConns, "POSTGRES_MAX_CONNS")

	setDuration(&c.Vending.SessionTTL, "SESSION_TTL")
	setDuration(&c.Vending.ReservationTTL, "RESERVATION_TTL")
	setDuration(&c.Vending.SweepInterval, "SWEEP_INTERVAL")
	setDuration(&c.Vending.AlertInterval, "ALERT_INTERVAL")
	setDuration(&c.Vending.ReconcileInterval, "RECONCILE_INTERVAL")
	setDuration(&c.Vending.ReconcileAfter, "RECONCILE_AFTER")
	setString(&c.Vending.Currency, "CURRENCY")

	setString(&c.Upstream.ServiceName, "SERVICE_NAME")
	setDuration(&c.Upstream.Timeout, "UPSTREAM_TIMEOUT")
	setString(&c.Upstream.WalletServiceURL, "WALLET_SERVICE_URL")
	setString(&c.Upstream.CouponServiceURL, "COUPON_SERVICE_URL")
	setString(&c.Upstream.NotificationURL, "NOTIFICATION_SERVICE_URL")
}

func (c *Config) Validate() error {
	positive := map[string]time.Duration{
		"vending.sessionTtl":        c.Vending.SessionTTL,
		"vending.reservationTtl":    c.Vending.ReservationTTL,
		"vending.sweepInterval":     c.Vending.SweepInterval,
		"vending.alertInterval":     c.Vending.AlertInterval,
		"vending.reconcileInterval": c.Vending.ReconcileInterval,
		"vending.reconcileAfter":    c.Vending.ReconcileAfter,
		"upstream.timeout":          c.Upstream.Timeout,
	}
	for name, d := range positive {
		if d <= 0 {
			return fmt.Errorf("config %s must be positive", name)
		}
	}
	if c.Postgres.MaxConns <= 0 {
		return fmt.Errorf("config postgres.maxConns must be positive")
	}
	if minimum := c.Upstream.Timeout * checkoutUpstreamCalls; c.Vending.ReconcileAfter <= minimum {
		return fmt.Errorf("config vending.reconcileAfter (%s) must exceed %s so a checkout still in flight is never reconciled",
			c.Vending.ReconcileAfter, minimum)
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

func (v *VendingConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		SessionTTL        duration `json:"sessionTtl"`
		ReservationTTL    duration `json:"reservationTtl"`
		SweepInterval     duration `json:"sweepInterval"`
		AlertInterval     duration `json:"alertInterval"`
		ReconcileInterval duration `json:"reconcileInterval"`
		ReconcileAfter    duration `json:"reconcileAfter"`
		Currency          *string  `json:"currency"`
	}{
		SessionTTL:        duration{&v.SessionTTL},
		ReservationTTL:    duration{&v.ReservationTTL},
		SweepInterval:     duration{&v.SweepInterval},
		AlertInterval:     duration{&v.AlertInterval},
		ReconcileInterval: duration{&v.ReconcileInterval},
		ReconcileAfter:    duration{&v.ReconcileAfter},
		Currency:          &v.Currency,
	}
	return json.Unmarshal(b, &aux)
}

func (u *UpstreamConfig) UnmarshalJSON(b []byte) error {
	aux := struct {
		ServiceName      *string  `json:"serviceName"`
		Timeout          duration `json:"timeout"`
		WalletServiceURL *string  `json:"walletServiceUrl"`
		CouponServiceURL *string  `json:"couponServiceUrl"`
		NotificationURL  *string  `json:"notificationServiceUrl"`
	}{
		ServiceName:      &u.ServiceName,
		Timeout:          duration{&u.Timeout},
		WalletServiceURL: &u.WalletServiceURL,
		CouponServiceURL: &u.CouponServiceURL,
		NotificationURL:  &u.NotificationURL,
	}
	return json.Unmarshal(b, &aux)
}

func setString(dst *string, key string) {
	if v, ok := os.LookupEnv(key); ok {
		*dst = v
	}
}

func setInt(dst *int, key string) {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		*dst = v
	}
}

func setDuration(dst *time.Duration, key string) {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil && v > 0 {
		*dst = v
	}
}
