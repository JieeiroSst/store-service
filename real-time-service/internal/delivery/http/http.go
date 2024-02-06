package http

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

func WsCall(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", "http://localhost:8081/ws", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("X-Api-Key", "test_api_key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("auth err: %v", err)
	}
	defer resp.Body.Close()

	// create ws conn
	u := url.URL{Scheme: "ws", Host: "localhost:8081", Path: "/ws"}
	u.RequestURI()
	fmt.Printf("ws url: %s", u.String())
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("dial err: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte("hellow websockets"))
	if err != nil {
		log.Fatalf("msg err: %v", err)
	}
}
