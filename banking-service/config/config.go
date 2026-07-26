package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	Mysql  MysqlConfig
	Cache  CacheConfig
	Kafka  KafkaConfig
}

type ServerConfig struct {
	ServerPort string
	GRPCServer string
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
	Host     string
	Password string
}

type KafkaConfig struct {
	Brokers          []string
	TransactionTopic string
	ConsumerGroup    string
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

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			ServerPort: getEnv("PORT", "3000"),
			GRPCServer: getEnv("GRPC_SERVER", ""),
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "banking_service"),
			MysqlSSLMode:  getEnv("MYSQL_SSL_MODE", "false") == "true",
			MysqlDriver:   getEnv("MYSQL_DRIVER", "mysql"),
		},
		Cache: CacheConfig{
			Host:     getEnv("CACHE_HOST", "localhost:6379"),
			Password: getEnv("CACHE_PASSWORD", ""),
		},
		Kafka: KafkaConfig{
			Brokers:          strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
			TransactionTopic: getEnv("KAFKA_TRANSACTION_TOPIC", "banking.transactions"),
			ConsumerGroup:    getEnv("KAFKA_CONSUMER_GROUP", "banking-service"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}