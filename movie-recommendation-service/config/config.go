package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/engine"
)

type Config struct {
	Server struct {
		Port           string
		AllowedOrigins []string
	}
	Postgres struct {
		Host, Port, User, Password, Database, SSLMode string
	}
	VideoService struct {
		BaseURL string
		Timeout time.Duration
	}
	Train struct {
		Window        time.Duration
		SyncInterval  time.Duration
		TrainInterval time.Duration
	}
	Redis struct {
		// Addr, when set, moves ranking snapshots from Postgres to Redis.
		Addr     string
		Password string
	}
	Snapshot struct {
		TTL time.Duration
	}
	Engine engine.Config
}

func Load() *Config {
	c := &Config{}
	c.Server.Port = env("PORT_HTTP_SERVER", "8080")
	c.Server.AllowedOrigins = splitCSV(env("ALLOWED_ORIGINS", "*"))

	c.Postgres.Host = env("POSTGRES_HOST", "localhost")
	c.Postgres.Port = env("POSTGRES_PORT", "5432")
	c.Postgres.User = env("POSTGRES_USER", "postgres")
	c.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	c.Postgres.Database = env("POSTGRES_DATABASE", "movie_recommendation_service")
	c.Postgres.SSLMode = env("POSTGRES_SSLMODE", "disable")

	c.VideoService.BaseURL = strings.TrimRight(env("VIDEO_SERVICE_BASE_URL", "http://localhost:8081"), "/")
	c.VideoService.Timeout = seconds("VIDEO_SERVICE_TIMEOUT_SECONDS", 15)

	c.Train.Window = time.Duration(envInt("TRAIN_WINDOW_DAYS", 90)) * 24 * time.Hour
	c.Train.SyncInterval = seconds("SYNC_INTERVAL_SECONDS", 300)
	c.Train.TrainInterval = seconds("TRAIN_INTERVAL_SECONDS", 60)

	c.Redis.Addr = os.Getenv("REDIS_ADDR")
	c.Redis.Password = os.Getenv("REDIS_PASSWORD")
	c.Snapshot.TTL = time.Duration(envInt("SNAPSHOT_TTL_MINUTES", 15)) * time.Minute

	c.Engine = engine.DefaultConfig()
	c.Engine.HalfLife = time.Duration(envInt("TRENDING_HALF_LIFE_HOURS", 168)) * time.Hour
	c.Engine.NewHalfLife = time.Duration(envInt("NEW_HALF_LIFE_DAYS", 14)) * 24 * time.Hour
	if v, err := strconv.ParseFloat(os.Getenv("DIVERSITY"), 64); err == nil && v >= 0 {
		c.Engine.Diversity = v
	}
	return c
}

func (c *Config) PostgresURL(database string) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.Postgres.User, c.Postgres.Password),
		Host:     fmt.Sprintf("%s:%s", c.Postgres.Host, c.Postgres.Port),
		Path:     "/" + database,
		RawQuery: "sslmode=" + url.QueryEscape(c.Postgres.SSLMode),
	}
	return u.String()
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

func seconds(key string, def int64) time.Duration {
	return time.Duration(envInt(key, def)) * time.Second
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
