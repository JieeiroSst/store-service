package middleware

import (
	"github.com/JIeeiroSst/search-service/pkg/logger"
	"github.com/JIeeiroSst/search-service/utils"
	"github.com/gin-gonic/gin"
)

type Middleware interface {
	AuthorizeControl() gin.HandlerFunc
}

type middleware struct {
	serect string
}

func Newmiddleware(serect string) Middleware {
	return &middleware{
		serect: serect,
	}
}

func (m *middleware) AuthorizeControl() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorzation := ctx.Request.Header.Get("Authorization")
		if ok := utils.DecodeBase(authorzation, m.serect); !ok {
			logger.Logger().Errorf("Unauthorized failed is %v", ctx.RemoteIP())
			ctx.AbortWithStatusJSON(403, gin.H{
				"message": "Unauthorized",
			})
			return
		}
		ctx.Next()
	}
}
