package handler

import (
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	svc services.ProjectService
}

func NewProjectHandler(svc services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		svc: svc,
	}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {

}

func (h *ProjectHandler) ReadProject(c *gin.Context) {
	
}

func (h *ProjectHandler) ReadProjects(c *gin.Context) {
	
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	
}
