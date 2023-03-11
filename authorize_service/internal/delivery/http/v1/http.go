package v1

import (
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase    *usecase.Usecase
	middleware middleware.Middleware
}

func NewHandler(usecase *usecase.Usecase, middleware middleware.Middleware) *Handler {
	return &Handler{
		usecase:    usecase,
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
		h.initCasbinRoutes(v1)
		h.initOtpRouters(v1)
		h.initSocketRouters(v1)
	}
}
