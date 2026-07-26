package airflowclient

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/config"
	"github.com/apache/airflow-client-go/airflow"
)

type Client struct {
	API      *airflow.APIClient
	username string
	password string
}

func NewClient(cfg *config.Config) *Client {
	conf := airflow.NewConfiguration()
	conf.Host = cfg.Airflow.Host
	conf.Scheme = cfg.Airflow.Scheme

	return &Client{
		API:      airflow.NewAPIClient(conf),
		username: cfg.Airflow.Username,
		password: cfg.Airflow.Password,
	}
}

func (c *Client) Auth(ctx context.Context) context.Context {
	return context.WithValue(ctx, airflow.ContextBasicAuth, airflow.BasicAuth{
		UserName: c.username,
		Password: c.password,
	})
}
