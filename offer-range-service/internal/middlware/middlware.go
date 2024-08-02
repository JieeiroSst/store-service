package middlware

import (
	"github.com/JIeeiroSst/offer-range-service/utils"
	"github.com/gin-gonic/gin"
)

type Middlware struct {
	dns    string
	serect string
}

func NewMiddlware(dns string, serect string) *Middlware {
	return &Middlware{
		dns:    dns,
		serect: serect,
	}
}

func (m *Middlware) AuthorizeControlHeader() gin.HandlerFunc {
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
