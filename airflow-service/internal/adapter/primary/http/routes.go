package http

import "gofr.dev/pkg/gofr"

func RegisterRoutes(app *gofr.App, h *Handler) {
	app.GET("/health", h.GetHealth)

	app.GET("/dags", h.ListDags)
	app.GET("/dags/{dagId}", h.GetDag)
	app.PATCH("/dags/{dagId}", h.SetDagPaused)

	app.POST("/dags/{dagId}/dagRuns", h.TriggerDagRun)
	app.GET("/dags/{dagId}/dagRuns", h.ListDagRuns)
	app.GET("/dags/{dagId}/dagRuns/{dagRunId}", h.GetDagRun)
	app.DELETE("/dags/{dagId}/dagRuns/{dagRunId}", h.DeleteDagRun)

	app.GET("/dags/{dagId}/dagRuns/{dagRunId}/taskInstances", h.ListTaskInstances)
	app.GET("/dags/{dagId}/dagRuns/{dagRunId}/taskInstances/{taskId}", h.GetTaskInstance)

	app.POST("/variables", h.CreateVariable)
	app.GET("/variables", h.ListVariables)
	app.GET("/variables/{key}", h.GetVariable)
	app.PATCH("/variables/{key}", h.UpdateVariable)
	app.DELETE("/variables/{key}", h.DeleteVariable)

	app.POST("/pools", h.CreatePool)
	app.GET("/pools", h.ListPools)
	app.GET("/pools/{name}", h.GetPool)
	app.PATCH("/pools/{name}", h.UpdatePool)
	app.DELETE("/pools/{name}", h.DeletePool)
}
