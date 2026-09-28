package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server  ServerConfig
	Mysql   MysqlConfig
	Shopify ShopifyConfig
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

type ShopifyConfig struct {
	// ShopDomain is e.g. "keikibook.myshopify.com"
	ShopDomain  string
	AccessToken string
	APISecret   string
	APIVersion  string
	Timeout     string
	BaseURL     string
}

func (c ShopifyConfig) GraphQLEndpoint() string {
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = "https://" + strings.TrimRight(c.ShopDomain, "/")
	}
	return fmt.Sprintf("%s/admin/api/%s/graphql.json", base, c.APIVersion)
}

func (c ShopifyConfig) TimeoutDuration() time.Duration {
	if d, err := time.ParseDuration(c.Timeout); err == nil {
		return d
	}
	return 10 * time.Second
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8092"),
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "shopify_service"),
		},
		Shopify: ShopifyConfig{
			ShopDomain:  getEnv("SHOPIFY_SHOP_DOMAIN", ""),
			AccessToken: getEnv("SHOPIFY_ACCESS_TOKEN", ""),
			APISecret:   getEnv("SHOPIFY_API_SECRET", ""),
			APIVersion:  getEnv("SHOPIFY_API_VERSION", "2026-07"),
			Timeout:     getEnv("SHOPIFY_TIMEOUT", "10s"),
			BaseURL:     getEnv("SHOPIFY_BASE_URL", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
