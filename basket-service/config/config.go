package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig `json:"server"`
	Mysql  MysqlConfig  `json:"mysql"`
	Secret SecretConfig `json:"secret"`
	Cache  CacheConfig  `json:"cache"`
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

type SecretConfig struct {
	JwtSecretKey string `json:"jwt_secret_key"`
	AuthorizeKey string `json:"authorize_key"`
}

type CacheConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
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

