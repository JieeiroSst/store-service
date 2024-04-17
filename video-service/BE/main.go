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
		}

		ctx.JSON(200, res)
	})

	r.Run()
}
