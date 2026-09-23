package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server       ServerConfig
	Mysql        MysqlConfig
	Wallet       WalletConfig
	Notification NotificationConfig
	UserService  UserServiceConfig
	Referral     ReferralConfig
	Redis        RedisConfig
	Fees         FeeConfig
	Exchange     ExchangeConfig
	Admin        AdminConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
}

type WalletConfig struct {
	BaseURL          string
	Timeout          string
	TreasuryWalletID string
}

type NotificationConfig struct {
	BaseURL string
	Timeout string
}

type ExchangeConfig struct {
	Currency      string
	ShareValue    int64
	MinOrderSize  int64
	DisputeWindow string
	DisputeBond   int64
	RewardEpoch   string
}

type UserServiceConfig struct {
	BaseURL  string
	Timeout  string
	AuthMode string
}

type ReferralConfig struct {
	BaseURL string
	Timeout string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type FeeConfig struct {
	TakerFeeBps    int64
	MakerRebateBps int64
	ReferralBps    int64
}

func (c UserServiceConfig) TimeoutDuration() time.Duration { return parseTimeout(c.Timeout) }
func (c ReferralConfig) TimeoutDuration() time.Duration    { return parseTimeout(c.Timeout) }

func (c ExchangeConfig) RewardEpochDuration() time.Duration {
	if d, err := time.ParseDuration(c.RewardEpoch); err == nil && d >= time.Second {
		return d
	}
	return time.Minute
}

func (c ExchangeConfig) DisputeWindowDuration() time.Duration {
	if d, err := time.ParseDuration(c.DisputeWindow); err == nil && d > 0 {
		return d
	}
	return 2 * time.Hour
}

type AdminConfig struct {
	Token string
}

func (c WalletConfig) TimeoutDuration() time.Duration       { return parseTimeout(c.Timeout) }
func (c NotificationConfig) TimeoutDuration() time.Duration { return parseTimeout(c.Timeout) }

func parseTimeout(raw string) time.Duration {
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	return 10 * time.Second
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8091"),
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "polymarket_service"),
		},
		Wallet: WalletConfig{
			BaseURL:          getEnv("WALLET_BASE_URL", "http://localhost:8088"),
			Timeout:          getEnv("WALLET_TIMEOUT", "10s"),
			TreasuryWalletID: getEnv("TREASURY_WALLET_ID", ""),
		},
		Notification: NotificationConfig{
			BaseURL: getEnv("NOTIFICATION_BASE_URL", "http://localhost:1235"),
			Timeout: getEnv("NOTIFICATION_TIMEOUT", "5s"),
		},
		Exchange: ExchangeConfig{
			Currency:      getEnv("EXCHANGE_CURRENCY", "USD"),
			ShareValue:    getEnvInt("SHARE_VALUE", 100),
			MinOrderSize:  getEnvInt("MIN_ORDER_SIZE", 1),
			DisputeWindow: getEnv("DISPUTE_WINDOW", "2h"),
			DisputeBond:   getEnvInt("DISPUTE_BOND", 0),
			RewardEpoch:   getEnv("REWARD_EPOCH", "1m"),
		},
		UserService: UserServiceConfig{
			BaseURL:  getEnv("USER_SERVICE_BASE_URL", ""),
			Timeout:  getEnv("USER_SERVICE_TIMEOUT", "5s"),
			AuthMode: getEnv("AUTH_MODE", "gateway"),
		},
		Referral: ReferralConfig{
			BaseURL: getEnv("REFERRAL_BASE_URL", ""),
			Timeout: getEnv("REFERRAL_TIMEOUT", "5s"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", ""),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       int(getEnvInt("REDIS_DB", 0)),
		},
		Fees: FeeConfig{
			TakerFeeBps:    getEnvInt("TAKER_FEE_BPS", 0),
			MakerRebateBps: getEnvInt("MAKER_REBATE_BPS", 2000),
			ReferralBps:    getEnvInt("REFERRAL_BPS", 1000),
		},
		Admin: AdminConfig{Token: getEnv("ADMIN_TOKEN", "")},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int64) int64 {
	if n, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil {
		return n
	}
	return fallback
}
