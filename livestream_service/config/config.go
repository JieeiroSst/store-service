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
	Auth      AuthConfig
	Internal  InternalConfig
	Playback  PlaybackConfig
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
	HTTPAddr   string // e.g. http://livestream-service-0.livestream-service-headless:8080 - lets the edge role reach this specific node for admin actions (force-unpublish)
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

// AuthConfig signs/verifies the JWTs this service accepts on protected
// routes. This service is a resource server, not an identity provider -
// tokens are issued elsewhere (e.g. user_service) and just need to share
// this secret.
type AuthConfig struct {
	JWTSecret string
}

// InternalConfig guards service-to-service routes (edge calling a specific
// node's HTTP port) that must never be reachable from outside the cluster.
type InternalConfig struct {
	SharedSecret string
}

// PlaybackConfig signs time-limited playback URLs (HMAC "token
// authentication", the scheme BunnyCDN/Cloudflare's built-in token-auth
// features expect) so raw HLS links can't be hotlinked/shared indefinitely.
// Wiring the CDN itself to validate these tokens is outside this service.
type PlaybackConfig struct {
	SigningSecret string
	CDNBaseURL    string
	TokenTTL      string
}

func (p PlaybackConfig) TokenTTLDuration() time.Duration {
	return parseDurationOr(p.TokenTTL, 6*time.Hour)
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
			HTTPAddr:   getEnv("NODE_HTTP_ADDR", ""),
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
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", ""),
		},
		Internal: InternalConfig{
			SharedSecret: getEnv("INTERNAL_SHARED_SECRET", ""),
		},
		Playback: PlaybackConfig{
			SigningSecret: getEnv("PLAYBACK_SIGNING_SECRET", ""),
			CDNBaseURL:    getEnv("PLAYBACK_CDN_BASE_URL", ""),
			TokenTTL:      getEnv("PLAYBACK_TOKEN_TTL", "6h"),
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
