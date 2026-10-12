package infrastructure

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary/userservice/userservicetest"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

func startWithUserService(t *testing.T, jidField string) (testServer, *userservicetest.Fake) {
	t.Helper()
	fake := userservicetest.NewFake("internal-secret",
		userservicetest.FakeUser{ID: 7, Username: "alice", Password: "alicepw"},
		userservicetest.FakeUser{ID: 8, Username: "bob", Password: "bobpw"},
	)
	us := httptest.NewServer(fake)
	t.Cleanup(us.Close)
	s := startServer(t, chatConfig+fmt.Sprintf(`
auth_method: user_service
user_service:
  url: %s
  internal_token: internal-secret
  jid_field: %s
`, us.URL, jidField))
	return s, fake
}

func plainAuth(c *client, user, pass string) *domain.Element {
	c.send(fmt.Sprintf("<auth xmlns='%s' mechanism='PLAIN'>%s</auth>", domain.NSSASL,
		base64.StdEncoding.EncodeToString([]byte("\x00"+user+"\x00"+pass))))
	return c.next()
}

func TestUserServiceLogin(t *testing.T) {
	s, fake := startWithUserService(t, "username")

	t.Log("only PLAIN is offered: passwords live in user_service")
	c := dial(t, s.c2s)
	feats := c.open()
	var mechs []string
	for _, m := range feats.Child("mechanisms", domain.NSSASL).ChildrenNamed("mechanism", "") {
		mechs = append(mechs, m.Text())
	}
	if strings.Join(mechs, ",") != "PLAIN" {
		t.Fatalf("mechanisms = %v", mechs)
	}

	t.Log("in-band registration is refused")
	res := c.iq("reg", "set", "", "<query xmlns='jabber:iq:register'><username>mallory</username><password>x</password></query>")
	if res.Child("error", "").Child("not-allowed", domain.NSStanzas) == nil {
		t.Fatalf("register: %s", res)
	}

	t.Log("someone else's token and garbage are rejected")
	aliceTok, _ := fake.Login("alice", "alicepw")
	if r := plainAuth(c, "bob", aliceTok); r.Name != "failure" {
		t.Fatalf("alice token as bob: %s", r)
	}
	if r := plainAuth(c, "alice", "aaa.bbb.ccc"); r.Name != "failure" {
		t.Fatalf("garbage token: %s", r)
	}

	t.Log("mobile: login to user_service, connect with the session token")
	alice := login(t, s.c2s, "alice", aliceTok, "android", "PLAIN")
	if alice.jid != "alice@localhost/android" {
		t.Fatalf("jid = %s", alice.jid)
	}

	t.Log("desktop: connect with the user_service password")
	bob := login(t, s.c2s, "bob", "bobpw", "desktop", "PLAIN")
	if fake.Logouts.Load() != 1 {
		t.Fatalf("password login left a session open in user_service (logouts=%d)", fake.Logouts.Load())
	}

	t.Log("chat between the two")
	api(t, s.http, "add_rosteritem", map[string]any{"localuser": "alice", "localhost": "localhost", "user": "bob", "host": "localhost", "nick": "Bob", "groups": []string{}, "subs": "both"})
	api(t, s.http, "add_rosteritem", map[string]any{"localuser": "bob", "localhost": "localhost", "user": "alice", "host": "localhost", "nick": "Alice", "groups": []string{}, "subs": "both"})
	alice.send("<presence/>")
	bob.send("<presence/>")
	bob.expect("alice online", isPresence("alice@localhost/android", "available"))
	alice.send("<message to='bob@localhost' type='chat'><body>hi from user_service</body></message>")
	bob.expect("message", isChat("alice@localhost/android", "hi from user_service"))

	t.Log("offline storage needs user_service to confirm the recipient exists")
	bob.send("<presence type='unavailable'/></stream:stream>")
	alice.expect("bob offline", isPresence("bob@localhost/desktop", "unavailable"))
	alice.send("<message to='bob@localhost' type='chat' id='o1'><body>later</body></message>")
	alice.send("<message to='ghost@localhost' type='chat' id='g1'><body>?</body></message>")
	alice.expect("ghost bounced", func(e *domain.Element) bool {
		return e.Attr("id") == "g1" && e.Child("error", "").Child("service-unavailable", domain.NSStanzas) != nil
	})
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, body := api(t, s.http, "get_offline_count", map[string]any{"user": "bob", "host": "localhost"}); body == "1" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("message for bob was not stored offline")
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Log("account management commands point to user_service")
	if code, body := api(t, s.http, "register", map[string]any{"user": "x", "host": "localhost", "password": "y"}); code != 400 || !strings.Contains(body, "user_service") {
		t.Fatalf("register = %d %s", code, body)
	}
	if _, body := api(t, s.http, "check_account", map[string]any{"user": "bob", "host": "localhost"}); body != "0" {
		t.Fatalf("check_account bob = %s", body)
	}
	if _, body := api(t, s.http, "check_account", map[string]any{"user": "ghost", "host": "localhost"}); body != "1" {
		t.Fatalf("check_account ghost = %s", body)
	}

	t.Log("revoked token cannot reconnect, a fresh one can")
	fake.Revoke(aliceTok)
	c2 := dial(t, s.c2s)
	c2.open()
	if r := plainAuth(c2, "alice", aliceTok); r.Name != "failure" {
		t.Fatalf("revoked token accepted: %s", r)
	}
	fresh, _ := fake.Login("alice", "alicepw")
	login(t, s.c2s, "alice", fresh, "android-2", "PLAIN")
}

func TestUserServiceUserIDMode(t *testing.T) {
	s, fake := startWithUserService(t, "user_id")
	tok, _ := fake.Login("bob", "bobpw")
	c := login(t, s.c2s, "8", tok, "phone", "PLAIN")
	if c.jid != "8@localhost/phone" {
		t.Fatalf("jid = %s", c.jid)
	}
	bad := dial(t, s.c2s)
	bad.open()
	if r := plainAuth(bad, "bob", tok); r.Name != "failure" {
		t.Fatalf("username must not work in user_id mode: %s", r)
	}
	if _, body := api(t, s.http, "check_account", map[string]any{"user": "7", "host": "localhost"}); body != "0" {
		t.Fatalf("check_account 7 = %s", body)
	}
}

func basicAuthCall(t *testing.T, port int, user, pass string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/api/connected_users_number", port), strings.NewReader("{}"))
	req.SetBasicAuth(user, pass)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestUserServiceBootstrapAdmin(t *testing.T) {
	fake := userservicetest.NewFake("")
	fake.Down.Store(2)
	us := httptest.NewServer(fake)
	t.Cleanup(us.Close)
	cfg := chatConfig + fmt.Sprintf(`
acl:
  admin:
    user: admin@localhost
auth_method: user_service
user_service:
  url: %s
  admins:
    - username: admin
      password: Admin12345
      email: ejabberd.admin@store.com
    - username: weakling
      password: weak
      email: ejabberd.weak@store.com
`, us.URL)
	s := startServer(t, cfg)

	t.Log("user_service is down at first: the admin is created after a retry")
	deadline := time.Now().Add(10 * time.Second)
	for !fake.HasUser("admin") {
		if time.Now().After(deadline) {
			t.Fatal("admin was not created in user_service")
		}
		time.Sleep(100 * time.Millisecond)
	}
	if fake.SignUps.Load() < 2 {
		t.Fatalf("sign-up attempts = %d: the outage was not retried", fake.SignUps.Load())
	}
	if fake.HasUser("weakling") {
		t.Fatal("account with a weak password must be rejected by user_service")
	}

	t.Log("the admin can call the HTTP API with Basic auth")
	if code := basicAuthCall(t, s.http, "admin@localhost", "Admin12345"); code != http.StatusOK {
		t.Fatalf("admin basic auth = %d", code)
	}
	if code := basicAuthCall(t, s.http, "admin@localhost", "wrong"); code != http.StatusUnauthorized {
		t.Fatalf("wrong password = %d", code)
	}

	t.Log("restart: the existing admin is left alone, no duplicate")
	signUps := fake.SignUps.Load()
	s2 := startServer(t, cfg)
	deadline = time.Now().Add(5 * time.Second)
	for fake.SignUps.Load() < signUps+1 {
		if time.Now().After(deadline) {
			t.Fatal("bootstrap did not run on restart")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if code := basicAuthCall(t, s2.http, "admin@localhost", "Admin12345"); code != http.StatusOK {
		t.Fatalf("admin basic auth after restart = %d", code)
	}
}
