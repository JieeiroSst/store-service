package airflowclient

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
	"github.com/apache/airflow-client-go/airflow"
)

type poolRepository struct {
	client *Client
}

func NewPoolRepository(client *Client) port.PoolRepository {
	return &poolRepository{client: client}
}

func (r *poolRepository) Create(ctx context.Context, pool *model.Pool) (*model.Pool, error) {
	body := airflow.NewPool()
	body.SetName(pool.Name)
	body.SetSlots(pool.Slots)
	if pool.Description != "" {
		body.SetDescription(pool.Description)
	}

	resp, _, err := r.client.API.PoolApi.PostPool(r.client.Auth(ctx)).Pool(*body).Execute()
	if err != nil {
		return nil, err
	}

	p := toPoolModel(resp)
	return &p, nil
}

func (r *poolRepository) Get(ctx context.Context, name string) (*model.Pool, error) {
	resp, _, err := r.client.API.PoolApi.GetPool(r.client.Auth(ctx), name).Execute()
	if err != nil {
		return nil, err
	}

	p := toPoolModel(resp)
	return &p, nil
}

func (r *poolRepository) Update(ctx context.Context, pool *model.Pool) (*model.Pool, error) {
	body := airflow.NewPool()
	body.SetName(pool.Name)
	body.SetSlots(pool.Slots)
	if pool.Description != "" {
		body.SetDescription(pool.Description)
	}

	resp, _, err := r.client.API.PoolApi.PatchPool(r.client.Auth(ctx), pool.Name).Pool(*body).Execute()
	if err != nil {
		return nil, err
	}

	p := toPoolModel(resp)
	return &p, nil
}

func (r *poolRepository) Delete(ctx context.Context, name string) error {
	_, err := r.client.API.PoolApi.DeletePool(r.client.Auth(ctx), name).Execute()
	return err
}

func (r *poolRepository) List(ctx context.Context, limit, offset int32) (*model.PoolList, error) {
	resp, _, err := r.client.API.PoolApi.GetPools(r.client.Auth(ctx)).Limit(limit).Offset(offset).Execute()
	if err != nil {
		return nil, err
	}

	pools := make([]model.Pool, 0, len(resp.GetPools()))
	for _, p := range resp.GetPools() {
		pools = append(pools, toPoolModel(p))
	}

	return &model.PoolList{Pools: pools, TotalEntries: resp.GetTotalEntries()}, nil
}

func toPoolModel(p airflow.Pool) model.Pool {
	return model.Pool{
		Name:          p.GetName(),
		Slots:         p.GetSlots(),
		OccupiedSlots: p.GetOccupiedSlots(),
		UsedSlots:     p.GetUsedSlots(),
		QueuedSlots:   p.GetQueuedSlots(),
		OpenSlots:     p.GetOpenSlots(),
		Description:   p.GetDescription(),
	}
}
