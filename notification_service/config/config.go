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

