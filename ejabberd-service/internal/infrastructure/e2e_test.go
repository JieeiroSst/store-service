package infrastructure

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

type testServer struct {
	c2s, tlsPort, http int
}

func startServer(t *testing.T, extra string) testServer {
	t.Helper()
	s := testServer{c2s: freePort(t), tlsPort: freePort(t), http: freePort(t)}
	raw := fmt.Sprintf(`
hosts: [localhost]
loglevel: error
registration_timeout: infinity
listen:
  - port: %d
    ip: "127.0.0.1"
    module: ejabberd_c2s
    max_stanza_size: 65536
    shaper: c2s_shaper
    access: c2s
  - port: %d
    ip: "127.0.0.1"
    module: ejabberd_c2s
    starttls_required: true
  - port: %d
    ip: "127.0.0.1"
    module: ejabberd_http
    request_handlers:
      /api: mod_http_api
      /upload: mod_http_upload
acl:
  local:
    user_regexp: ""
  loopback:
    ip: 127.0.0.0/8
  blocked:
    user: spammer
access_rules:
  local:
    allow: local
  c2s:
    deny: blocked
    allow: all
  trusted_network:
    allow: loopback
api_permissions:
  "loopback":
    who:
      ip: 127.0.0.1/8
    what: "*"
shaper:
  normal:
    rate: 3000
    burst_size: 20000
shaper_rules:
  max_user_sessions: 10
  max_user_offline_messages: 2
  c2s_shaper: normal
modules:
  mod_roster:
    versioning: true
  mod_offline:
    access_max_user_messages: max_user_offline_messages
  mod_ping: {}
  mod_disco: {}
  mod_version:
    show_os: false
  mod_register:
    ip_access: trusted_network
  mod_http_api: {}
  mod_last: {}
%s`, s.c2s, s.tlsPort, s.http, extra)
	raw = strings.ReplaceAll(raw, "@HTTP_PORT@", strconv.Itoa(s.http))
	cfg, err := config.Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	app := fxtest.New(t, fx.Supply(cfg), CoreModule, fx.NopLogger)
	app.RequireStart()
	t.Cleanup(app.RequireStop)
	return s
}

type readResult struct {
	el  *domain.Element
	err error
}

type client struct {
	t    *testing.T
	conn net.Conn
	dec  *xml.Decoder
	jid  string
	pump chan readResult
}

func (c *client) startPump() {
	c.pump = make(chan readResult, 4096)
	_ = c.conn.SetReadDeadline(time.Time{})
	go func() {
		for {
			el, err := c.readElement()
			c.pump <- readResult{el, err}
			if err != nil {
				return
			}
		}
	}()
}

func (c *client) readElement() (*domain.Element, error) {
	for {
		tok, err := c.dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			return domain.ReadElement(c.dec, t, domain.NSClient)
		case xml.EndElement:
			return nil, io.EOF
		}
	}
}

func dial(t *testing.T, port int) *client {
	t.Helper()
	conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	c := &client{t: t, conn: conn}
	t.Cleanup(func() { conn.Close() })
	return c
}

func (c *client) send(s string) {
	c.t.Helper()
	if _, err := io.WriteString(c.conn, s); err != nil {
		c.t.Fatalf("send: %v", err)
	}
}

func (c *client) open() *domain.Element {
	c.t.Helper()
	c.dec = xml.NewDecoder(c.conn)
	c.send(`<?xml version='1.0'?><stream:stream to='localhost' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams' version='1.0'>`)
	_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		tok, err := c.dec.Token()
		if err != nil {
			c.t.Fatalf("open: %v", err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "stream" {
			break
		}
	}
	return c.next()
}

func (c *client) next() *domain.Element {
	c.t.Helper()
	el, err := c.tryNext(5 * time.Second)
	if err != nil {
		c.t.Fatalf("%s: read: %v", c.jid, err)
	}
	return el
}

func (c *client) tryNext(d time.Duration) (*domain.Element, error) {
	if c.pump == nil {
		_ = c.conn.SetReadDeadline(time.Now().Add(d))
		return c.readElement()
	}
	if d <= 0 {
		d = time.Millisecond
	}
	select {
	case r := <-c.pump:
		if r.err != nil {
			c.pump <- r
		}
		return r.el, r.err
	case <-time.After(d):
		return nil, errors.New("timeout")
	}
}

func (c *client) expect(desc string, pred func(*domain.Element) bool) *domain.Element {
	c.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var seen []string
	for time.Now().Before(deadline) {
		el, err := c.tryNext(time.Until(deadline))
		if err != nil {
			c.t.Fatalf("%s: waiting for %s: %v (seen: %s)", c.jid, desc, err, strings.Join(seen, " | "))
		}
		if pred(el) {
			return el
		}
		seen = append(seen, el.String())
	}
	c.t.Fatalf("%s: timeout waiting for %s (seen: %s)", c.jid, desc, strings.Join(seen, " | "))
	return nil
}

func (c *client) silent(d time.Duration, pred func(*domain.Element) bool) {
	c.t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		el, err := c.tryNext(time.Until(deadline))
		if err != nil {
			return
		}
		if pred(el) {
			c.t.Fatalf("%s: unexpected stanza %s", c.jid, el)
		}
	}
}

func (c *client) iq(id, typ, to, payload string) *domain.Element {
	c.t.Helper()
	toAttr := ""
	if to != "" {
		toAttr = " to='" + to + "'"
	}
	c.send(fmt.Sprintf("<iq type='%s' id='%s'%s>%s</iq>", typ, id, toAttr, payload))
	return c.expect("iq "+id, func(e *domain.Element) bool { return e.Name == "iq" && e.Attr("id") == id })
}

func register(t *testing.T, port int, user, pass string) *domain.Element {
	t.Helper()
	c := dial(t, port)
	feats := c.open()
	if feats.Child("register", domain.NSFeatureIQReg) == nil {
		t.Fatalf("register feature missing: %s", feats)
	}
	res := c.iq("reg1", "set", "", fmt.Sprintf("<query xmlns='jabber:iq:register'><username>%s</username><password>%s</password></query>", user, pass))
	c.conn.Close()
	return res
}

func login(t *testing.T, port int, user, pass, resource, mech string) *client {
	t.Helper()
	c := dial(t, port)
	feats := c.open()
	mechs := feats.Child("mechanisms", domain.NSSASL)
	if mechs == nil {
		t.Fatalf("no mechanisms: %s", feats)
	}
	switch mech {
	case "PLAIN":
		c.send(fmt.Sprintf("<auth xmlns='%s' mechanism='PLAIN'>%s</auth>", domain.NSSASL,
			base64.StdEncoding.EncodeToString([]byte("\x00"+user+"\x00"+pass))))
	case "SCRAM-SHA-256":
		c.scram(domain.ScramSHA256, mech, user, pass)
	}
	res := c.next()
	if res.Name != "success" {
		t.Fatalf("auth failed: %s", res)
	}
	feats = c.open()
	if feats.Child("bind", domain.NSBind) == nil || feats.Child("ver", domain.NSRosterVer) == nil {
		t.Fatalf("post-auth features: %s", feats)
	}
	bind := c.iq("bind1", "set", "", fmt.Sprintf("<bind xmlns='%s'><resource>%s</resource></bind>", domain.NSBind, resource))
	if bind.Attr("type") != "result" {
		t.Fatalf("bind: %s", bind)
	}
	c.jid = bind.Child("bind", domain.NSBind).ChildText("jid", "")
	c.startPump()
	return c
}

func (c *client) scram(h domain.ScramHash, mech, user, pass string) {
	nonce := make([]byte, 12)
	_, _ = rand.Read(nonce)
	cfb := "n=" + user + ",r=" + base64.RawStdEncoding.EncodeToString(nonce)
	c.send(fmt.Sprintf("<auth xmlns='%s' mechanism='%s'>%s</auth>", domain.NSSASL, mech, base64.StdEncoding.EncodeToString([]byte("n,,"+cfb))))
	ch := c.next()
	if ch.Name != "challenge" {
		c.t.Fatalf("expected challenge: %s", ch)
	}
	sf, _ := base64.StdEncoding.DecodeString(ch.Text())
	attrs := map[string]string{}
	for _, kv := range strings.Split(string(sf), ",") {
		attrs[kv[:1]] = kv[2:]
	}
	salt, _ := base64.StdEncoding.DecodeString(attrs["s"])
	iter, _ := strconv.Atoi(attrs["i"])
	salted := domain.SaltedPassword(h, pass, salt, iter)
	ck := domain.ClientKey(h, salted)
	final := "c=biws,r=" + attrs["r"]
	am := cfb + "," + string(sf) + "," + final
	sig := domain.HMAC(h, domain.StoredKey(h, ck), []byte(am))
	for i := range ck {
		ck[i] ^= sig[i]
	}
	c.send(fmt.Sprintf("<response xmlns='%s'>%s</response>", domain.NSSASL,
		base64.StdEncoding.EncodeToString([]byte(final+",p="+base64.StdEncoding.EncodeToString(ck)))))
}

func isPresence(from, typ string) func(*domain.Element) bool {
	return func(e *domain.Element) bool {
		return e.Name == "presence" && strings.HasPrefix(e.Attr("from"), from) && domain.StanzaType(e) == typ
	}
}

func isRosterPush(jid, sub string) func(*domain.Element) bool {
	return func(e *domain.Element) bool {
		if e.Name != "iq" || e.Attr("type") != "set" {
			return false
		}
		q := e.Child("query", domain.NSRoster)
		if q == nil {
			return false
		}
		it := q.Child("item", "")
		return it != nil && it.Attr("jid") == jid && it.Attr("subscription") == sub
	}
}

func api(t *testing.T, port int, cmd string, body any) (int, string) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/api/%s", port, cmd), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, strings.TrimSpace(string(out))
}

func TestEndToEnd(t *testing.T) {
	s := startServer(t, "")

	if res := register(t, s.c2s, "alice", "alicepw"); res.Attr("type") != "result" {
		t.Fatalf("register alice: %s", res)
	}
	if res := register(t, s.c2s, "bob", "bobpw"); res.Attr("type") != "result" {
		t.Fatalf("register bob: %s", res)
	}
	if res := register(t, s.c2s, "alice", "x"); res.Child("error", "").Child("conflict", domain.NSStanzas) == nil {
		t.Fatalf("duplicate register must conflict: %s", res)
	}

	bad := dial(t, s.c2s)
	bad.open()
	bad.send(fmt.Sprintf("<auth xmlns='%s' mechanism='PLAIN'>%s</auth>", domain.NSSASL, base64.StdEncoding.EncodeToString([]byte("\x00alice\x00nope"))))
	if f := bad.next(); f.Name != "failure" || f.Child("not-authorized", "") == nil {
		t.Fatalf("wrong password: %s", f)
	}

	alice := login(t, s.c2s, "alice", "alicepw", "home", "PLAIN")
	bob := login(t, s.c2s, "bob", "bobpw", "work", "SCRAM-SHA-256")

	ver := alice.iq("v1", "get", "localhost", "<query xmlns='jabber:iq:version'/>")
	if ver.Child("query", domain.NSVersion).ChildText("name", "") != "ejabberd" {
		t.Fatalf("version: %s", ver)
	}
	if p := alice.iq("p1", "get", "localhost", "<ping xmlns='urn:xmpp:ping'/>"); p.Attr("type") != "result" {
		t.Fatalf("ping: %s", p)
	}
	if d := alice.iq("d1", "get", "localhost", "<query xmlns='http://jabber.org/protocol/disco#info'/>"); d.Child("query", domain.NSDiscoInfo) == nil {
		t.Fatalf("disco: %s", d)
	}
	if u := alice.iq("u1", "get", "localhost", "<query xmlns='urn:unknown'/>"); u.Child("error", "").Child("service-unavailable", domain.NSStanzas) == nil {
		t.Fatalf("unknown iq must be service-unavailable: %s", u)
	}

	alice.send("<presence/>")
	alice.expect("own presence", isPresence("alice@localhost/home", "available"))
	bob.send("<presence><priority>5</priority></presence>")
	bob.expect("own presence", isPresence("bob@localhost/work", "available"))

	alice.send("<presence to='bob@localhost' type='subscribe'/>")
	alice.expect("push ask", func(e *domain.Element) bool {
		return isRosterPush("bob@localhost", "none")(e) && e.Child("query", domain.NSRoster).Child("item", "").Attr("ask") == "subscribe"
	})
	bob.expect("subscribe request", func(e *domain.Element) bool {
		return isPresence("alice@localhost", "subscribe")(e) && e.Attr("from") == "alice@localhost"
	})

	bob.send("<presence to='alice@localhost' type='subscribed'/>")
	bob.expect("bob push from", isRosterPush("alice@localhost", "from"))
	alice.expect("alice push to", isRosterPush("bob@localhost", "to"))
	alice.expect("bob presence", isPresence("bob@localhost/work", "available"))

	roster := alice.iq("r1", "get", "", "<query xmlns='jabber:iq:roster' ver=''/>")
	q := roster.Child("query", domain.NSRoster)
	if q == nil || q.Attr("ver") == "" || len(q.ChildrenNamed("item", "")) != 1 {
		t.Fatalf("roster get: %s", roster)
	}
	again := alice.iq("r2", "get", "", fmt.Sprintf("<query xmlns='jabber:iq:roster' ver='%s'/>", q.Attr("ver")))
	if again.Attr("type") != "result" || len(again.Elements()) != 0 {
		t.Fatalf("roster versioning must return empty result: %s", again)
	}
	if r := alice.iq("r3", "set", "", "<query xmlns='jabber:iq:roster'><item jid='x@localhost' ask='subscribe'/></query>"); r.Child("error", "").Child("bad-request", domain.NSStanzas) == nil {
		t.Fatalf("ask attribute must be rejected: %s", r)
	}

	alice.send("<message to='bob@localhost' type='chat' id='m1'><body>hello bob</body></message>")
	bob.expect("message", func(e *domain.Element) bool {
		return e.Name == "message" && e.ChildText("body", "") == "hello bob" && e.Attr("from") == "alice@localhost/home"
	})

	if code, body := api(t, s.http, "connected_users_number", map[string]any{}); code != 200 || body != "2" {
		t.Fatalf("connected_users_number = %d %s", code, body)
	}

	bob.send("</stream:stream>")
	alice.expect("bob unavailable", isPresence("bob@localhost/work", "unavailable"))

	alice.send("<message to='bob@localhost' type='chat' id='m2'><body>are you there?</body></message>")
	alice.send("<message to='bob@localhost' type='chat' id='m3'><body>second</body></message>")
	alice.send("<message to='bob@localhost' type='chat' id='m4'><body>third</body></message>")
	alice.expect("offline queue full", func(e *domain.Element) bool {
		return e.Name == "message" && e.Attr("id") == "m4" && e.Attr("type") == "error" &&
			e.Child("error", "").Child("resource-constraint", domain.NSStanzas) != nil
	})
	if code, body := api(t, s.http, "get_offline_count", map[string]any{"user": "bob", "host": "localhost"}); code != 200 || body != "2" {
		t.Fatalf("get_offline_count = %d %s", code, body)
	}

	bob2 := login(t, s.c2s, "bob", "bobpw", "work", "PLAIN")
	bob2.send("<presence/>")
	msg := bob2.expect("offline message", func(e *domain.Element) bool {
		return e.Name == "message" && e.ChildText("body", "") == "are you there?"
	})
	if d := msg.Child("delay", domain.NSDelay); d == nil || d.Text() != "Offline Storage" || d.Attr("from") != "localhost" {
		t.Fatalf("offline delay missing: %s", msg)
	}
	alice.expect("bob back online", isPresence("bob@localhost/work", "available"))

	bob3 := login(t, s.c2s, "bob", "bobpw", "work", "PLAIN")
	bob2.expect("replaced", func(e *domain.Element) bool {
		return e.Name == "stream:error" || (e.Name == "error" && e.Child("conflict", domain.NSStreams) != nil)
	})
	_ = bob3

	alice.send("<message to='nobody@localhost' type='chat' id='m5'><body>?</body></message>")
	alice.expect("unknown user bounce", func(e *domain.Element) bool {
		return e.Attr("id") == "m5" && e.Attr("type") == "error" && e.Child("error", "").Child("service-unavailable", domain.NSStanzas) != nil
	})

	alice.send("<message to='bob@example.org' type='chat' id='m6'><body>?</body></message>")
	alice.expect("remote server", func(e *domain.Element) bool {
		return e.Attr("id") == "m6" && e.Child("error", "").Child("remote-server-not-found", domain.NSStanzas) != nil
	})
}

func TestHTTPAPI(t *testing.T) {
	s := startServer(t, "")
	code, body := api(t, s.http, "register", map[string]any{"user": "carol", "host": "localhost", "password": "pw"})
	if code != 200 || body != `"User carol@localhost successfully registered"` {
		t.Fatalf("register = %d %s", code, body)
	}
	code, body = api(t, s.http, "register", map[string]any{"user": "carol", "host": "localhost", "password": "pw"})
	if code != 409 || !strings.Contains(body, "already registered") {
		t.Fatalf("duplicate register = %d %s", code, body)
	}
	if code, body = api(t, s.http, "check_account", map[string]any{"user": "carol", "host": "localhost"}); body != "0" {
		t.Fatalf("check_account = %d %s", code, body)
	}
	if code, body = api(t, s.http, "check_account", map[string]any{"user": "dave", "host": "localhost"}); body != "1" {
		t.Fatalf("check_account missing = %d %s", code, body)
	}
	api(t, s.http, "register", map[string]any{"user": "dave", "host": "localhost", "password": "pw"})
	if code, body = api(t, s.http, "add_rosteritem", map[string]any{
		"localuser": "carol", "localhost": "localhost", "user": "dave", "host": "localhost",
		"nick": "Dave", "groups": []string{"Friends"}, "subs": "both",
	}); body != "0" {
		t.Fatalf("add_rosteritem = %d %s", code, body)
	}
	code, body = api(t, s.http, "get_roster", map[string]any{"user": "carol", "host": "localhost"})
	if code != 200 || !strings.Contains(body, `"jid":"dave@localhost"`) || !strings.Contains(body, `"subscription":"both"`) {
		t.Fatalf("get_roster = %d %s", code, body)
	}
	if code, body = api(t, s.http, "registered_users", map[string]any{"host": "localhost"}); body != `["carol","dave"]` {
		t.Fatalf("registered_users = %d %s", code, body)
	}
	if code, _ = api(t, s.http, "nope", map[string]any{}); code != 404 {
		t.Fatalf("unknown command = %d", code)
	}
	if code, body = api(t, s.http, "unregister", map[string]any{"user": "dave", "host": "localhost"}); code != 200 {
		t.Fatalf("unregister = %d %s", code, body)
	}
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", s.http))
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("health = %v %v", resp, err)
	}
	resp.Body.Close()
}

func TestStartTLSRequired(t *testing.T) {
	s := startServer(t, "")
	register(t, s.c2s, "erin", "pw")
	c := dial(t, s.tlsPort)
	feats := c.open()
	st := feats.Child("starttls", domain.NSTLS)
	if st == nil || st.Child("required", "") == nil || feats.Child("mechanisms", domain.NSSASL) != nil {
		t.Fatalf("features before tls: %s", feats)
	}
	c.send(fmt.Sprintf("<starttls xmlns='%s'/>", domain.NSTLS))
	if p := c.next(); p.Name != "proceed" {
		t.Fatalf("expected proceed: %s", p)
	}
	tc := tls.Client(c.conn, &tls.Config{InsecureSkipVerify: true, ServerName: "localhost"})
	if err := tc.Handshake(); err != nil {
		t.Fatal(err)
	}
	c.conn = tc
	feats = c.open()
	if feats.Child("mechanisms", domain.NSSASL) == nil || feats.Child("starttls", domain.NSTLS) != nil {
		t.Fatalf("features after tls: %s", feats)
	}
	c.send(fmt.Sprintf("<auth xmlns='%s' mechanism='PLAIN'>%s</auth>", domain.NSSASL, base64.StdEncoding.EncodeToString([]byte("\x00erin\x00pw"))))
	if r := c.next(); r.Name != "success" {
		t.Fatalf("auth over tls: %s", r)
	}
}

func TestAccessDeniedBind(t *testing.T) {
	s := startServer(t, "")
	_, _ = api(t, s.http, "register", map[string]any{"user": "spammer", "host": "localhost", "password": "pw"})
	c := dial(t, s.c2s)
	c.open()
	c.send(fmt.Sprintf("<auth xmlns='%s' mechanism='PLAIN'>%s</auth>", domain.NSSASL, base64.StdEncoding.EncodeToString([]byte("\x00spammer\x00pw"))))
	if r := c.next(); r.Name != "success" {
		t.Fatalf("auth: %s", r)
	}
	c.open()
	res := c.iq("b", "set", "", fmt.Sprintf("<bind xmlns='%s'/>", domain.NSBind))
	if res.Child("error", "").Child("not-allowed", domain.NSStanzas) == nil {
		t.Fatalf("blocked user bind: %s", res)
	}

}

func TestServerPingKillsDeadSession(t *testing.T) {
	s2 := startServer(t, `
  mod_ping:
    send_pings: true
    ping_interval: 1
    ping_ack_timeout: 1
    timeout_action: kill
`)
	register(t, s2.c2s, "frank", "pw")
	register(t, s2.c2s, "gina", "pw")

	alive := login(t, s2.c2s, "gina", "pw", "phone", "PLAIN")
	dead := login(t, s2.c2s, "frank", "pw", "phone", "PLAIN")
	alive.send("<presence/>")
	dead.send("<presence/>")

	ping := dead.expect("server ping", func(e *domain.Element) bool {
		return e.Name == "iq" && e.Attr("type") == "get" && e.Child("ping", domain.NSPing) != nil
	})
	if ping.Attr("from") != "localhost" {
		t.Fatalf("ping from = %s", ping.Attr("from"))
	}

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		el, err := alive.tryNext(time.Until(deadline))
		if err != nil {
			t.Fatalf("alive session dropped: %v", err)
		}
		if el.Name == "iq" && el.Child("ping", domain.NSPing) != nil {
			alive.send(fmt.Sprintf("<iq type='result' id='%s' to='localhost'/>", el.Attr("id")))
		}
		if el.Name == "error" && el.Child("conflict", domain.NSStreams) != nil {
			t.Fatal("unexpected conflict")
		}
		code, body := api(t, s2.http, "user_resources", map[string]any{"user": "frank", "host": "localhost"})
		if code == 200 && body == "[]" {
			code, body = api(t, s2.http, "user_resources", map[string]any{"user": "gina", "host": "localhost"})
			if body != `["phone"]` {
				t.Fatalf("responsive session was killed: %d %s", code, body)
			}
			return
		}
	}
	t.Fatal("unresponsive session was not killed by server ping")
}
