package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// consul.json is what ops seed into Consul: every key in it must land in Config.
func TestConsulJSONMapsOntoConfig(t *testing.T) {
	raw, err := os.ReadFile("../consul.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Server.PortHttpServer != "8091" || cfg.Wallet.BaseURL == "" || cfg.Notification.BaseURL == "" {
		t.Fatalf("server/wallet/notification not mapped: %+v", cfg)
	}
	if cfg.Exchange.Currency != "USD" || cfg.Exchange.ShareValue != 100 || cfg.Exchange.MinOrderSize != 1 {
		t.Fatalf("exchange not mapped: %+v", cfg.Exchange)
	}
	if cfg.UserService.BaseURL == "" || cfg.UserService.AuthMode != "gateway" || cfg.Referral.BaseURL == "" || cfg.Redis.Addr == "" {
		t.Fatalf("user-service/referral/redis not mapped: %+v %+v %+v", cfg.UserService, cfg.Referral, cfg.Redis)
	}
	if cfg.Fees.MakerRebateBps != 2000 || cfg.Fees.ReferralBps != 1000 || cfg.Exchange.RewardEpochDuration() != time.Minute {
		t.Fatalf("fees/rewards not mapped: %+v %+v", cfg.Fees, cfg.Exchange)
	}
	if cfg.Exchange.DisputeWindowDuration() != 2*time.Hour || cfg.Wallet.TimeoutDuration() != 10*time.Second {
		t.Fatal("durations not parsed")
	}
}

func TestEnvDefaults(t *testing.T) {
	t.Setenv("SHARE_VALUE", "1000")
	t.Setenv("DISPUTE_WINDOW", "not-a-duration")
	cfg := FromEnv()
	if cfg.Exchange.ShareValue != 1000 || cfg.Exchange.Currency != "USD" {
		t.Fatalf("%+v", cfg.Exchange)
	}
	if cfg.Exchange.DisputeWindowDuration() != 2*time.Hour {
		t.Fatal("a bad duration must fall back to the default")
	}
}
