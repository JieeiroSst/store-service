package v1

import (
	"github.com/JIeeiroSst/search-service/internal/service"
	"github.com/JIeeiroSst/search-service/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	middleware middleware.Middleware
	service    service.Service
}

func NewHandler(service service.Service, middleware middleware.Middleware) *Handler {
	return &Handler{
		service:    service,
		middleware: middleware,
	}
}

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initSearchTaskRoutes(v1)
	}
}
