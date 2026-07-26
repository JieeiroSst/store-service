package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Mysql    MysqlConfig
	Secret   SecretConfig
	RabbitMQ RabbitMQ
	Redis    Redis
	Email    Email
	Postgres PostgresConfig
	Token    TokenPolicy
}

type TokenPolicy struct {
	AccessTokenExpMinutes int
	RefreshTokenExpHours  int
}

func (t TokenPolicy) AccessTokenTTL() time.Duration {
	if t.AccessTokenExpMinutes <= 0 {
		return 15 * time.Minute
	}
	return time.Duration(t.AccessTokenExpMinutes) * time.Minute
}

func (t TokenPolicy) RefreshTokenTTL() time.Duration {
	if t.RefreshTokenExpHours <= 0 {
		return 24 * 7 * time.Hour
	}
	return time.Duration(t.RefreshTokenExpHours) * time.Hour
}

type ServerConfig struct {
	PortHttpServer string
	PortGrpcServer string
}

type Redis struct {
	Dns string
}

type RabbitMQ struct {
	Host           string
	Port           string
	User           string
	Password       string
	Exchange       string
	Queue          string
	RoutingKey     string
	ConsumerTag    string
	WorkerPoolSize int
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
	PgDriver           string
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
	MysqlSSLMode  bool
	MysqlDriver   string
}

type SecretConfig struct {
	JwtSecretKey string
}

type ElasticsearchConfig struct {
	Dns string
}

type Email struct {
	NameEmail     string
	PasswordEmail string
	Port          string
	Host          string
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

func InitializeConfiguration(dir string) (*Config, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	consul := consul.NewConfigConsul(os.Getenv("HostConsul"), os.Getenv("KeyConsul"), os.Getenv("ServiceConsul"))
	conf, err := consul.ConnectConfigConsul()
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(conf, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
