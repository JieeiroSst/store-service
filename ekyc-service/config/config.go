package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server      ServerConfig
	Postgres    PostgresConfig
	UserService UserServiceConfig
	Storage     StorageConfig
	Ekyc        EkycConfig
	Providers   ProvidersConfig
	Python      PythonConfig
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

type UserServiceConfig struct {
	BaseURL string
	Timeout string
}

func (u UserServiceConfig) TimeoutDuration() time.Duration {
	if d, err := time.ParseDuration(u.Timeout); err == nil {
		return d
	}
	return 5 * time.Second
}

type EkycConfig struct {
	MRZMinConfidence   float64
	FaceMatchThreshold float64
	FaceMinQuality     float64
}

type StorageConfig struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type ProvidersConfig struct {
	CardReader   string
	FaceAnalyzer string
}

type PythonConfig struct {
	Bin           string
	ScriptPath    string
	TimeoutString string
}

func (p PythonConfig) Timeout() time.Duration {
	if d, err := time.ParseDuration(p.TimeoutString); err == nil {
		return d
	}
	return 30 * time.Second
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8087"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", ""),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "ekyc"),
			PostgresqlSSLMode:  getEnv("POSTGRES_SSLMODE", "disable") == "require",
		},
		UserService: UserServiceConfig{
			BaseURL: getEnv("USER_SERVICE_BASE_URL", "http://user-service-svc"),
			Timeout: getEnv("USER_SERVICE_TIMEOUT", "5s"),
		},
		Storage: StorageConfig{
			Endpoint:  getEnv("S3_ENDPOINT", "localhost:9000"),
			Bucket:    getEnv("S3_BUCKET", "ekyc-service"),
			AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""),
			UseSSL:    getEnv("S3_USE_SSL", "false") == "true",
		},
		Ekyc: EkycConfig{
			MRZMinConfidence:   getEnvFloat("EKYC_MRZ_MIN_CONFIDENCE", 0.85),
			FaceMatchThreshold: getEnvFloat("EKYC_FACE_MATCH_THRESHOLD", 0.6),
			FaceMinQuality:     getEnvFloat("EKYC_FACE_MIN_QUALITY", 15.0),
		},
		Providers: ProvidersConfig{
			CardReader:   getEnv("CARD_READER_PROVIDER", "native"),
			FaceAnalyzer: getEnv("FACE_ANALYZER_PROVIDER", "native"),
		},
		Python: PythonConfig{
			Bin:           getEnv("PYTHON_BIN", "python3"),
			ScriptPath:    getEnv("PYTHON_EXTRACT_SCRIPT", "python/extract_id.py"),
			TimeoutString: getEnv("PYTHON_TIMEOUT", "30s"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
