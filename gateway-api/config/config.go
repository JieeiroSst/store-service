package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	PortServer string `json:"port_server"`
}

type ServiceConfig struct {
	HttpAddress       string            `json:"http_address"`
	GrpcAddress       string            `json:"grpc_address"`
	TimeoutSeconds    int               `json:"timeout_seconds"`
	BasicAuthUser     string            `json:"basic_auth_user,omitempty"`
	BasicAuthPassword string            `json:"basic_auth_password,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	ForwardUserToken  *bool             `json:"forward_user_token,omitempty"`
}

type Config struct {
	Server   ServerConfig             `json:"server"`
	Services map[string]ServiceConfig `json:"services"`
}

type Dir struct {
	HostConsul    string
	KeyConsul     string
	ServiceConsul string
}

func ReadFileEnv(dir string) (*Dir, error) {
	err := godotenv.Load(dir)
	if err != nil {
		return nil, err
	}

	data := &Dir{
		HostConsul:    os.Getenv("HostConsul"),
		KeyConsul:     os.Getenv("KeyConsul"),
		ServiceConsul: os.Getenv("ServiceConsul"),
	}
	return data, nil
}
