package config

type Config struct {
	Server    ServerConfig    `json:"server"`
	Postgres  PostgresConfig  `json:"postgres"`
	Scheduler SchedulerConfig `json:"scheduler"`
	Telegram  TelegramConfig  `json:"telegram"`
	Twitter   TwitterConfig   `json:"twitter"`
	Facebook  FacebookConfig  `json:"facebook"`
}

type ServerConfig struct {
	PortHttpServer string `json:"port_http_server"`
}

type PostgresConfig struct {
	PostgresqlHost     string `json:"postgresql_host"`
	PostgresqlPort     string `json:"postgresql_port"`
	PostgresqlUser     string `json:"postgresql_user"`
	PostgresqlPassword string `json:"postgresql_password"`
	PostgresqlDbname   string `json:"postgresql_dbname"`
	PostgresqlSSLMode  bool   `json:"postgresql_ssl_mode"`
}

type SchedulerConfig struct {
	PollInterval string `json:"poll_interval"`
}

type TelegramConfig struct {
	BotToken        string `json:"bot_token"`
	Debug           bool   `json:"debug"`
	BroadcastChatID string `json:"broadcast_chat_id"`
}

type TwitterConfig struct {
	Enabled             bool   `json:"enabled"`
	UserID              string `json:"user_id"`
	BearerToken         string `json:"bearer_token"`
	ConsumerKey         string `json:"consumer_key"`
	ConsumerSecret      string `json:"consumer_secret"`
	AccessToken         string `json:"access_token"`
	AccessTokenSecret   string `json:"access_token_secret"`
	PollIntervalSeconds int    `json:"poll_interval_seconds"`
}

type FacebookConfig struct {
	Enabled         bool   `json:"enabled"`
	PageID          string `json:"page_id"`
	PageAccessToken string `json:"page_access_token"`
	VerifyToken     string `json:"verify_token"`
	AppSecret       string `json:"app_secret"`
}
