package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server      ServerConfig
	Mongo       MongoConfig
	Storage     StorageConfig
	Upload      UploadConfig
	Auth        AuthConfig
	UserService UserServiceConfig
}

type ServerConfig struct {
	Port            string
	CORSOrigins     []string
	ShutdownTimeout time.Duration
}

type MongoConfig struct {
	URI      string
	Database string
}

type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type UploadConfig struct {
	MaxBytes     int64
	AllowedTypes []string
}

const (
	AuthToken = "token"
	AuthOff   = "off"
)

type AuthConfig struct {
	Mode     string
	CacheTTL time.Duration
}

type UserServiceConfig struct {
	BaseURL string
	Timeout time.Duration
}

func Load() *Config {
	_ = godotenv.Load(".env")

	return &Config{
		Server: ServerConfig{
			Port:            get("PORT", "8080"),
			CORSOrigins:     split(get("CORS_ALLOW_ORIGINS", "")),
			ShutdownTimeout: duration(get("SHUTDOWN_TIMEOUT", "15s"), 15*time.Second),
		},
		Mongo: MongoConfig{
			URI:      get("MONGO_URI", "mongodb://localhost:27017"),
			Database: get("MONGO_DATABASE", "upload"),
		},
		Storage: StorageConfig{
			Endpoint:  get("MINIO_ENDPOINT", ""),
			AccessKey: get("MINIO_ACCESS_KEY", ""),
			SecretKey: get("MINIO_SECRET_KEY", ""),
			Bucket:    get("MINIO_BUCKET", "upload-service"),
			UseSSL:    get("MINIO_USE_SSL", "") == "true",
		},
		Upload: UploadConfig{
			MaxBytes: int64(intOr(get("UPLOAD_MAX_MB", "20"), 20)) << 20,
			AllowedTypes: split(get("UPLOAD_ALLOWED_TYPES",
				"application/pdf,image/jpeg,image/png,image/webp,image/gif")),
		},
		Auth: AuthConfig{
			Mode:     strings.ToLower(get("AUTH_MODE", AuthToken)),
			CacheTTL: nonNegative(get("AUTH_CACHE_TTL", "5s"), 5*time.Second),
		},
		UserService: UserServiceConfig{
			BaseURL: strings.TrimRight(get("USER_SERVICE_BASE_URL", ""), "/"),
			Timeout: duration(get("USER_SERVICE_TIMEOUT", "5s"), 5*time.Second),
		},
	}
}

// Validate rejects configurations that would fail on the first request.
func (c *Config) Validate() error {
	switch c.Auth.Mode {
	case AuthToken:
		if c.UserService.BaseURL == "" {
			return errors.New("AUTH_MODE=token needs USER_SERVICE_BASE_URL (or set AUTH_MODE=off)")
		}
	case AuthOff:
	default:
		return errors.New("AUTH_MODE must be 'token' or 'off'")
	}
	if c.Storage.Endpoint == "" || c.Storage.AccessKey == "" || c.Storage.SecretKey == "" {
		return errors.New("MINIO_ENDPOINT, MINIO_ACCESS_KEY and MINIO_SECRET_KEY are required")
	}
	if len(c.Upload.AllowedTypes) == 0 {
		return errors.New("UPLOAD_ALLOWED_TYPES must list at least one type")
	}
	return nil
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func intOr(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return def
}

func duration(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}

func nonNegative(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d >= 0 {
		return d
	}
	if n, err := strconv.Atoi(s); err == nil && n >= 0 {
		return time.Duration(n) * time.Second
	}
	return def
}
