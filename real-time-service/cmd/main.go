package main


import (
    "fmt"
    "log"
    "net/http"

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

func main() {
    http.Handle("/ws", Middleware(
        http.HandlerFunc(wsHandler),
        authMiddleware,
    ))
    log.Fatal(http.ListenAndServe(":8081", nil))
}

