package airflowclient

import (
	"context"

	"github.com/JIeeiroSst/airflow-service/internal/domain/model"
	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
	"github.com/apache/airflow-client-go/airflow"
)

type variableRepository struct {
	client *Client
}

func NewVariableRepository(client *Client) port.VariableRepository {
	return &variableRepository{client: client}
}

func (r *variableRepository) Create(ctx context.Context, variable *model.Variable) (*model.Variable, error) {
	body := airflow.NewVariable()
	body.SetKey(variable.Key)
	body.SetValue(variable.Value)
	if variable.Description != "" {
		body.SetDescription(variable.Description)
	}

	resp, _, err := r.client.API.VariableApi.PostVariables(r.client.Auth(ctx)).Variable(*body).Execute()
	if err != nil {
		return nil, err
	}

	v := toVariableModel(resp)
	return &v, nil
}

func (r *variableRepository) Get(ctx context.Context, key string) (*model.Variable, error) {
	resp, _, err := r.client.API.VariableApi.GetVariable(r.client.Auth(ctx), key).Execute()
	if err != nil {
		return nil, err
	}

	v := toVariableModel(resp)
	return &v, nil
}

func (r *variableRepository) Update(ctx context.Context, variable *model.Variable) (*model.Variable, error) {
	body := airflow.NewVariable()
	body.SetKey(variable.Key)
	body.SetValue(variable.Value)
	if variable.Description != "" {
		body.SetDescription(variable.Description)
	}

	resp, _, err := r.client.API.VariableApi.PatchVariable(r.client.Auth(ctx), variable.Key).Variable(*body).Execute()
	if err != nil {
		return nil, err
	}

	v := toVariableModel(resp)
	return &v, nil
}

func (r *variableRepository) Delete(ctx context.Context, key string) error {
	_, err := r.client.API.VariableApi.DeleteVariable(r.client.Auth(ctx), key).Execute()
	return err
}

func (r *variableRepository) List(ctx context.Context, limit, offset int32) (*model.VariableList, error) {
	resp, _, err := r.client.API.VariableApi.GetVariables(r.client.Auth(ctx)).Limit(limit).Offset(offset).Execute()
	if err != nil {
		return nil, err
	}

	variables := make([]model.Variable, 0, len(resp.GetVariables()))
	for _, v := range resp.GetVariables() {
		variables = append(variables, model.Variable{
			Key:         v.GetKey(),
			Description: v.GetDescription(),
		})
	}

	return &model.VariableList{Variables: variables, TotalEntries: resp.GetTotalEntries()}, nil
}

func toVariableModel(v airflow.Variable) model.Variable {
	return model.Variable{
		Key:         v.GetKey(),
		Value:       v.GetValue(),
		Description: v.GetDescription(),
	}
}
