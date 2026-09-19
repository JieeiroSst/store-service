package config

import "os"

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Auth   AuthConfig
	Minio  MinioConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

type AuthConfig struct {
	JWTSecret string
}

type MinioConfig struct {
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "1236"),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
			Dbname:   getEnv("MYSQL_DBNAME", "post"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", ""),
		},
		Minio: MinioConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:       getEnv("MINIO_ACCESS_KEY", ""),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", ""),
			BucketName:      getEnv("MINIO_BUCKET", "post-service"),
			UseSSL:          getEnv("MINIO_USE_SSL", "true") == "true",
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
