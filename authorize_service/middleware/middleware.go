package middleware

import (
	"github.com/JieeiroSst/authorize-service/utils"
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
			ctx.AbortWithStatusJSON(403, gin.H{
				"message": "Unauthorized",
			})
			return
		}
		ctx.Next()
	}
}
