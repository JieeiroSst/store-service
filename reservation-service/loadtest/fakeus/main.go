// A user-service stand-in that accepts tokens of the form "u<id>" and counts how often it is called.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

var validates, finds atomic.Int64

func main() {
	http.HandleFunc("POST /api/v1/validate", func(w http.ResponseWriter, r *http.Request) {
		validates.Add(1)
		var b struct {
			T string `json:"session_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		id := map[string]int{"admin": 1, "mgr": 5}[b.T]
		if id == 0 && strings.HasPrefix(b.T, "u") {
			id, _ = strconv.Atoi(b.T[1:])
		}
		w.Header().Set("Content-Type", "application/json")
		if id == 0 {
			fmt.Fprint(w, `{}`)
			return
		}
		fmt.Fprintf(w, `{"valid":true,"user_id":"%d"}`, id)
	})
	http.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		finds.Add(1)
		id, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
		role := "guest"
		if id == 1 {
			role = "admin"
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"users":{"id":%d,"email":"u%d@x.com","roles":[{"id":1,"name":"%s"}]}}`, id, id, role)
	})
	http.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"validate":%d,"find":%d}`, validates.Load(), finds.Load())
	})
	_ = http.ListenAndServe("0.0.0.0:18999", nil)
}
