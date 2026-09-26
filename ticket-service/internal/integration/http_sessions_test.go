package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestHTTPSessions(t *testing.T) {
	a := startApp(t, nil)
	const org, adm = int64(1_000_000_002), int64(1_000_000_001)
	eid, tid := publishedEvent(a, 5)
	ep := fmt.Sprintf("/api/v1/events/%d", eid)

	first := a.must(200, "GET", ep+"/sessions", 0, nil)["items"].([]any)
	if len(first) != 1 {
		t.Fatalf("an event has its one showtime: %v", first)
	}
	s1 := int64(first[0].(map[string]any)["id"].(float64))

	start := time.Now().Add(200 * time.Hour).UTC().Truncate(time.Second)
	s2m := a.must(201, "POST", ep+"/sessions", org, map[string]any{"starts_at": start, "ends_at": start.Add(2 * time.Hour), "label": "Matinee", "copy_from": s1})
	s2 := id(s2m)
	if s2m["status"] != "scheduled" || s2m["label"] != "Matinee" {
		t.Fatalf("session: %v", s2m)
	}
	a.must(403, "POST", ep+"/sessions", 5, map[string]any{"starts_at": start, "ends_at": start.Add(time.Hour)})
	a.must(400, "POST", ep+"/sessions", org, map[string]any{"starts_at": start, "ends_at": start.Add(-time.Hour)})
	a.must(401, "POST", ep+"/sessions", 0, map[string]any{"starts_at": start, "ends_at": start.Add(time.Hour)})

	ev := a.must(200, "GET", ep, 0, nil)
	if len(ev["sessions"].([]any)) != 2 {
		t.Fatalf("event: %v", ev["sessions"])
	}
	if items := a.must(200, "GET", "/api/v1/events", 0, nil)["items"].([]any); len(items) != 1 || items[0].(map[string]any)["next_session_at"] == nil {
		t.Fatalf("a listing shows the next showtime once per event: %v", items)
	}
	var t2 int64
	for _, x := range ev["ticket_types"].([]any) {
		if tt := x.(map[string]any); int64(tt["session_id"].(float64)) == s2 {
			t2 = id(tt)
		}
	}
	if t2 == 0 || t2 == tid {
		t.Fatalf("the copied ticket type: %v", ev["ticket_types"])
	}

	// each showtime sells its own; an order says which one it is for
	o1 := buy(a, 5, eid, tid, 5)
	o2 := buy(a, 6, eid, t2, 1)
	if int64(o1["session_id"].(float64)) != s1 || int64(o2["session_id"].(float64)) != s2 {
		t.Fatalf("orders: %v %v", o1["session_id"], o2["session_id"])
	}
	if tk := o2["tickets"].([]any)[0].(map[string]any); int64(tk["session_id"].(float64)) != s2 {
		t.Fatalf("ticket: %v", tk)
	}
	a.must(409, "POST", "/api/v1/orders", 7, map[string]any{"event_id": eid, "buyer_name": "B", "request_id": "sold-out-1", "items": []map[string]any{{"ticket_type_id": tid, "quantity": 1}}})

	// move one showtime, then cancel it: only its buyers are refunded
	moved := start.Add(24 * time.Hour)
	a.must(200, "PUT", fmt.Sprintf("%s/sessions/%d", ep, s2), org, map[string]any{"starts_at": moved, "ends_at": moved.Add(2 * time.Hour), "label": "Matinee"})
	a.must(409, "DELETE", fmt.Sprintf("%s/sessions/%d", ep, s2), org, nil)
	a.must(400, "POST", fmt.Sprintf("%s/sessions/%d/cancel", ep, s2), org, map[string]any{"reason": ""})
	c := a.must(200, "POST", fmt.Sprintf("%s/sessions/%d/cancel", ep, s2), org, map[string]any{"reason": "weather"})
	if c["status"] != "cancelled" || c["cancel_reason"] != "weather" {
		t.Fatalf("cancel: %v", c)
	}
	a.must(409, "POST", "/api/v1/orders", 8, map[string]any{"event_id": eid, "buyer_name": "B", "request_id": "after-cancel", "items": []map[string]any{{"ticket_type_id": t2, "quantity": 1}}})
	// the event's own dates cannot be edited any more; nothing else is affected
	ev = a.must(200, "GET", ep, 0, nil)
	if ev["status"] != "published" {
		t.Fatalf("the event goes on: %v", ev["status"])
	}
	a.must(200, "GET", fmt.Sprintf("/api/v1/orders/%d", id(o1)), 5, nil)
	_ = adm
}

func TestHTTPSeatMapEditorPage(t *testing.T) {
	a := startApp(t, nil)
	const org = int64(1_000_000_002)
	code, h, body := a.raw("GET", "/editor/seat-map", 0, nil)
	if code != 200 || !strings.HasPrefix(h.Get("Content-Type"), "text/html") {
		t.Fatalf("editor: %d %s", code, h.Get("Content-Type"))
	}
	for _, want := range []string{"Seat map editor", "/venues/preview", "/seat-map/apply"} {
		if !strings.Contains(body, want) {
			t.Errorf("the page lacks %q", want)
		}
	}
	csp := h.Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'none'") || !strings.Contains(csp, "connect-src 'self'") || strings.Contains(csp, "http") {
		t.Errorf("CSP: %q", csp)
	}
	if h.Get("X-Frame-Options") != "DENY" || h.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("headers: %v", h)
	}
	// what the editor starts from: a template comes back as its blueprint, which previews to the same seats
	pv := a.must(200, "POST", "/api/v1/venues/preview", org, map[string]any{"template": "hall", "params": map[string]any{"rows": 3, "seats_per_row": 4}})
	m, ok := pv["map"].(map[string]any)
	if !ok || len(m["sections"].([]any)) == 0 {
		t.Fatalf("a template has to come back as a blueprint: %v", pv)
	}
	pv2 := a.must(200, "POST", "/api/v1/venues/preview", org, map[string]any{"map": m})
	if pv2["total"] != pv["total"] || pv2["map"] != nil {
		t.Fatalf("the blueprint previews to the same seats: %v vs %v", pv2["total"], pv["total"])
	}
}
