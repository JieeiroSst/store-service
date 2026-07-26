package http

import (
	"strconv"

	"github.com/JIeeiroSst/airflow-service/internal/domain/port"
	"gofr.dev/pkg/gofr"
)

type Handler struct {
	dag          port.DAGUsecase
	dagRun       port.DAGRunUsecase
	taskInstance port.TaskInstanceUsecase
	variable     port.VariableUsecase
	pool         port.PoolUsecase
	health       port.HealthUsecase
}

func NewHandler(
	dag port.DAGUsecase,
	dagRun port.DAGRunUsecase,
	taskInstance port.TaskInstanceUsecase,
	variable port.VariableUsecase,
	pool port.PoolUsecase,
	health port.HealthUsecase,
) *Handler {
	return &Handler{
		dag:          dag,
		dagRun:       dagRun,
		taskInstance: taskInstance,
		variable:     variable,
		pool:         pool,
		health:       health,
	}
}

// pagination reads the standard limit/offset query parameters used across the
// Airflow REST API, defaulting limit to 100 as Airflow itself does.
func pagination(c *gofr.Context) (limit, offset int32) {
	limit = 100
	if v, err := strconv.Atoi(c.Param("limit")); err == nil {
		limit = int32(v)
	}
	if v, err := strconv.Atoi(c.Param("offset")); err == nil {
		offset = int32(v)
	}
	return limit, offset
}
