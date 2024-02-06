package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/JIeeiroSst/real-time-service/config"
	"github.com/gorilla/websocket"
)

type HttpDelivery struct {
	config *config.Config
}

func NewHttpDelivery(config *config.Config) *HttpDelivery {
	return &HttpDelivery{
		config: config,
	}
}

func (ww *HttpDelivery) WsCall(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%v/ws", ww.config.Server.ServerPort), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("X-Api-Key", ww.config.Serect.Key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("auth err: %v", err)
	}
	defer resp.Body.Close()

	// create ws conn
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("http://localhost:%v", ww.config.Server.ServerPort), Path: "/ws"}
	u.RequestURI()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("dial err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, reqBody)
	if err != nil {
		log.Fatalf("msg err: %v", err)
	}
	fmt.Fprintf(w, `{"message": "successfully"}`)
}
