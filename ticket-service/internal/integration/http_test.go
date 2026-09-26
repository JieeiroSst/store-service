package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/fx/fxtest"

	"github.com/JIeeiroSst/ticket-service/internal/bootstrap"
)

// fakeUserService accepts tokens of the form "user-<id>"; user 1000000001 is an admin.
func fakeUserService(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/validate", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Token string `json:"session_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		id, ok := strings.CutPrefix(in.Token, "user-")
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"valid": true, "user_id": id})
	})
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
		roles := []map[string]string{}
		if id == 1_000_000_001 {
			roles = append(roles, map[string]string{"name": "admin"})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"users": map[string]any{"id": id, "email": fmt.Sprintf("user%d@example.com", id), "roles": roles}})
	})
	s := httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func fakeWalletService(t *testing.T) *httptest.Server {
	var n atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/wallets/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"wallet_id": "w-" + r.PathValue("id"), "user_id": r.PathValue("id"), "balance": 1 << 40, "currency": "VND"})
	})
	mux.HandleFunc("POST /api/v1/transfers", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"transfer_id": fmt.Sprintf("tr-%d", n.Add(1))})
	})
	mux.HandleFunc("GET /api/v1/transfers/{id}", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "COMPLETED"})
	})
	mux.HandleFunc("POST /api/v1/transfers/{id}/reverse", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{}")) })
	s := httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

type api struct {
	t    *testing.T
	base string
	c    *http.Client
}

// do sends a request as the given user ("" = anonymous) and decodes the JSON answer.
func (a *api) do(method, path string, as int64, body any) (int, map[string]any) {
	a.t.Helper()
	var rd io.Reader
	if body != nil {
		if s, ok := body.(string); ok {
			rd = strings.NewReader(s)
		} else {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
	}
	req, _ := http.NewRequest(method, a.base+path, rd)
	if as != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer user-%d", as))
	}
	resp, err := a.c.Do(req)
	if err != nil {
		a.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func (a *api) must(want int, method, path string, as int64, body any) map[string]any {
	a.t.Helper()
	st, out := a.do(method, path, as, body)
	if st != want {
		a.t.Fatalf("%s %s = %d %v, want %d", method, path, st, out, want)
	}
	return out
}

func id(m map[string]any) int64 { return int64(m["id"].(float64)) }

// startApp runs the real fx module against a throwaway database, with fake user- and wallet-services.
func startApp(t *testing.T, extraEnv map[string]string) *api {
	t.Helper()
	d := newDatabase(t)
	users, wallets := fakeUserService(t), fakeWalletService(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	ln.Close()

	u, err := url.Parse(d.dsn)
	if err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		"PORT": port, "HostPostgres": u.Hostname(), "PortPostgres": u.Port(), "DatabasePostgres": d.name, "UserPostgres": "postgres",
		"UserServiceURL": users.URL, "WalletServiceURL": wallets.URL, "AdminRole": "admin", "UserRatePerSecond": "1000", "UserBurst": "1000",
		"MaxPendingPerUser": "3", "AuthCacheTTLSeconds": "30",
	}
	for k, v := range extraEnv {
		env[k] = v
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
	app := fxtest.New(t, bootstrap.Module)
	app.RequireStart()
	t.Cleanup(app.RequireStop)
	return &api{t: t, base: "http://127.0.0.1:" + port, c: &http.Client{Timeout: 20 * time.Second, Transport: &http.Transport{MaxIdleConnsPerHost: 200}}}
}

func TestHTTPEndToEnd(t *testing.T) {
	a := startApp(t, nil)
	const org, adm = int64(1_000_000_002), int64(1_000_000_001)

	// probes
	a.must(200, "GET", "/healthz", 0, nil)
	a.must(200, "GET", "/readyz", 0, nil)

	// input hygiene
	if st, _ := a.do("POST", "/api/v1/events", org, `{"title": "x", "surprise": 1}`); st != 400 {
		t.Fatalf("unknown field: %d", st)
	}
	if st, _ := a.do("POST", "/api/v1/events", org, `not json`); st != 400 {
		t.Fatalf("bad json: %d", st)
	}
	if st, _ := a.do("POST", "/api/v1/events", 0, `{}`); st != 401 {
		t.Fatalf("anonymous create: %d", st)
	}
	if st, _ := a.do("GET", "/api/v1/events/abc", 0, nil); st != 400 {
		t.Fatalf("bad id: %d", st)
	}
	if st, _ := a.do("GET", "/api/v1/events/999999", 0, nil); st != 404 {
		t.Fatalf("missing event: %d", st)
	}

	// organizer builds an event with two tickets on sale
	starts := time.Now().Add(96 * time.Hour).UTC().Truncate(time.Second)
	ev := a.must(201, "POST", "/api/v1/events", org, map[string]any{
		"title": "Hà Anh Tuấn Live", "category": "music", "city": "Hà Nội", "venue": "Mỹ Đình", "wallet_id": "org-wallet",
		"starts_at": starts, "ends_at": starts.Add(3 * time.Hour), "refund_cutoff_hours": 48})
	eid := id(ev)
	ep := fmt.Sprintf("/api/v1/events/%d", eid)
	tt := a.must(201, "POST", ep+"/ticket-types", org, map[string]any{"name": "Standing", "price": 800000, "total": 2, "max_per_order": 2})
	tid := id(tt)
	a.must(404, "POST", ep+"/ticket-types", 42, map[string]any{"name": "Hack", "price": 1, "total": 1}) // a stranger cannot even see a draft
	a.must(201, "POST", ep+"/promotions", org, map[string]any{"code": "HELLO", "kind": "percent", "value": 10, "max_uses": 1})
	a.must(200, "POST", ep+"/submit", org, nil)
	a.must(403, "POST", ep+"/approve", org, nil)
	a.must(200, "POST", ep+"/approve", adm, nil)
	a.must(200, "PUT", ep+"/featured", adm, map[string]any{"featured": true})

	// buyers browse
	list := a.must(200, "GET", "/api/v1/events?q=anh+tu%E1%BA%A5n&city=H%C3%A0+N%E1%BB%99i&category=music&featured=true&sort=popular", 0, nil)
	if items := list["items"].([]any); len(items) != 1 {
		t.Fatalf("search returned %v", list)
	}
	if got := a.must(200, "GET", "/api/v1/events?max_price=100", 0, nil)["items"].([]any); len(got) != 0 {
		t.Fatalf("price filter returned %v", got)
	}
	a.must(400, "GET", "/api/v1/events?category=cooking", 0, nil)
	a.must(200, "GET", "/api/v1/categories", 0, nil)
	view := a.must(200, "GET", ep, 0, nil)
	if _, leaked := view["wallet_id"]; leaked {
		t.Fatal("the organizer's wallet id is public")
	}

	// a crowd goes for the 2 tickets
	order := func(u int64) map[string]any {
		return map[string]any{"event_id": eid, "buyer_name": "Buyer", "request_id": "r-" + strconv.FormatInt(u, 10),
			"items": []map[string]any{{"ticket_type_id": tid, "quantity": 1}}}
	}
	a.must(401, "POST", "/api/v1/orders", 0, order(1)) // tickets left, no identity: unauthorized
	const crowd = 400
	var wg sync.WaitGroup
	var mu sync.Mutex
	winners := map[int64]int64{}
	codes := map[int]int{}
	for u := int64(1); u <= crowd; u++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			st, out := a.do("POST", "/api/v1/orders", u, order(u))
			// "busy" (429) means try again: a real client does, and so does this one
			for try := 0; st == 429 && try < 100; try++ {
				time.Sleep(20 * time.Millisecond)
				st, out = a.do("POST", "/api/v1/orders", u, order(u))
			}
			mu.Lock()
			defer mu.Unlock()
			codes[st]++
			if st == 201 {
				winners[u] = id(out)
			}
		}()
	}
	wg.Wait()
	if len(winners) != 2 || codes[201] != 2 || codes[201]+codes[409] != crowd {
		t.Fatalf("outcomes %v, winners %v: want 2 orders and everyone else 409", codes, winners)
	}
	if st, out := a.do("POST", "/api/v1/orders", 0, order(500)); st != 409 || !strings.Contains(fmt.Sprint(out["error"]), "sold out") {
		t.Fatalf("anonymous request for sold-out tickets: %d %v (want 409 without an identity lookup)", st, out)
	}
	// the public event page is cached for a moment (PublicCacheTTLSeconds), so "sold out" shows up within a few seconds
	soldOutShown := false
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline) && !soldOutShown; time.Sleep(200 * time.Millisecond) {
		soldOutShown = a.must(200, "GET", ep, 0, nil)["sold_out"] == true
	}
	if !soldOutShown {
		t.Fatal("the event page never said sold out")
	}

	// the winners pay, get QR codes, and are scanned at the gate
	var code string
	for u, oid := range winners {
		op := fmt.Sprintf("/api/v1/orders/%d", oid)
		a.must(404, "GET", op, 999, nil) // somebody else's order
		paid := a.must(200, "POST", op+"/pay", u, map[string]any{"method": "wallet"})
		if paid["status"] != "paid" || len(paid["tickets"].([]any)) != 1 {
			t.Fatalf("paid order: %v", paid)
		}
		code = paid["tickets"].([]any)[0].(map[string]any)["code"].(string)
		my := a.must(200, "GET", "/api/v1/my/tickets", u, nil)
		if len(my["items"].([]any)) != 1 {
			t.Fatalf("my tickets: %v", my)
		}
	}
	a.must(403, "POST", ep+"/check-in", 5, map[string]any{"code": code})
	a.must(200, "POST", ep+"/check-in", org, map[string]any{"code": code})
	if st, out := a.do("POST", ep+"/check-in", org, map[string]any{"code": code}); st != 409 || out["ticket"] == nil {
		t.Fatalf("second scan: %d %v", st, out)
	}
	rep := a.must(200, "GET", ep+"/report", org, nil)
	if rep["tickets_sold"].(float64) != 2 || rep["checked_in"].(float64) != 1 || rep["revenue"].(float64) != 1_600_000 {
		t.Fatalf("report: %v", rep)
	}
	att := a.must(200, "GET", ep+"/attendees", org, nil)
	if len(att["items"].([]any)) != 2 {
		t.Fatalf("attendees: %v", att)
	}

	// a buyer gets their notification and the event goes on a wishlist
	for u := range winners {
		n := a.must(200, "GET", "/api/v1/notifications/unread-count", u, nil)
		if n["unread"].(float64) != 1 {
			t.Fatalf("unread: %v", n)
		}
		a.must(200, "POST", "/api/v1/notifications/read-all", u, nil)
		a.must(204, "PUT", "/api/v1/wishlist/"+strconv.FormatInt(eid, 10), u, nil)
		if got := a.must(200, "GET", "/api/v1/wishlist", u, nil); len(got["items"].([]any)) != 1 {
			t.Fatalf("wishlist: %v", got)
		}
		break
	}
}
