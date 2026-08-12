package config

import (
	"fmt"
	"net/url"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Postgres PostgresConfig `json:"postgres"`
	JWT      JWTConfig      `json:"jwt"`
}

type ServerConfig struct {
	Port string `json:"port"`
	Env  string `json:"env"` // "development" | "production" — drives logger mode
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
	SSLMode  string `json:"sslMode"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode)
}

func (p PostgresConfig) URL() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     fmt.Sprintf("%s:%s", p.Host, p.Port),
		Path:     "/" + p.DBName,
		RawQuery: "sslmode=" + p.SSLMode,
	}
	return u.String()
}

type JWTConfig struct {
	Secret        string `json:"secret"`
	ExpiryMinutes int    `json:"expiryMinutes"`
}
