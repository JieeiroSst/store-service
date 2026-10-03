package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	WebSocket WebSocketConfig
	Room      RoomConfig
	ICE       ICEConfig
}

type ServerConfig struct {
	PortHttpServer string
	AllowedOrigins []string
}

type WebSocketConfig struct {
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageBytes int64
	SendBuffer      int
}

type RoomConfig struct {
	MaxPeers int
}

type ICEConfig struct {
	STUNURLs   []string
	TURNURLs   []string
	TURNSecret string
	TURNTTL    time.Duration
}

func FromEnv() *Config {
	pongWait := time.Duration(getInt("WS_PONG_WAIT_SECONDS", 60)) * time.Second
	return &Config{
		Server: ServerConfig{
			PortHttpServer: getEnv("PORT_HTTP_SERVER", "8080"),
			AllowedOrigins: getList("ALLOWED_ORIGINS"),
		},
		WebSocket: WebSocketConfig{
			WriteWait:  10 * time.Second,
			PongWait:   pongWait,
			PingPeriod: pongWait * 9 / 10,

			MaxMessageBytes: int64(getInt("WS_MAX_MESSAGE_KB", 64)) << 10,
			SendBuffer:      getInt("WS_SEND_BUFFER", 256),
		},
		Room: RoomConfig{
			MaxPeers: getInt("MAX_PEERS_PER_ROOM", 8),
		},
		ICE: ICEConfig{
			STUNURLs:   getListOr("STUN_URLS", []string{"stun:stun.l.google.com:19302"}),
			TURNURLs:   getList("TURN_URLS"),
			TURNSecret: os.Getenv("TURN_SECRET"),
			TURNTTL:    time.Duration(getInt("TURN_TTL_SECONDS", 86400)) * time.Second,
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

func getListOr(key string, fallback []string) []string {
	if _, ok := os.LookupEnv(key); !ok {
		return fallback
	}
	return getList(key)
}

func getList(key string) []string {
	var out []string
	for _, v := range strings.Split(os.Getenv(key), ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
