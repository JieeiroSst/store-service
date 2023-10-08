package handler

import "github.com/JIeeiroSst/partner-service/internal/core/services"

type ProjectHandler struct {
	svc services.ProjectService
}

func NewProjectHandler(svc services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		svc: svc,
	}
}
