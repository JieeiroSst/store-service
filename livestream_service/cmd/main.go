package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gorilla/websocket"
)


var sessionGroupMap = make(map[string]map[uuid.UUID]*websocket.Conn)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func main() {
	app := gin.Default()
	app.Any("stream", wsHandler)
	app.Run("localhost:8080")
}

func wsHandler(ginContext *gin.Context) { 
	wsSession, err := upgrader.Upgrade(ginContext.Writer, ginContext.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	uid := uuid.New()
	wsURL := ginContext.Request.URL
	wsURLParam, err := url.ParseQuery(wsURL.RawQuery)
	if err != nil {
		wsSession.Close()
		log.Println(err)
	}
	if _, ok := wsURLParam["name"]; ok {
		threadName := wsURLParam["name"][0]
		log.Printf("A client connect to %s", threadName)
		if _, ok := sessionGroupMap[threadName]; ok { 
			sessionGroupMap[threadName][uid] = wsSession
		} else {
			sessionGroupMap[threadName] = make(map[uuid.UUID]*websocket.Conn)
			sessionGroupMap[threadName][uid] = wsSession
		}
		defer wsSession.Close()
		echo(wsSession, threadName, uid)
	} else {
		wsSession.Close()
	}
}
func echo(wsSession *websocket.Conn, threadName string, uid uuid.UUID) {
	for {
		messageType, messageContent, err := wsSession.ReadMessage()
		if messageType == 1 {
			log.Printf("Recv:%s from %s", messageContent, threadName)
			broadcast(threadName, messageContent)
		}
		if err != nil {
			wsSession.Close()
			delete(sessionGroupMap[threadName], uid)
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Printf("Client disconnected in %s", threadName)
			} else {
				log.Printf("Reading Error in %s. %s", threadName, err)
			}
			break 
		}
	}
}
func broadcast(threadName string, messageContent []byte) {
	for _, wsSession := range sessionGroupMap[threadName] {
		err := wsSession.WriteMessage(1, messageContent)
		if err != nil {
			log.Println(err)
		}
	}
}