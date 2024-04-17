package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	apivideosdk "github.com/apivideo/api.video-go-client"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Video struct {
	LiveStreamId string `json:"live_stream_id"`
	StreamKey    string `json:"stream_key"`
	PlayerId     string `json:"player_id"`
}

type CreateVideo struct {
	NameStream string `json:"name_stream"`
}

type PaginationVideo struct {
	StreamKey   string `json:"stream_key" form:"stream_key"`
	Name        string `json:"name" form:"name"`
	SortBy      string `json:"sort_by" form:"sort_by"`
	SortOrder   string `json:"sort_order" form:"sort_order"`
	CurrentPage int    `json:"current_page" form:"current_page"`
	PageSize    int    `json:"page_size" form:"page_size"`
}

func CreateLiveStream(apiKeys, nameStream string) (string, string, string) {
	client := apivideosdk.ClientBuilder(apiKeys).Build()

	liveStreamCreationPayload := *apivideosdk.NewLiveStreamCreationPayload(nameStream)

	res, err := client.LiveStreams.Create(liveStreamCreationPayload)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LiveStreams.Create``: %v\n", err)
	}

	return *res.StreamKey, *res.Assets.Player, res.LiveStreamId
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r.Use(CORSMiddleware())

	secretKey := os.Getenv("API_keys")
	r.POST("/video", func(c *gin.Context) {
		var createVideo CreateVideo
		if err := c.ShouldBindJSON(&createVideo); err != nil {
			c.JSON(400, "")
			return
		}
		streamKey, playerId, liveStreamId := CreateLiveStream(secretKey, createVideo.NameStream)
		video := Video{
			StreamKey:    streamKey,
			PlayerId:     playerId,
			LiveStreamId: liveStreamId,
		}

		c.JSON(http.StatusOK, video)
	})

	r.GET("/video", func(ctx *gin.Context) {
		client := apivideosdk.ClientBuilder(secretKey).Build()
		liveStreamId := ctx.Query("live-stream-id")

		res, err := client.LiveStreams.Get(liveStreamId)
		if err != nil {
			ctx.JSON(500, "")
			return
		}

		ctx.JSON(200, res)
	})

	r.DELETE("/video", func(ctx *gin.Context) {
		client := apivideosdk.ClientBuilder(secretKey).Build()

		liveStreamId := ctx.Query("live-stream-id")

		err := client.LiveStreams.Delete(liveStreamId)

		if err != nil {
			ctx.JSON(500, err)
			return
		}

		ctx.JSON(200, "done")
	})

	r.GET("/pagination", func(ctx *gin.Context) {
		client := apivideosdk.ClientBuilder(secretKey).Build()

		var reqAPi PaginationVideo
		if err := ctx.ShouldBindQuery(&reqAPi); err != nil {
			ctx.JSON(400, err)
			return
		}

		req := apivideosdk.LiveStreamsApiListRequest{}

		if len(reqAPi.StreamKey) > 0 {
			req.StreamKey(reqAPi.StreamKey)
		}

		if len(reqAPi.Name) > 0 {
			req.Name(reqAPi.Name)
		}

		if len(reqAPi.SortBy) > 0 {
			req.SortBy("createdAt")
		}

		if len(reqAPi.SortOrder) > 0 {
			req.SortOrder("desc")
		}

		if reqAPi.CurrentPage > 0 {
			req.CurrentPage(int32(reqAPi.CurrentPage))
		} else {
			req.CurrentPage(int32(0))
		}

		if reqAPi.PageSize > 0 {
			req.PageSize(int32(reqAPi.PageSize))
		} else {
			req.PageSize(int32(30))
		}

		res, err := client.LiveStreams.List(req)
		if err != nil {
			ctx.JSON(500, err)
			return
		}
		ctx.JSON(200, res)
	})

	r.Run()
}
