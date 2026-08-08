package config

type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Scheduler SchedulerConfig
	Telegram  TelegramConfig
	Twitter   TwitterConfig
	Facebook  FacebookConfig
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

type SchedulerConfig struct {
	PollInterval string
}

type TelegramConfig struct {
	BotToken        string
	Debug           bool
	BroadcastChatID string
}

type TwitterConfig struct {
	Enabled             bool
	UserID              string
	BearerToken         string
	ConsumerKey         string
	ConsumerSecret      string
	AccessToken         string
	AccessTokenSecret   string
	PollIntervalSeconds int
}

type FacebookConfig struct {
	Enabled         bool
	PageID          string
	PageAccessToken string
	VerifyToken     string
	AppSecret       string
}
