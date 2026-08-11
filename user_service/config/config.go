package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/JIeeiroSst/utils/consul"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Mysql    MysqlConfig    `json:"mysql"`
	Secret   SecretConfig   `json:"secret"`
	RabbitMQ RabbitMQ       `json:"rabbit_mq"`
	Redis    Redis          `json:"redis"`
	Email    Email          `json:"email"`
	Postgres PostgresConfig `json:"postgres"`
	Token    TokenPolicy    `json:"token"`
}

type TokenPolicy struct {
	AccessTokenExpMinutes int `json:"access_token_exp_minutes"`
	RefreshTokenExpHours  int `json:"refresh_token_exp_hours"`
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
	PortHttpServer string `json:"port_http_server"`
	PortGrpcServer string `json:"port_grpc_server"`
}

type Redis struct {
	Dns string `json:"dns"`
}

type RabbitMQ struct {
	Host           string `json:"host"`
	Port           string `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	Exchange       string `json:"exchange"`
	Queue          string `json:"queue"`
	RoutingKey     string `json:"routing_key"`
	ConsumerTag    string `json:"consumer_tag"`
	WorkerPoolSize int    `json:"worker_pool_size"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresql_host"`
	PostgresqlPort     string `json:"postgresql_port"`
	PostgresqlUser     string `json:"postgresql_user"`
	PostgresqlPassword string `json:"postgresql_password"`
	PostgresqlDbname   string `json:"postgresql_dbname"`
	PostgresqlSSLMode  bool   `json:"postgresql_ssl_mode"`
	PgDriver           string `json:"pg_driver"`
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

type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
}

type ElasticsearchConfig struct {
	Dns string `json:"dns"`
}

type Email struct {
	NameEmail     string `json:"name_email"`
	PasswordEmail string `json:"password_email"`
	Port          string `json:"port"`
	Host          string `json:"host"`
}

type Consul struct {
	LockIndex int    `json:"lock_index"`
	Key       int    `json:"key"`
	Flags     int    `json:"flags"`
	Value     string `json:"value"`
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
