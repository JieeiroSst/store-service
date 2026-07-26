package infrastructure

import (
	"github.com/JIeeiroSst/airflow-service/config"
	"github.com/JIeeiroSst/airflow-service/pkg/consul"
)

func newConfig() (*config.Config, error) {
	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	consulClient := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)

	return consulClient.ConnectConfigConsul()
}
