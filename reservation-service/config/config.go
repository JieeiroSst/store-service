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
	Port string

	PostgresHost     string
	PostgresPort     string
	PostgresDatabase string
	PostgresUser     string
	PostgresPassword string
	PostgresSSLMode  string

	// user-service owns accounts, login, sessions and roles (guests, hotel managers, admins).
	UserServiceURL          string
	UserServiceValidatePath string // POST, body {"session_token": ...}
	UserServiceUserPath     string // GET, ?user_id=N
	UserServiceTimeout      time.Duration
	// Users holding this role in user-service are platform admins: they verify hotels.
	AdminRole string
	// How long a resolved token is cached; 0 disables. Revocation and role changes in
	// user-service take up to this long to be noticed.
	AuthCacheTTL time.Duration

	// payment-wallet-service: internal e-wallets. Empty disables wallet payment.
	WalletServiceURL string
	// payment_service: external providers (paypal, stripe, ...). Empty disables gateway payment.
	PaymentServiceURL     string
	PaymentServiceTimeout time.Duration

	// recompense-service: points and tiers. Empty turns off host ranking boosts and points.
	RecompenseServiceURL string
	RecompenseCacheTTL   time.Duration

	// notification-service: delivers notifications to people's devices. Empty keeps them in the inbox only.
	NotificationServiceURL string

	// Postgres connections this process may open. Every replica opens its own pool, so replicas x PostgresMaxConns
	// (plus the connections of everything else on that server) must stay under Postgres max_connections (100 by
	// default), or the surplus replicas fail with "too many clients".
	PostgresMaxConns int
	// Protection for a rush on the last rooms (see internal/application/service/hotpath.go).
	// ReserveConcurrency requests at a time may reach the database; others wait ReserveQueueWait, then get 429.
	ReserveConcurrency int
	ReserveQueueWait   time.Duration
	// SoldOutTTL: how long "sold out" is remembered without asking the database again (0 disables).
	SoldOutTTL time.Duration
	// One account may attempt UserRatePerSecond reservations a second (burst UserBurst).
	UserRatePerSecond float64
	UserBurst         int
	// MaxPendingPerGuest: unpaid reservations one guest may hold at once (0 = unlimited).
	MaxPendingPerGuest int
	// MaxInFlight: requests handled at once by this process; beyond it, 503 with Retry-After.
	MaxInFlight int

	// How long rooms stay held for an unpaid reservation.
	HoldTTL time.Duration
	// How often expired holds are released.
	SweepInterval time.Duration
}

func Load() (Config, error) {
	atoi := func(key, def string) (int, error) {
		v, err := strconv.Atoi(get(key, def))
		if err != nil {
			return 0, fmt.Errorf("%s: %w", key, err)
		}
		return v, nil
	}
	timeout, err := atoi("UserServiceTimeoutSeconds", "5")
	if err != nil {
		return Config{}, err
	}
	cacheTTL, err := atoi("AuthCacheTTLSeconds", "30")
	if err != nil {
		return Config{}, err
	}
	holdMin, err := atoi("ReservationHoldMinutes", "15")
	if err != nil || holdMin <= 0 {
		return Config{}, fmt.Errorf("ReservationHoldMinutes must be a positive integer")
	}
	sweepSec, err := atoi("SweepIntervalSeconds", "60")
	if err != nil || sweepSec <= 0 {
		return Config{}, fmt.Errorf("SweepIntervalSeconds must be a positive integer")
	}
	payTimeout, err := atoi("PaymentServiceTimeoutSeconds", "15")
	if err != nil {
		return Config{}, err
	}
	intOf := func(key, def string, min int) (int, error) {
		v, err := atoi(key, def)
		if err != nil || v < min {
			return 0, fmt.Errorf("%s must be an integer >= %d", key, min)
		}
		return v, nil
	}
	maxConns, err := intOf("PostgresMaxConns", "10", 1)
	if err != nil {
		return Config{}, err
	}
	reserveConc, err := intOf("ReserveConcurrency", "16", 0)
	if err != nil {
		return Config{}, err
	}
	reserveWait, err := intOf("ReserveQueueWaitMillis", "250", 0)
	if err != nil {
		return Config{}, err
	}
	soldOutSec, err := intOf("SoldOutTTLSeconds", "2", 0)
	if err != nil {
		return Config{}, err
	}
	userRate, err := intOf("UserRatePerSecond", "2", 0)
	if err != nil {
		return Config{}, err
	}
	userBurst, err := intOf("UserBurst", "10", 0)
	if err != nil {
		return Config{}, err
	}
	maxPending, err := intOf("MaxPendingPerGuest", "3", 0)
	if err != nil {
		return Config{}, err
	}
	maxInFlight, err := intOf("MaxInFlight", "10000", 0)
	if err != nil {
		return Config{}, err
	}
	c := Config{
		Port:                    get("PORT", "8080"),
		PostgresHost:            get("HostPostgres", "localhost"),
		PostgresPort:            get("PortPostgres", "5432"),
		PostgresDatabase:        get("DatabasePostgres", "reservation_service"),
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
		NotificationServiceURL:  get("NotificationServiceURL", ""),
		PostgresMaxConns:        maxConns,
		ReserveConcurrency:      reserveConc,
		ReserveQueueWait:        time.Duration(reserveWait) * time.Millisecond,
		SoldOutTTL:              time.Duration(soldOutSec) * time.Second,
		UserRatePerSecond:       float64(userRate),
		UserBurst:               userBurst,
		MaxPendingPerGuest:      maxPending,
		MaxInFlight:             maxInFlight,
		HoldTTL:                 time.Duration(holdMin) * time.Minute,
		SweepInterval:           time.Duration(sweepSec) * time.Second,
	}
	if c.UserServiceURL == "" {
		return Config{}, fmt.Errorf("UserServiceURL must be set")
	}
	if c.WalletServiceURL == "" && c.PaymentServiceURL == "" {
		return Config{}, fmt.Errorf("set WalletServiceURL and/or PaymentServiceURL: reservations need a way to be paid")
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
