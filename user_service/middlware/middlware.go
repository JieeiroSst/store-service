package middlware

import (
	"net/http"

	"github.com/JIeeiroSst/user-service/pb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Middlware struct {
	dns string
}

func NewMiddlware(dns string) *Middlware {
	return &Middlware{
		dns: dns,
	}
}

func (m *Middlware) AccessControl() gin.HandlerFunc {
	conn, err := grpc.Dial(m.dns, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewAuthorizeClient(conn)
	return func(ctx *gin.Context) {
		req := &pb.CasbinRuleRequest{
			Sub: "",
			Obj: ctx.Request.URL.Path,
			Act: ctx.Request.Method,
		}

		_, err := client.EnforceCasbin(ctx, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}
		ctx.Next()
	}
}
