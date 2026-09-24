package logger

type Config struct {
	AppEnv     string
	AppName    string
	AppVersion string
	Level      string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}
