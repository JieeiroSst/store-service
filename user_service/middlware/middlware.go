package middlware

import (
	"net/http"

	"github.com/JIeeiroSst/user-service/pb"
	"github.com/JIeeiroSst/user-service/pkg/log"
	"github.com/JIeeiroSst/user-service/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
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
		log.Info(req)

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
