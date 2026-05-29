package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(Load),
)

type Config struct {
	App      AppConfig
	AWS      AWSConfig
	DynamoDB DynamoDBConfig
	DeepLink DeepLinkConfig
	Referral ReferralConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	Env     string
	Port    int
	Name    string
	Version string
}

type AWSConfig struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	DynamoDBEndpoint string 
}

type DynamoDBConfig struct {
	TableReferralLinks  string
	TableReferralEvents string
	TableRewards        string
	TableUserStats      string
}

type DeepLinkConfig struct {
	BaseURL      string
	AppStoreURL  string
	PlayStoreURL string
	AppURLScheme string // e.g. "yourapp://open" — used to attempt opening installed app
}

type ReferralConfig struct {
	TTLDays   int
	MaxPerDay int
}

type LoggerConfig struct {
	Level string
	FilePath string
	MaxSizeMB int
	MaxBackups int
	MaxAgeDays int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("config: APP_PORT must be integer: %w", err)
	}

	ttl, err := strconv.Atoi(getEnv("REFERRAL_TTL_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("config: REFERRAL_TTL_DAYS must be integer: %w", err)
	}

	maxPerDay, err := strconv.Atoi(getEnv("MAX_REFERRAL_PER_DAY", "50"))
	if err != nil {
		return nil, fmt.Errorf("config: MAX_REFERRAL_PER_DAY must be integer: %w", err)
	}

	logMaxSize, err := strconv.Atoi(getEnv("LOG_MAX_SIZE_MB", "100"))
	if err != nil {
		return nil, fmt.Errorf("config: LOG_MAX_SIZE_MB must be integer: %w", err)
	}

	logMaxBackups, err := strconv.Atoi(getEnv("LOG_MAX_BACKUPS", "7"))
	if err != nil {
		return nil, fmt.Errorf("config: LOG_MAX_BACKUPS must be integer: %w", err)
	}

	logMaxAge, err := strconv.Atoi(getEnv("LOG_MAX_AGE_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("config: LOG_MAX_AGE_DAYS must be integer: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Port:    port,
			Name:    getEnv("APP_NAME", "referral-service"),
			Version: getEnv("APP_VERSION", "0.0.0"),
		},
		AWS: AWSConfig{
			Region:           requireEnv("AWS_REGION"),
			AccessKeyID:      getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey:  getEnv("AWS_SECRET_ACCESS_KEY", ""),
			DynamoDBEndpoint: getEnv("DYNAMODB_ENDPOINT", ""),
		},
		DynamoDB: DynamoDBConfig{
			TableReferralLinks:  getEnv("TABLE_REFERRAL_LINKS", "referral_links"),
			TableReferralEvents: getEnv("TABLE_REFERRAL_EVENTS", "referral_events"),
			TableRewards:        getEnv("TABLE_REFERRAL_REWARDS", "referral_rewards"),
			TableUserStats:      getEnv("TABLE_USER_STATS", "user_referral_stats"),
		},
		DeepLink: DeepLinkConfig{
			BaseURL:      requireEnv("DEEP_LINK_BASE_URL"),
			AppStoreURL:  requireEnv("APP_STORE_URL"),
			PlayStoreURL: requireEnv("PLAY_STORE_URL"),
			AppURLScheme: getEnv("APP_URL_SCHEME", ""),
		},
		Referral: ReferralConfig{
			TTLDays:   ttl,
			MaxPerDay: maxPerDay,
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			FilePath:   getEnv("LOG_FILE_PATH", ""),
			MaxSizeMB:  logMaxSize,
			MaxBackups: logMaxBackups,
			MaxAgeDays: logMaxAge,
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required env var %q is not set", key))
	}
	return v
}
