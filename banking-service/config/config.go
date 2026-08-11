package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig `json:"server"`
	Mysql  MysqlConfig  `json:"mysql"`
	Cache  CacheConfig  `json:"cache"`
	Kafka  KafkaConfig  `json:"kafka"`
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
	Host     string `json:"host"`
	Password string `json:"password"`
}

type KafkaConfig struct {
	Brokers          []string `json:"brokers"`
	TransactionTopic string   `json:"transaction_topic"`
	ConsumerGroup    string   `json:"consumer_group"`
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

