package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server struct {
		Port           string
		AllowedOrigins []string
	}
	Minio struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
	}
	Redis struct {
		Addr     string
		Password string
	}
	Worker struct {
		Concurrency int
		JobTimeout  time.Duration
		ClaimIdle   time.Duration
	}
	Upload struct {
		MaxBytes int64
	}
	Stream struct {
		ChunkSize      int64
		CacheBytes     int64
		PrefetchChunks int
		MetaTTL        time.Duration
		ListTTL        time.Duration
	}
}

func Load() *Config {
	c := &Config{}
	c.Server.Port = env("PORT_HTTP_SERVER", "8080")
	c.Server.AllowedOrigins = splitCSV(env("ALLOWED_ORIGINS", "*"))

	c.Minio.Endpoint = env("MINIO_ENDPOINT", "localhost:9000")
	c.Minio.AccessKey = env("MINIO_ACCESS_KEY", "minioadmin")
	c.Minio.SecretKey = env("MINIO_SECRET_KEY", "minioadmin123")
	c.Minio.Bucket = env("MINIO_BUCKET", "video-service")
	c.Minio.UseSSL = envBool("MINIO_USE_SSL", false)

	c.Redis.Addr = env("REDIS_ADDR", "localhost:6379")
	c.Redis.Password = os.Getenv("REDIS_PASSWORD")

	c.Worker.Concurrency = int(envInt("WORKER_CONCURRENCY", 1))
	c.Worker.JobTimeout = time.Duration(envInt("TRANSCODE_TIMEOUT_MINUTES", 60)) * time.Minute
	c.Worker.ClaimIdle = c.Worker.JobTimeout + 15*time.Minute

	c.Upload.MaxBytes = envInt("MAX_UPLOAD_MB", 2048) << 20

	c.Stream.ChunkSize = envInt("CHUNK_SIZE_KB", 1024) << 10
	c.Stream.CacheBytes = envInt("CACHE_SIZE_MB", 512) << 20
	c.Stream.PrefetchChunks = int(envInt("PREFETCH_CHUNKS", 2))
	c.Stream.MetaTTL = time.Duration(envInt("META_TTL_SECONDS", 300)) * time.Second
	c.Stream.ListTTL = time.Duration(envInt("LIST_TTL_SECONDS", 30)) * time.Second
	return c
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int64) int64 {
	if v, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil && v > 0 {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	if v, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
