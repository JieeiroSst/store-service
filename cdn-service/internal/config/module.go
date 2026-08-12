package config

import "go.uber.org/fx"

func NewConfig() (*Config, error) {
	dir, err := ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}
	return loadFromConsul(dir.HostConsul, dir.KeyConsul, dir.ServiceConsul)
}

var Module = fx.Options(
	fx.Provide(NewConfig),
)
