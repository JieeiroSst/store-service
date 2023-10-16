package handler

import (
	"strconv"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
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
	userId := c.Query("user_id")
	if userId == "" {
		c.JSON(400, "")
		return
	}
	var project domain.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(400, err)
		return
	}
	if err := h.svc.CreateProject(userId, project); err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, project)
}

func (h *ProjectHandler) ReadProject(c *gin.Context) {
	limmit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))
	pagination := domain.Pagination{
		Limit: limmit,
		Page:  page,
		Sort:  c.Query("sort"),
	}

	projects, err := h.svc.ReadProjects(pagination)
	if err != nil {
		c.JSON(500, err)
	}
	c.JSON(200, projects)
}

func (h *ProjectHandler) ReadProjects(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	project, err := h.svc.ReadProject(id)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, project)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}
	var project domain.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(400, err)
		return
	}
	if err := h.svc.UpdateProject(id, project); err != nil {
		c.JSON(400, err)
		return
	}
	c.JSON(200, "update success")
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, "")
		return
	}

	if err := h.svc.DeleteProject(id); err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(200, "delete success")
}
