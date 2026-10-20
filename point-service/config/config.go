package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Mysql    MysqlConfig    `json:"mysql"`
	Cache    CacheConfig    `json:"cache"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`
}

type RabbitMQConfig struct {
	Enabled  bool     `json:"enabled"`
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	VHost    string   `json:"vhost"`
	Queue    string   `json:"queue"`
	Bindings []string `json:"bindings"`
	Prefetch int      `json:"prefetch"`
}

func (c *RabbitMQConfig) ApplyEnv() {
	if v := os.Getenv("RABBITMQ_ENABLED"); v != "" {
		c.Enabled = v == "true" || v == "1"
	}
	for env, dst := range map[string]*string{
		"RABBITMQ_HOST":     &c.Host,
		"RABBITMQ_PORT":     &c.Port,
		"RABBITMQ_USER":     &c.User,
		"RABBITMQ_PASSWORD": &c.Password,
		"RABBITMQ_VHOST":    &c.VHost,
		"RABBITMQ_QUEUE":    &c.Queue,
	} {
		if v := os.Getenv(env); v != "" {
			*dst = v
		}
	}
	if c.Port == "" {
		c.Port = "5672"
	}
	if c.Queue == "" {
		c.Queue = "point-service.rewards"
	}
	if len(c.Bindings) == 0 {
		c.Bindings = []string{"lucky:lucky.reward.granted"}
	}
	if c.Prefetch <= 0 {
		c.Prefetch = 20
	}
}

type ServerConfig struct {
	ServerPort string `json:"server_port"`
	GRPCServer string `json:"grpc_server"`
}

type MysqlConfig struct {
	MysqlHost     string `json:"mysql_host"`
	MysqlPort     string `json:"mysql_port"`
	MysqlUser     string `json:"mysql_user"`
	MysqlPassword string `json:"mysql_password"`
	MysqlDbname   string `json:"mysql_dbname"`
	MysqlSSLMode  bool   `json:"mysql_ssl_mode"`
	MysqlDriver   string `json:"mysql_driver"`
}

type Consul struct {
	LockIndex int
	Key       int
	Flags     int
	Value     string
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

type CacheConfig struct {
	Host string `json:"host"`
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
