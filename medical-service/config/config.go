package config

import "os"

type Config struct {
	Server   ServerConfig
	Mysql    MysqlConfig
	Pharmacy PharmacyConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
}

type PharmacyConfig struct {
	TimeZone string
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8093"),
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "medical_service"),
		},
		Pharmacy: PharmacyConfig{
			TimeZone: getEnv("PHARMACY_TIMEZONE", "Asia/Ho_Chi_Minh"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
