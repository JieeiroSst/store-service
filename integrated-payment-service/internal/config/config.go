package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	MoMo     MoMoConfig     `mapstructure:"momo"`
	VNPay    VNPayConfig    `mapstructure:"vnpay"`
	ZaloPay  ZaloPayConfig  `mapstructure:"zalopay"`
	PayPal   PayPalConfig   `mapstructure:"paypal"`
	Stripe   StripeConfig   `mapstructure:"stripe"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	LogLevel string         `mapstructure:"log_level"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type MoMoConfig struct {
	PartnerCode string `mapstructure:"partner_code"`
	AccessKey   string `mapstructure:"access_key"`
	SecretKey   string `mapstructure:"secret_key"`
	Endpoint    string `mapstructure:"endpoint"`
	ReturnURL   string `mapstructure:"return_url"`
	NotifyURL   string `mapstructure:"notify_url"`
}

type VNPayConfig struct {
	TmnCode    string `mapstructure:"tmn_code"`
	HashSecret string `mapstructure:"hash_secret"`
	Endpoint   string `mapstructure:"endpoint"`
	ReturnURL  string `mapstructure:"return_url"`
}

type ZaloPayConfig struct {
	AppID     string `mapstructure:"app_id"`
	Key1      string `mapstructure:"key1"`
	Key2      string `mapstructure:"key2"`
	Endpoint  string `mapstructure:"endpoint"`
	ReturnURL string `mapstructure:"return_url"`
}

type PayPalConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Endpoint     string `mapstructure:"endpoint"`
	ReturnURL    string `mapstructure:"return_url"`
}

type StripeConfig struct {
	SecretKey     string `mapstructure:"secret_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	ReturnURL     string `mapstructure:"return_url"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("log_level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
