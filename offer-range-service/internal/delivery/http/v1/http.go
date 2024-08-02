package v1

import (
	"github.com/JIeeiroSst/offer-range-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase *usecase.Usecase
}

func NewHandler(usecase *usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// @title           User Service API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:1235
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
func (h *Handler) Init(api *gin.RouterGroup) {
	_ = api.Group("/v1")
	{

	}
}
