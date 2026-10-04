package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

var Version = "1.0.0"

type Config struct {
	Server     ServerConfig
	Agent      AgentConfig
	ACL        ACLConfig
	Health     HealthConfig
	Blocking   BlockingConfig
	KV         KVConfig
	Intentions IntentionsConfig
	DNS        DNSConfig
	Snapshot   SnapshotConfig
	Store      StoreConfig
}

type StoreConfig struct {
	Backend           string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisPrefix       string
	RedisPollInterval time.Duration
	RedisOpTimeout    time.Duration
	LeaderTTL         time.Duration
}

type ServerConfig struct {
	PortHttpServer string
}

type AgentConfig struct {
	NodeName    string
	NodeAddress string
	Datacenter  string
}

type ACLConfig struct {
	DefaultPolicy   string
	ManagementToken string
	ReadTokens      []string
}

type HealthConfig struct {
	TickInterval   time.Duration
	MaxConcurrent  int
	DefaultTimeout time.Duration
}

type BlockingConfig struct {
	DefaultWait time.Duration
	MaxWait     time.Duration
}

type KVConfig struct {
	MaxValueBytes int
}

type IntentionsConfig struct {
	DefaultAllow bool
}

type DNSConfig struct {
	Port        string
	Domain      string
	TTL         time.Duration
	OnlyPassing bool
}

type SnapshotConfig struct {
	DataDir  string
	Interval time.Duration
}

func FromEnv() *Config {
	host, _ := os.Hostname()
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8500"),
		},
		Agent: AgentConfig{
			NodeName:    getEnv("NODE_NAME", host),
			NodeAddress: getEnv("NODE_ADDRESS", "127.0.0.1"),
			Datacenter:  getEnv("DATACENTER", "dc1"),
		},
		ACL: ACLConfig{
			DefaultPolicy:   getEnv("ACL_DEFAULT_POLICY", "allow"),
			ManagementToken: os.Getenv("ACL_MANAGEMENT_TOKEN"),
			ReadTokens:      getList("ACL_READ_TOKENS"),
		},
		Health: HealthConfig{
			TickInterval:   getDuration("CHECK_TICK_INTERVAL", time.Second),
			MaxConcurrent:  getInt("CHECK_MAX_CONCURRENT", 64),
			DefaultTimeout: getDuration("CHECK_DEFAULT_TIMEOUT", 10*time.Second),
		},
		Blocking: BlockingConfig{
			DefaultWait: getDuration("BLOCKING_DEFAULT_WAIT", 5*time.Minute),
			MaxWait:     getDuration("BLOCKING_MAX_WAIT", 10*time.Minute),
		},
		KV: KVConfig{
			MaxValueBytes: getInt("KV_MAX_VALUE_KB", 512) << 10,
		},
		Intentions: IntentionsConfig{
			DefaultAllow: getEnv("INTENTIONS_DEFAULT", "allow") != "deny",
		},
		DNS: DNSConfig{
			Port:        getEnv("PORT_DNS_SERVER", "8600"),
			Domain:      strings.Trim(getEnv("DNS_DOMAIN", "consul"), "."),
			TTL:         getDuration("DNS_TTL", 0),
			OnlyPassing: getEnv("DNS_ONLY_PASSING", "false") == "true",
		},
		Snapshot: SnapshotConfig{
			DataDir:  os.Getenv("DATA_DIR"),
			Interval: getDuration("SNAPSHOT_INTERVAL", 30*time.Second),
		},
		Store: StoreConfig{
			Backend:           getEnv("STORE_BACKEND", "memory"),
			RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
			RedisPassword:     os.Getenv("REDIS_PASSWORD"),
			RedisDB:           getNonNegInt("REDIS_DB", 0),
			RedisPrefix:       getEnv("REDIS_PREFIX", "networking:"),
			RedisPollInterval: getDuration("REDIS_POLL_INTERVAL", time.Second),
			RedisOpTimeout:    getDuration("REDIS_OP_TIMEOUT", 5*time.Second),
			LeaderTTL:         getDuration("LEADER_TTL", 10*time.Second),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return fallback
}

func getNonNegInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v >= 0 {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil && v >= 0 {
		return v
	}
	return fallback
}

func getList(key string) []string {
	var out []string
	for _, v := range strings.Split(os.Getenv(key), ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
