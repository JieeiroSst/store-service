package config

type Config struct {
	Server   ServerConfig
}

type ServerConfig struct {
	ServerPort string
	GRPCServer string
}