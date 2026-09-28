package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Mysql    MysqlConfig    `json:"mysql"`
	Rabbit   RabbitConfig   `json:"rabbit"`
	Firebase FirebaseConfig `json:"firebase"`
	Email    EmailConfig    `json:"email"`
	Slack    SlackConfig    `json:"slack"`
	Worker   WorkerConfig   `json:"worker"`
	Campaign CampaignConfig `json:"campaign"`
	Push     PushConfig     `json:"push"`
	Audit    AuditConfig    `json:"audit"`
}

type AuditConfig struct {
	RetentionDays int `json:"retention_days"`
}

type PushConfig struct {
	SingleDevicePerUser bool  `json:"single_device_per_user"`
	StaleTokenDays      int   `json:"stale_token_days"`
	ValidateOnRegister  *bool `json:"validate_on_register"`
}

type WorkerConfig struct {
	Concurrency int `json:"concurrency"`
	Prefetch    int `json:"prefetch"`
}

type CampaignConfig struct {
	PushBatchSize  int     `json:"push_batch_size"`
	EmailBatchSize int     `json:"email_batch_size"`
	MaxRecipients  int     `json:"max_recipients"`
	MaxAttempts    int     `json:"max_attempts"`
	RatePerSecond  float64 `json:"rate_per_second"`
	Concurrency    int     `json:"concurrency"`
	Prefetch       int     `json:"prefetch"`
}

type ServerConfig struct {
	ServerPort string `json:"server_port"`
}

type MysqlConfig struct {
	MysqlHost     string `json:"mysql_host"`
	MysqlPort     string `json:"mysql_port"`
	MysqlUser     string `json:"mysql_user"`
	MysqlPassword string `json:"mysql_password"`
	MysqlDbname   string `json:"mysql_dbname"`
}

type RabbitConfig struct {
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	Username    string        `json:"username"`
	Password    string        `json:"password"`
	VirtualHost string        `json:"virtual_host"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
}

type FirebaseConfig struct {
	CredentialsFile string `json:"credentials_file"`
}

type EmailConfig struct {
	APIKey string `json:"api_key"`
	From   string `json:"from"`
}

type SlackConfig struct {
	WebhookSecret string `json:"webhook_secret"`
	Channel       string `json:"channel"`
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
