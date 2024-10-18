package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

type Message struct {
	Message        string
	QuotaRemaining int
	OTP            string
	TextID         string
}

func Response(c *gin.Context, code int, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	c.Render(
		http.StatusOK, render.Data{
			ContentType: "application/json",
			Data:        data,
		})
}
