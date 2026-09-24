package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server       ServerConfig
	Mysql        MysqlConfig
	Notification NotificationConfig
	Scheduler    SchedulerConfig
	Signature    SignatureConfig
	Company      CompanyConfig
	Temporal     TemporalConfig
	Storage      StorageConfig
	Files        FilesConfig
	Auth         AuthConfig
}

type ServerConfig struct {
	Port   string
	APIKey string
}

type MysqlConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type SchedulerConfig struct {
	ContractExpiryCron string
}

type SignatureConfig struct {
	TrustRootsFile string
	AllowUntrusted bool
	Revocation     string
	HTTPTimeout    time.Duration
	TSAURL         string
	TSARootsFile   string
}

type CompanyConfig struct {
	Name           string
	Address        string
	TaxCode        string
	Phone          string
	Representative string
	Title          string
	Place          string
	SignKeyFile    string
	SignCertFile   string
}

type TemporalConfig struct {
	Address   string
	Namespace string
	TaskQueue string
	Reminders []time.Duration
}

type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type FilesConfig struct {
	MaxBytes     int64
	PDFPolicy    string
	ScanAddress  string
	ScanTimeout  time.Duration
	ScanRequired bool
}

type AuthConfig struct {
	APIKeyRole   string
	OIDCIssuer   string
	OIDCJWKSURL  string
	OIDCAudience string
	RolesClaim   string
	RolePrefix   string
}

type NotificationConfig struct {
	URL       string
	Type      string // slack | email | push
	Recipient string
	Timeout   time.Duration
}

func Load() *Config {
	_ = godotenv.Load(".env")

	timeout := durationOr(get("NOTIFICATION_TIMEOUT", "3s"), 3*time.Second)

	return &Config{
		Server: ServerConfig{Port: get("PORT", "8080"), APIKey: get("API_KEY", "")},
		Mysql: MysqlConfig{
			Host:     get("MYSQL_HOST", "localhost"),
			Port:     get("MYSQL_PORT", "3306"),
			User:     get("MYSQL_USER", "root"),
			Password: get("MYSQL_PASSWORD", ""),
			DBName:   get("MYSQL_DBNAME", "customer_relationship_service"),
		},
		Scheduler: SchedulerConfig{ContractExpiryCron: get("CONTRACT_EXPIRY_CRON", "0 * * * *")},
		Auth: AuthConfig{
			APIKeyRole:   get("API_KEY_ROLE", "admin"),
			OIDCIssuer:   strings.TrimRight(get("OIDC_ISSUER", ""), "/"),
			OIDCJWKSURL:  get("OIDC_JWKS_URL", ""),
			OIDCAudience: get("OIDC_AUDIENCE", ""),
			RolesClaim:   get("OIDC_ROLES_CLAIM", "realm_access.roles"),
			RolePrefix:   get("ROLE_PREFIX", "crm-"),
		},
		Files: FilesConfig{
			MaxBytes:     int64(intOr(get("CONTRACT_FILE_MAX_MB", "25"), 25)) << 20,
			PDFPolicy:    strings.ToLower(get("PDF_POLICY", "strict")),
			ScanAddress:  get("CLAMAV_ADDRESS", ""),
			ScanTimeout:  durationOr(get("CLAMAV_TIMEOUT", "30s"), 30*time.Second),
			ScanRequired: get("CLAMAV_REQUIRED", "true") != "false",
		},
		Storage: StorageConfig{
			Endpoint:  get("MINIO_ENDPOINT", ""),
			AccessKey: get("MINIO_ACCESS_KEY", ""),
			SecretKey: get("MINIO_SECRET_KEY", ""),
			Bucket:    get("MINIO_BUCKET", "customer-relationship-service"),
			UseSSL:    get("MINIO_USE_SSL", "") == "true",
		},
		Temporal: TemporalConfig{
			Address:   get("TEMPORAL_ADDRESS", ""),
			Namespace: get("TEMPORAL_NAMESPACE", "default"),
			TaskQueue: get("TEMPORAL_TASK_QUEUE", "crm-contract-lifecycle"),
			Reminders: reminderDays(get("TEMPORAL_REMINDER_DAYS", "30,7,1")),
		},
		Company: CompanyConfig{
			Name:           get("COMPANY_NAME", ""),
			Address:        get("COMPANY_ADDRESS", ""),
			TaxCode:        get("COMPANY_TAX_CODE", ""),
			Phone:          get("COMPANY_PHONE", ""),
			Representative: get("COMPANY_REPRESENTATIVE", ""),
			Title:          get("COMPANY_REPRESENTATIVE_TITLE", ""),
			Place:          get("COMPANY_PLACE", ""),
			SignKeyFile:    get("COMPANY_SIGN_KEY_FILE", ""),
			SignCertFile:   get("COMPANY_SIGN_CERT_FILE", ""),
		},
		Signature: SignatureConfig{
			TrustRootsFile: get("SIGNATURE_TRUST_ROOTS_FILE", ""),
			AllowUntrusted: get("SIGNATURE_ALLOW_UNTRUSTED", "") == "true",
			Revocation:     get("SIGNATURE_REVOCATION", "hard"),
			HTTPTimeout:    durationOr(get("SIGNATURE_HTTP_TIMEOUT", "10s"), 10*time.Second),
			TSAURL:         get("SIGNATURE_TSA_URL", ""),
			TSARootsFile:   get("SIGNATURE_TSA_ROOTS_FILE", ""),
		},
		Notification: NotificationConfig{
			URL:       strings.TrimRight(get("NOTIFICATION_SERVICE_URL", ""), "/"),
			Type:      get("NOTIFICATION_TYPE", "slack"),
			Recipient: get("NOTIFICATION_RECIPIENT", ""),
			Timeout:   timeout,
		},
	}
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durationOr(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}

func reminderDays(s string) []time.Duration {
	var out []time.Duration
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && n > 0 {
			out = append(out, time.Duration(n)*24*time.Hour)
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
