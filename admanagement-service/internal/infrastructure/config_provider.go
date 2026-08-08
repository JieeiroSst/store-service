package infrastructure

import (
	"github.com/JIeeiroSst/admanagement-service/config"
	"github.com/JIeeiroSst/admanagement-service/pkg/consul"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	return consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul).ConnectConfigConsul()
}
