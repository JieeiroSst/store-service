package middleware

import (
	"github.com/JIeeiroSst/upload-service/utils"
	"github.com/gofiber/fiber/v2"
)

type Middleware interface {
	AuthorizeControl() fiber.Handler
}

type middleware struct {
	serect string
}

func Newmiddleware(serect string) Middleware {
	return &middleware{
		serect: serect,
	}
}

func (m *middleware) AuthorizeControl() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authorzation := ctx.Request().Header.Peek("Authorzation")
		if ok := utils.DecodeBase(string(authorzation), m.serect); !ok {
			return ctx.JSON("Unauthorized")
		}
		return ctx.Next()
	}
}
