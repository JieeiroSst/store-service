package config

import (
	"crypto/sha256"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	PostgresSSLMode  string

	RedisHost string
	RedisPort string
	RedisDB   int

	KafkaBrokers []string

	JWTSecret     string
	JWTExpiration time.Duration

	PartnerHMACEncKey []byte

	HostConsul    string
	KeyConsul     string
	ServiceConsul string
	ConsulEnabled bool

	OutboxRelayInterval        time.Duration
	VoucherExpirySweepInterval time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port: getEnv("PORT", "3000"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "voucher"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "voucher"),
		PostgresDBName:   getEnv("POSTGRES_DBNAME", "voucher"),
		PostgresSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		RedisDB:   getEnvInt("REDIS_DB", 0),

		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},

		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiration: time.Duration(getEnvInt("JWT_EXPIRATION_MINUTES", 60)) * time.Minute,

		PartnerHMACEncKey: derivePartnerEncKey(getEnv("PARTNER_HMAC_ENC_KEY", "dev-partner-hmac-key-change-me")),

		HostConsul:    getEnv("HostConsul", ""),
		KeyConsul:     getEnv("KeyConsul", "voucher_service"),
		ServiceConsul: getEnv("ServiceConsul", "consul"),
		ConsulEnabled: getEnvBool("CONSUL_ENABLED", false),

		OutboxRelayInterval:        time.Duration(getEnvInt("OUTBOX_RELAY_INTERVAL_MS", 500)) * time.Millisecond,
		VoucherExpirySweepInterval: time.Duration(getEnvInt("VOUCHER_EXPIRY_SWEEP_MINUTES", 5)) * time.Minute,
	}
	return cfg, nil
}
func derivePartnerEncKey(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
