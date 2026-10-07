package config

import (
	"os"
	"testing"
	"time"
)

func TestConsulJSONMatchesConfig(t *testing.T) {
	raw, err := os.ReadFile("../consul.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Defaults()
	if err := cfg.Merge(raw); err != nil {
		t.Fatal(err)
	}
	cfg.Postgres.User = "payment_auth_svc"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.CardService.BaseURL != "http://sard-card-service-svc:8080" || cfg.Challenge.OTPTTL != 5*time.Minute ||
		cfg.Challenge.ValidityTime != 10*time.Minute || cfg.Challenge.MaxAttempts != 3 || cfg.Challenge.MaxResends != 3 ||
		cfg.OTPDelivery.Backend != "log" || cfg.AuthorizeService.GRPCAddress != "authorize-service-grpc-svc:90" {
		t.Fatalf("config %+v", cfg)
	}
}

func TestEnvAndValidate(t *testing.T) {
	t.Setenv("CONSUL_ADDR", "")
	t.Setenv("AUTHORIZE_SERVICE_GRPC_ADDR", "authorize:9090")
	t.Setenv("USER_SERVICE_TOKEN", "ut")
	t.Setenv("INTERNAL_TOKENS", "a,b")
	t.Setenv("EXPOSE_OTP", "true")
	t.Setenv("OTP_TTL", "2m")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Server.InternalTokens) != 2 || !cfg.Challenge.ExposeOTP || cfg.Challenge.OTPTTL != 2*time.Minute ||
		cfg.AuthorizeService.GRPCAddress != "authorize:9090" || cfg.UserService.Token != "ut" {
		t.Fatalf("%+v", cfg)
	}
	for name, mut := range map[string]func(*Config){
		"no authorize":   func(c *Config) { c.AuthorizeService.GRPCAddress = "" },
		"webhook no url": func(c *Config) { c.OTPDelivery.Backend = "webhook" },
		"bad delivery":   func(c *Config) { c.OTPDelivery.Backend = "sms" },
		"relative url":   func(c *Config) { c.CardService.BaseURL = "cards" },
		"zero attempts":  func(c *Config) { c.Challenge.MaxAttempts = 0 },
	} {
		c := Defaults()
		mut(c)
		if c.Validate() == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
