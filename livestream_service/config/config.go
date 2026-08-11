package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Storage   StorageConfig
	Node      NodeConfig
	Transcode TranscodeConfig
	Viewer    ViewerConfig
}

type ServerConfig struct {
	PortHttpServer string
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type StorageConfig struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type NodeConfig struct {
	ID         string
	RTMPAddr   string // e.g. rtmp://livestream-service-0.livestream-service-headless:1935/live
	LocalRTMP  string // e.g. rtmp://127.0.0.1:1935/live (SRS sidecar, same pod)
	MaxStreams int
}

type TranscodeConfig struct {
	HLSDir                string
	FFmpegPath            string
	SegmentTime           int
	SegmentListSize       int
	MaxRestartsPerWindow  int
	RestartWindow         string
	RestartDelay          string
	WatchdogStaleAfter    string
	WatchdogCheckInterval string
	HeartbeatInterval     string
	HeartbeatTTL          string
}

func (t TranscodeConfig) RestartWindowDuration() time.Duration {
	return parseDurationOr(t.RestartWindow, 60*time.Second)
}

func (t TranscodeConfig) RestartDelayDuration() time.Duration {
	return parseDurationOr(t.RestartDelay, 2*time.Second)
}

func (t TranscodeConfig) WatchdogStaleAfterDuration() time.Duration {
	return parseDurationOr(t.WatchdogStaleAfter, 30*time.Second)
}

func (t TranscodeConfig) WatchdogCheckIntervalDuration() time.Duration {
	return parseDurationOr(t.WatchdogCheckInterval, 30*time.Second)
}

func (t TranscodeConfig) HeartbeatIntervalDuration() time.Duration {
	return parseDurationOr(t.HeartbeatInterval, 5*time.Second)
}

func (t TranscodeConfig) HeartbeatTTLDuration() time.Duration {
	return parseDurationOr(t.HeartbeatTTL, 15*time.Second)
}

type ViewerConfig struct {
	HeartbeatWindow string
}

func (v ViewerConfig) HeartbeatWindowDuration() time.Duration {
	return parseDurationOr(v.HeartbeatWindow, 40*time.Second)
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8080"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", ""),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "livestream"),
			PostgresqlSSLMode:  getEnv("POSTGRES_SSLMODE", "disable") == "require",
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Storage: StorageConfig{
			Endpoint:  getEnv("S3_ENDPOINT", "localhost:9000"),
			Bucket:    getEnv("S3_BUCKET", "livestream-hls"),
			AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""),
			UseSSL:    getEnv("S3_USE_SSL", "false") == "true",
		},
		Node: NodeConfig{
			ID:         getEnv("NODE_ID", hostnameFallback()),
			RTMPAddr:   getEnv("NODE_RTMP_ADDR", ""),
			LocalRTMP:  getEnv("NODE_LOCAL_RTMP", "rtmp://127.0.0.1:1935/live"),
			MaxStreams: getEnvInt("MAX_STREAMS", 20),
		},
		Transcode: TranscodeConfig{
			HLSDir:                getEnv("HLS_DIR", "/var/hls"),
			FFmpegPath:            getEnv("FFMPEG_PATH", "ffmpeg"),
			SegmentTime:           getEnvInt("HLS_SEGMENT_TIME", 4),
			SegmentListSize:       getEnvInt("HLS_LIST_SIZE", 6),
			MaxRestartsPerWindow:  getEnvInt("TRANSCODE_MAX_RESTARTS", 5),
			RestartWindow:         getEnv("TRANSCODE_RESTART_WINDOW", "60s"),
			RestartDelay:          getEnv("TRANSCODE_RESTART_DELAY", "2s"),
			WatchdogStaleAfter:    getEnv("WATCHDOG_STALE_AFTER", "30s"),
			WatchdogCheckInterval: getEnv("WATCHDOG_CHECK_INTERVAL", "30s"),
			HeartbeatInterval:     getEnv("HEARTBEAT_INTERVAL", "5s"),
			HeartbeatTTL:          getEnv("HEARTBEAT_TTL", "15s"),
		},
		Viewer: ViewerConfig{
			HeartbeatWindow: getEnv("VIEWER_HEARTBEAT_WINDOW", "40s"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func parseDurationOr(s string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return fallback
}

func hostnameFallback() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "node-local"
	}
	return h
}
