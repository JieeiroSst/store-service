package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Email    EmailConfig
	SMS      SMSConfig
	Firebase FirebaseConfig
	Worker   WorkerConfig
}

type ServerConfig struct {
	Port      string
	Mode      string // development | production
	APIKey    string
	RateLimit int
}

type DatabaseConfig struct {
	DSN             string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // seconds
}

type EmailConfig struct {
	Provider    string // smtp | sendgrid
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	SMTPPass    string
	FromAddr    string
	SendGridKey string
}

type SMSConfig struct {
	Provider     string // twilio | vonage
	TwilioSID    string
	TwilioToken  string
	TwilioFrom   string
	VonageKey    string
	VonageSecret string
	VonageFrom   string
}

type FirebaseConfig struct {
	CredentialsFile string
	ProjectID       string
}

type WorkerConfig struct {
	PoolSize        int
	QueueSize       int
	RetryMax        int
	RetryDelay      int // milliseconds
	FetcherPoolSize int
	FetchTimeoutSec int
}

func (d DatabaseConfig) DataSourceName() string {
	if d.DSN != "" {
		return d.DSN
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC&clientFoundRows=true",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

func (d DatabaseConfig) ConnMaxLifetimeDuration() time.Duration {
	return time.Duration(d.ConnMaxLifetime) * time.Second
}

func (w WorkerConfig) RetryDelayDuration() time.Duration {
	return time.Duration(w.RetryDelay) * time.Millisecond
}

func (w WorkerConfig) FetchTimeout() time.Duration {
	return time.Duration(w.FetchTimeoutSec) * time.Second
}

func Load(path string) *Config {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("env")
	v.AutomaticEnv()
	_ = v.ReadInConfig()

	str := func(key, def string) string {
		if s := v.GetString(key); s != "" {
			return s
		}
		return def
	}
	num := func(key string, def int) int {
		if n := v.GetInt(key); n > 0 {
			return n
		}
		return def
	}

	return &Config{
		Server: ServerConfig{
			Port:      str("PORT", "8095"),
			Mode:      str("MODE", "production"),
			APIKey:    str("API_KEY", v.GetString("JWT_SECRET")),
			RateLimit: num("RATE_LIMIT", 100),
		},
		Database: DatabaseConfig{
			DSN:             v.GetString("DB_DSN"),
			Host:            str("MYSQL_HOST", "localhost"),
			Port:            str("MYSQL_PORT", "3306"),
			User:            str("MYSQL_USER", "root"),
			Password:        v.GetString("MYSQL_PASSWORD"),
			Name:            str("MYSQL_DBNAME", "notifyhub"),
			MaxOpenConns:    num("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:    num("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: num("DB_CONN_MAX_LIFETIME", 300),
		},
		Email: EmailConfig{
			Provider:    str("EMAIL_PROVIDER", "smtp"),
			SMTPHost:    v.GetString("SMTP_HOST"),
			SMTPPort:    num("SMTP_PORT", 587),
			SMTPUser:    v.GetString("SMTP_USER"),
			SMTPPass:    v.GetString("SMTP_PASS"),
			FromAddr:    v.GetString("EMAIL_FROM"),
			SendGridKey: v.GetString("SENDGRID_API_KEY"),
		},
		SMS: SMSConfig{
			Provider:     v.GetString("SMS_PROVIDER"),
			TwilioSID:    v.GetString("TWILIO_SID"),
			TwilioToken:  v.GetString("TWILIO_TOKEN"),
			TwilioFrom:   v.GetString("TWILIO_FROM"),
			VonageKey:    v.GetString("VONAGE_API_KEY"),
			VonageSecret: v.GetString("VONAGE_API_SECRET"),
			VonageFrom:   str("VONAGE_FROM", v.GetString("TWILIO_FROM")),
		},
		Firebase: FirebaseConfig{
			CredentialsFile: v.GetString("FIREBASE_CREDENTIALS_FILE"),
			ProjectID:       v.GetString("FIREBASE_PROJECT_ID"),
		},
		Worker: WorkerConfig{
			PoolSize:        num("WORKER_POOL_SIZE", 10),
			QueueSize:       num("WORKER_QUEUE_SIZE", 1000),
			RetryMax:        num("WORKER_RETRY_MAX", 3),
			RetryDelay:      num("WORKER_RETRY_DELAY_MS", 1000),
			FetcherPoolSize: num("FETCHER_POOL_SIZE", 20),
			FetchTimeoutSec: num("FETCHER_TIMEOUT_SEC", 30),
		},
	}
}

func Path() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return ".env"
}
