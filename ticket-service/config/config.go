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
	PostgresMaxConns int

	UserServiceURL          string
	UserServiceValidatePath string // POST, body {"session_token": ...}
	UserServiceUserPath     string // GET, ?user_id=N
	UserServiceTimeout      time.Duration
	AdminRole               string
	AuthCacheTTL            time.Duration

	WalletServiceURL       string
	PaymentServiceURL      string
	PaymentServiceTimeout  time.Duration
	NotificationServiceURL string

	UploadServiceURL string
	UploadServiceKey string

	TemporalAddress   string
	TemporalNamespace string
	TemporalTaskQueue string

	TimeZone string

	ReserveConcurrency int
	ReserveQueueWait   time.Duration
	SoldOutTTL         time.Duration
	UserRatePerSecond  float64
	UserBurst          int
	MaxPendingPerUser  int
	MaxInFlight        int
	PublicCacheTTL     time.Duration

	InternalAPIKey string
	TransferTTL    time.Duration
	MaxTransfers   int

	InvoiceSellerName    string
	InvoiceSellerTaxID   string
	InvoiceSellerAddress string
	VATPercent           int

	ResaleFeePercent int
	PlatformWalletID string

	HoldTTL       time.Duration
	SweepInterval time.Duration
}

func Load() (Config, error) {
	var firstErr error
	num := func(key string, def, min int) int {
		v, err := strconv.Atoi(get(key, strconv.Itoa(def)))
		if err != nil || v < min {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s must be an integer >= %d", key, min)
			}
			return def
		}
		return v
	}
	c := Config{
		Port:                    get("PORT", "8080"),
		PostgresHost:            get("HostPostgres", "localhost"),
		PostgresPort:            get("PortPostgres", "5432"),
		PostgresDatabase:        get("DatabasePostgres", "ticket_service"),
		PostgresUser:            get("UserPostgres", "postgres"),
		PostgresPassword:        get("PasswordPostgres", ""),
		PostgresSSLMode:         get("SSLModePostgres", "disable"),
		PostgresMaxConns:        num("PostgresMaxConns", 10, 1),
		UserServiceURL:          get("UserServiceURL", ""),
		UserServiceValidatePath: get("UserServiceValidatePath", "/api/v1/validate"),
		UserServiceUserPath:     get("UserServiceUserPath", "/user"),
		UserServiceTimeout:      time.Duration(num("UserServiceTimeoutSeconds", 5, 1)) * time.Second,
		AdminRole:               get("AdminRole", "admin"),
		AuthCacheTTL:            time.Duration(num("AuthCacheTTLSeconds", 30, 0)) * time.Second,
		WalletServiceURL:        get("WalletServiceURL", ""),
		PaymentServiceURL:       get("PaymentServiceURL", ""),
		PaymentServiceTimeout:   time.Duration(num("PaymentServiceTimeoutSeconds", 15, 1)) * time.Second,
		NotificationServiceURL:  get("NotificationServiceURL", ""),
		TimeZone:                get("TimeZone", "Asia/Ho_Chi_Minh"),
		UploadServiceURL:        get("UploadServiceURL", ""),
		UploadServiceKey:        get("UploadServiceKey", ""),
		TemporalAddress:         get("TemporalAddress", ""),
		TemporalNamespace:       get("TemporalNamespace", "default"),
		TemporalTaskQueue:       get("TemporalTaskQueue", "ticket-service-orders"),
		ReserveConcurrency:      num("ReserveConcurrency", 16, 0),
		ReserveQueueWait:        time.Duration(num("ReserveQueueWaitMillis", 250, 0)) * time.Millisecond,
		SoldOutTTL:              time.Duration(num("SoldOutTTLSeconds", 2, 0)) * time.Second,
		UserRatePerSecond:       float64(num("UserRatePerSecond", 2, 0)),
		UserBurst:               num("UserBurst", 10, 0),
		MaxPendingPerUser:       num("MaxPendingPerUser", 3, 0),
		MaxInFlight:             num("MaxInFlight", 10000, 0),
		PublicCacheTTL:          time.Duration(num("PublicCacheTTLSeconds", 2, 0)) * time.Second,
		InternalAPIKey:          get("InternalAPIKey", ""),
		TransferTTL:             time.Duration(num("TransferTTLHours", 48, 1)) * time.Hour,
		MaxTransfers:            num("MaxTicketTransfers", 3, 0),
		InvoiceSellerName:       get("InvoiceSellerName", "Ticket Service"),
		InvoiceSellerTaxID:      get("InvoiceSellerTaxID", ""),
		InvoiceSellerAddress:    get("InvoiceSellerAddress", ""),
		VATPercent:              num("VATPercent", 0, 0),
		ResaleFeePercent:        num("ResaleFeePercent", 10, 0),
		PlatformWalletID:        get("PlatformWalletID", ""),
		HoldTTL:                 time.Duration(num("OrderHoldMinutes", 10, 1)) * time.Minute,
		SweepInterval:           time.Duration(num("SweepIntervalSeconds", 30, 1)) * time.Second,
	}
	if firstErr != nil {
		return Config{}, firstErr
	}
	if c.UploadServiceURL != "" && c.UploadServiceKey == "" {
		return Config{}, fmt.Errorf("UploadServiceURL needs UploadServiceKey")
	}
	if c.UserServiceURL == "" {
		return Config{}, fmt.Errorf("UserServiceURL must be set")
	}
	if c.WalletServiceURL == "" && c.PaymentServiceURL == "" {
		return Config{}, fmt.Errorf("set WalletServiceURL and/or PaymentServiceURL: orders need a way to be paid")
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
