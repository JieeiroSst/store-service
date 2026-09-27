package queuecheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fake(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); !ok || u != "guest" || p != "pw" {
			w.WriteHeader(401)
			return
		}
		w.Write([]byte(body))
	}))
}

func TestCheckRabbitMQ(t *testing.T) {
	srv := fake(`[
	 {"name":"orders","vhost":"/","messages":10,"messages_ready":10,"messages_unacknowledged":0,"consumers":2,"durable":true},
	 {"name":"payments","vhost":"/","messages":5,"messages_ready":5,"consumers":0,"durable":true},
	 {"name":"orders.dlq","vhost":"/","messages":3,"consumers":0,"durable":true},
	 {"name":"emails","vhost":"/","messages":5000,"messages_ready":5000,"consumers":1,"durable":true},
	 {"name":"stuck","vhost":"/","messages":2000,"messages_ready":0,"messages_unacknowledged":2000,"consumers":1,"durable":true},
	 {"name":"tmp","vhost":"/","messages":1,"messages_ready":1,"consumers":1,"durable":false},
	 {"name":"idle","vhost":"/","messages":0,"consumers":0,"durable":true}]`)
	defer srv.Close()

	got, err := CheckRabbitMQ(context.Background(), srv.Client(), RabbitMQ{URL: srv.URL, User: "guest", Password: "pw"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]int{}
	for _, f := range got {
		ids[f.ID]++
	}
	for id, n := range map[string]int{"Q-DLQ-001": 1, "Q-CONS-001": 1, "Q-BACKLOG-001": 1, "Q-UNACK-001": 1, "Q-DUR-001": 1} {
		if ids[id] != n {
			t.Errorf("%s: want %d got %d (%+v)", id, n, ids[id], got)
		}
	}
	if got[0].Severity != "high" {
		t.Errorf("highest severity must be first: %+v", got)
	}
}

func TestCheckRabbitMQErrors(t *testing.T) {
	srv := fake(`[]`)
	defer srv.Close()
	if _, err := CheckRabbitMQ(context.Background(), srv.Client(), RabbitMQ{URL: srv.URL, User: "x", Password: "y"}, Options{}); err == nil {
		t.Fatal("bad credentials must be an error, not an empty (healthy) report")
	}
	bad := fake(`not json`)
	defer bad.Close()
	if _, err := CheckRabbitMQ(context.Background(), bad.Client(), RabbitMQ{URL: bad.URL, User: "guest", Password: "pw"}, Options{}); err == nil {
		t.Fatal("garbage response must be an error")
	}
}
