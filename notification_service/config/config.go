package config

import (
	"os"
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
	APIKey string
	From   string
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

