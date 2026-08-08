package infrastructure

import (
	"github.com/JIeeiroSst/billing-service/config"
	"github.com/JIeeiroSst/billing-service/pkg/consul"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	return consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul).ConnectConfigConsul()
}
