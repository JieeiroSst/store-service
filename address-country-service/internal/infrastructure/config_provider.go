package infrastructure

import (
	"github.com/JIeeiroSst/address-country-service/config"
	"github.com/JIeeiroSst/address-country-service/pkg/consul"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	consulClient := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)

	return consulClient.ConnectConfigConsul()
}
