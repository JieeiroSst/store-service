package airflowclient

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
)

type healthRepository struct {
	client *Client
}

func NewHealthRepository(client *Client) port.HealthRepository {
	return &healthRepository{client: client}
}

func (r *healthRepository) Get(ctx context.Context) (*model.HealthStatus, error) {
	resp, _, err := r.client.API.MonitoringApi.GetHealth(r.client.Auth(ctx)).Execute()
	if err != nil {
		return nil, err
	}

	metadatabase := resp.GetMetadatabase()
	scheduler := resp.GetScheduler()

	return &model.HealthStatus{
		MetadatabaseStatus:       string(metadatabase.GetStatus()),
		SchedulerStatus:          string(scheduler.GetStatus()),
		LatestSchedulerHeartbeat: scheduler.GetLatestSchedulerHeartbeat(),
	}, nil
}
