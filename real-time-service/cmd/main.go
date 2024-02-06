package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Payment struct {
	ID   string
	Name string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Middleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}

func authMiddleware(next http.Handler) http.Handler {
	TestApiKey := "test_api_key"
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var apiKey string
		if apiKey = req.Header.Get("X-Api-Key"); apiKey != TestApiKey {
			log.Printf("bad auth api key: %s", apiKey)
			rw.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(rw, req)
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		fmt.Println("Received message:", string(message))

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			return
		}
	}
}

func wsCall(w http.ResponseWriter, r *http.Request) {
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

func main() {
	http.Handle("/ws", Middleware(
		http.HandlerFunc(wsHandler),
		authMiddleware,
	))

	http.Handle("/call", Middleware(
		http.HandlerFunc(wsCall),
		authMiddleware,
	))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
