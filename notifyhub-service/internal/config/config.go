package config

import (
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
	Port         string `mapstructure:"PORT"`
	Mode         string `mapstructure:"MODE"`
	JWTSecret    string `mapstructure:"JWT_SECRET"`
	RateLimit    int    `mapstructure:"RATE_LIMIT"`
}

type DatabaseConfig struct {
	DSN             string `mapstructure:"DB_DSN"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime int    `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

type EmailConfig struct {
	Provider  string `mapstructure:"EMAIL_PROVIDER"` // smtp | sendgrid
	SMTPHost  string `mapstructure:"SMTP_HOST"`
	SMTPPort  int    `mapstructure:"SMTP_PORT"`
	SMTPUser  string `mapstructure:"SMTP_USER"`
	SMTPPass  string `mapstructure:"SMTP_PASS"`
	FromAddr  string `mapstructure:"EMAIL_FROM"`
	SendGridKey string `mapstructure:"SENDGRID_API_KEY"`
}

type SMSConfig struct {
	Provider      string `mapstructure:"SMS_PROVIDER"` // twilio | vonage
	TwilioSID     string `mapstructure:"TWILIO_SID"`
	TwilioToken   string `mapstructure:"TWILIO_TOKEN"`
	TwilioFrom    string `mapstructure:"TWILIO_FROM"`
	VonageKey     string `mapstructure:"VONAGE_API_KEY"`
	VonageSecret  string `mapstructure:"VONAGE_API_SECRET"`
}

type FirebaseConfig struct {
	CredentialsFile string `mapstructure:"FIREBASE_CREDENTIALS_FILE"`
	ProjectID       string `mapstructure:"FIREBASE_PROJECT_ID"`
}

type WorkerConfig struct {
	PoolSize        int `mapstructure:"WORKER_POOL_SIZE"`
	QueueSize       int `mapstructure:"WORKER_QUEUE_SIZE"`
	RetryMax        int `mapstructure:"WORKER_RETRY_MAX"`
	RetryDelay      int `mapstructure:"WORKER_RETRY_DELAY_MS"`
	FetcherPoolSize int `mapstructure:"FETCHER_POOL_SIZE"`
	FetchTimeoutSec int `mapstructure:"FETCHER_TIMEOUT_SEC"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// fallback to env only
		viper.SetConfigType("env")
	}

	cfg := &Config{}
	cfg.Server.Port = viper.GetString("PORT")
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	cfg.Server.Mode = viper.GetString("MODE")
	cfg.Server.JWTSecret = viper.GetString("JWT_SECRET")
	cfg.Server.RateLimit = viper.GetInt("RATE_LIMIT")
	if cfg.Server.RateLimit == 0 {
		cfg.Server.RateLimit = 100
	}

	cfg.Database.DSN = viper.GetString("DB_DSN")
	cfg.Database.MaxOpenConns = viper.GetInt("DB_MAX_OPEN_CONNS")
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 50
	}
	cfg.Database.MaxIdleConns = viper.GetInt("DB_MAX_IDLE_CONNS")
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	cfg.Database.ConnMaxLifetime = viper.GetInt("DB_CONN_MAX_LIFETIME")
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 300
	}

	cfg.Email.Provider = viper.GetString("EMAIL_PROVIDER")
	cfg.Email.SMTPHost = viper.GetString("SMTP_HOST")
	cfg.Email.SMTPPort = viper.GetInt("SMTP_PORT")
	cfg.Email.SMTPUser = viper.GetString("SMTP_USER")
	cfg.Email.SMTPPass = viper.GetString("SMTP_PASS")
	cfg.Email.FromAddr = viper.GetString("EMAIL_FROM")
	cfg.Email.SendGridKey = viper.GetString("SENDGRID_API_KEY")

	cfg.SMS.Provider = viper.GetString("SMS_PROVIDER")
	cfg.SMS.TwilioSID = viper.GetString("TWILIO_SID")
	cfg.SMS.TwilioToken = viper.GetString("TWILIO_TOKEN")
	cfg.SMS.TwilioFrom = viper.GetString("TWILIO_FROM")
	cfg.SMS.VonageKey = viper.GetString("VONAGE_API_KEY")
	cfg.SMS.VonageSecret = viper.GetString("VONAGE_API_SECRET")

	cfg.Firebase.CredentialsFile = viper.GetString("FIREBASE_CREDENTIALS_FILE")
	cfg.Firebase.ProjectID = viper.GetString("FIREBASE_PROJECT_ID")

	cfg.Worker.PoolSize = viper.GetInt("WORKER_POOL_SIZE")
	if cfg.Worker.PoolSize == 0 {
		cfg.Worker.PoolSize = 10
	}
	cfg.Worker.QueueSize = viper.GetInt("WORKER_QUEUE_SIZE")
	if cfg.Worker.QueueSize == 0 {
		cfg.Worker.QueueSize = 1000
	}
	cfg.Worker.RetryMax = viper.GetInt("WORKER_RETRY_MAX")
	if cfg.Worker.RetryMax == 0 {
		cfg.Worker.RetryMax = 3
	}
	cfg.Worker.RetryDelay = viper.GetInt("WORKER_RETRY_DELAY_MS")
	if cfg.Worker.RetryDelay == 0 {
		cfg.Worker.RetryDelay = 1000
	}
	cfg.Worker.FetcherPoolSize = viper.GetInt("FETCHER_POOL_SIZE")
	if cfg.Worker.FetcherPoolSize == 0 {
		cfg.Worker.FetcherPoolSize = 20
	}
	cfg.Worker.FetchTimeoutSec = viper.GetInt("FETCHER_TIMEOUT_SEC")
	if cfg.Worker.FetchTimeoutSec == 0 {
		cfg.Worker.FetchTimeoutSec = 30
	}

	return cfg, nil
}
