package config

import (
	"os"
	"strings"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL               string
	RedisURL                  string
	Port                      string
	AppEnv                    string
	CORSOrigin                string
	TrustProxy                domain.TrustProxy
	IOSTeamID                 string
	IOSBundleID               string
	AndroidPackageName        string
	AndroidSHA256Fingerprints string
	ShortlinkDomain           string
	TrustEdgeBotHeader        bool
	GeoIPDBPath               string
	AbuseReportURL            string
}

func Load() Config {
	_ = godotenv.Load()

	appEnv := getenv("APP_ENV", "")
	if appEnv == "" {
		appEnv = getenv("NODE_ENV", "development")
	}

	return Config{
		DatabaseURL:               getenv("DATABASE_URL", "postgresql://postgres:password@localhost:5432/linkforty"),
		RedisURL:                  os.Getenv("REDIS_URL"),
		Port:                      getenv("PORT", "3000"),
		AppEnv:                    appEnv,
		CORSOrigin:                getenv("CORS_ORIGIN", "*"),
		TrustProxy:                domain.ParseTrustProxyEnv(os.Getenv("TRUST_PROXY")),
		IOSTeamID:                 os.Getenv("IOS_TEAM_ID"),
		IOSBundleID:               os.Getenv("IOS_BUNDLE_ID"),
		AndroidPackageName:        os.Getenv("ANDROID_PACKAGE_NAME"),
		AndroidSHA256Fingerprints: os.Getenv("ANDROID_SHA256_FINGERPRINTS"),
		ShortlinkDomain:           os.Getenv("SHORTLINK_DOMAIN"),
		TrustEdgeBotHeader:        strings.EqualFold(os.Getenv("TRUST_EDGE_BOT_HEADER"), "true"),
		GeoIPDBPath:               os.Getenv("GEOIP_DB_PATH"),
		AbuseReportURL:            os.Getenv("ABUSE_REPORT_URL"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
