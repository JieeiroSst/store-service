package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server struct {
		Port           string
		AllowedOrigins []string
	}
	MySQL struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
	}
	Redis struct {
		Addr     string
		Password string
	}
	UserService struct {
		GRPCAddr string
		AuthCacheTTL time.Duration
		Timeout      time.Duration
	}
	Chat struct {
		MaxMessageLen int
		HistoryLimit  int
	}
}

func (c *Config) MySQLDSN() string {
	m := c.MySQL
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC", m.User, m.Password, m.Host, m.Port, m.Name)
}

func Load() *Config {
	c := &Config{}
	c.Server.Port = env("PORT_HTTP_SERVER", "8081")
	c.Server.AllowedOrigins = splitCSV(env("ALLOWED_ORIGINS", "*"))

	c.MySQL.Host = env("MYSQL_HOST", "localhost")
	c.MySQL.Port = env("MYSQL_PORT", "3306")
	c.MySQL.User = env("MYSQL_USER", "chatuser")
	c.MySQL.Password = env("MYSQL_PASSWORD", "chatpass123")
	c.MySQL.Name = env("MYSQL_DATABASE", "chatdb")

	c.Redis.Addr = os.Getenv("REDIS_ADDR")
	c.Redis.Password = os.Getenv("REDIS_PASSWORD")

	c.UserService.GRPCAddr = env("USER_SERVICE_GRPC_ADDR", "localhost:1236")
	c.UserService.AuthCacheTTL = time.Duration(envInt("AUTH_CACHE_TTL_SECONDS", 30)) * time.Second
	c.UserService.Timeout = time.Duration(envInt("USER_SERVICE_TIMEOUT_SECONDS", 5)) * time.Second

	c.Chat.MaxMessageLen = int(envInt("MAX_MESSAGE_LEN", 4000))
	c.Chat.HistoryLimit = int(envInt("HISTORY_LIMIT", 50))
	return c
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int64) int64 {
	if v, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil && v > 0 {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
