package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database Database
	AI       AI
	Server   Server
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type AI struct {
	ClaudeAPIKey   string
	DeepSeekAPIKey string
}

type Server struct {
	Port string
}

func Load() (*Config, error) {
	config := &Config{
		Database: Database{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "chatbot_db"),
		},
		AI: AI{
			ClaudeAPIKey:   os.Getenv("CLAUDE_API_KEY"),
			DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"),
		},
		Server: Server{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}

	return config, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
