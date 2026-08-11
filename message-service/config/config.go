package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig `json:"server"`
	Cache  CacheConfig  `json:"cache"`
	Mysql  MysqlConfig  `json:"mysql"`
	Secret SecretConfig `json:"secret"`
	Kafka  KafkaConfig  `json:"kafka"`
}

type ServerConfig struct {
	ServerPort string `json:"server_port"`
	GRPCServer string `json:"grpc_server"`
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

type Kafka struct {
	Server string
}

type CacheConfig struct {
	Host string `json:"host"`
}

type KafkaConfig struct {
	KafkaURL string `json:"kafka_url"`
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
	AuthorizeKey string `json:"authorize_key"`
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
