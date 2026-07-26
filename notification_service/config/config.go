package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Mysql    MysqlConfig
	Rabbit   RabbitConfig
	Firebase FirebaseConfig
	Email    EmailConfig
	Slack    SlackConfig
}

type ServerConfig struct {
	ServerPort string
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
}

type RabbitConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	VirtualHost string
	MaxRetries  int
	RetryDelay  time.Duration
}

type FirebaseConfig struct {
	CredentialsFile string
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SlackConfig struct {
	WebhookSecret string
	Channel       string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func ReadFileEnv(dir string) (*Dir, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	data := &Dir{
		HostConsul:    os.Getenv("HostConsul"),
		KeyConsul:     os.Getenv("KeyConsul"),
		ServiceConsul: os.Getenv("ServiceConsul"),
	}
	return data, nil
}

// FromEnv builds configuration straight from environment variables. Used
// when Consul is not configured (local dev, CI, Consul-less deployments).
func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			ServerPort: getEnv("PORT", "1235"),
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "notification_service"),
		},
		Rabbit: RabbitConfig{
			Host:        getEnv("RABBIT_HOST", "localhost"),
			Port:        getEnvInt("RABBIT_PORT", 5672),
			Username:    getEnv("RABBIT_USERNAME", "guest"),
			Password:    getEnv("RABBIT_PASSWORD", "guest"),
			VirtualHost: getEnv("RABBIT_VHOST", "/"),
			MaxRetries:  getEnvInt("RABBIT_MAX_RETRIES", 5),
			RetryDelay:  time.Duration(getEnvInt("RABBIT_RETRY_DELAY_SECONDS", 5)) * time.Second,
		},
		Firebase: FirebaseConfig{
			CredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
		},
		Email: EmailConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnvInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		Slack: SlackConfig{
			WebhookSecret: getEnv("SLACK_WEBHOOK_SECRET", ""),
			Channel:       getEnv("SLACK_CHANNEL", "#notifications"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
