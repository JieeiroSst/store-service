package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server  ServerConfig
	Collage CollageConfig
	Minio   MinioConfig
	Mysql   MysqlConfig
}

type ServerConfig struct {
	PortHttpServer string
	MaxUploadBytes int64
}

type CollageConfig struct {
	CellWidth   int
	CellHeight  int
	Columns     int
	MaxImages   int
	MaxPixels   int
	JPEGQuality  int
	CenterWidth  int
	CenterHeight int
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type MysqlConfig struct {
	MysqlHost     string
	MysqlPort     string
	MysqlUser     string
	MysqlPassword string
	MysqlDbname   string
}

func FromEnv() *Config {
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8080"),
			MaxUploadBytes: int64(getInt("MAX_UPLOAD_MB", 100)) << 20,
		},
		Collage: CollageConfig{
			CellWidth:   getInt("CELL_WIDTH", 100),
			CellHeight:  getInt("CELL_HEIGHT", 100),
			Columns:     getInt("DEFAULT_COLUMNS", 2),
			MaxImages:   getInt("MAX_IMAGES", 50),
			MaxPixels:   getInt("MAX_PIXELS", 50_000_000),
			JPEGQuality:  getInt("JPEG_QUALITY", 95),
			CenterWidth:  getInt("CENTER_WIDTH", 800),
			CenterHeight: getInt("CENTER_HEIGHT", 800),
		},
		Minio: MinioConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "draw-image-service"),
			UseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		},
		Mysql: MysqlConfig{
			MysqlHost:     getEnv("MYSQL_HOST", "localhost"),
			MysqlPort:     getEnv("MYSQL_PORT", "3306"),
			MysqlUser:     getEnv("MYSQL_USER", "root"),
			MysqlPassword: getEnv("MYSQL_PASSWORD", ""),
			MysqlDbname:   getEnv("MYSQL_DBNAME", "draw_image_service"),
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
