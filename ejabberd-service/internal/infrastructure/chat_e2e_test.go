package infrastructure

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

const chatConfig = `
shaper_rules:
  max_user_sessions: 10
  max_user_offline_messages: 1000
  c2s_shaper: normal
shaper:
  normal:
    rate: 10000000
`

func isChat(from, body string) func(*domain.Element) bool {
	return func(e *domain.Element) bool {
		return e.Name == "message" && strings.HasPrefix(e.Attr("from"), from) && e.ChildText("body", "") == body
	}
}

func (c *client) messages(n int, timeout time.Duration, keep func(*domain.Element) bool) []*domain.Element {
	c.t.Helper()
	deadline := time.Now().Add(timeout)
	var out []*domain.Element
	for len(out) < n {
		el, err := c.tryNext(time.Until(deadline))
		if err != nil {
			c.t.Fatalf("%s: got %d/%d messages: %v", c.jid, len(out), n, err)
		}
		if el.Name == "message" && keep(el) {
			out = append(out, el)
		}
	}
	return out
}

func (c *client) expectAll(timeout time.Duration, want map[string]func(*domain.Element) bool) {
	c.t.Helper()
	deadline := time.Now().Add(timeout)
	var seen []string
	for len(want) > 0 {
		el, err := c.tryNext(time.Until(deadline))
		if err != nil {
			var missing []string
			for k := range want {
				missing = append(missing, k)
			}
			c.t.Fatalf("%s: missing %v: %v (seen: %s)", c.jid, missing, err, strings.Join(seen, " | "))
		}
		matched := false
		for k, pred := range want {
			if pred(el) {
				delete(want, k)
				matched = true
				break
			}
		}
		if !matched {
			seen = append(seen, el.String())
		}
	}
}

func provision(t *testing.T, s testServer, users ...string) {
	t.Helper()
	for _, u := range users {
		if code, body := api(t, s.http, "register", map[string]any{"user": u, "host": "localhost", "password": u + "pw"}); code != 200 {
			t.Fatalf("register %s: %d %s", u, code, body)
		}
	}
	for _, a := range users {
		for _, b := range users {
			if a == b {
				continue
			}
			code, body := api(t, s.http, "add_rosteritem", map[string]any{
				"localuser": a, "localhost": "localhost", "user": b, "host": "localhost",
				"nick": strings.ToUpper(b[:1]) + b[1:], "groups": []string{"Friends"}, "subs": "both",
			})
			if code != 200 || body != "0" {
				t.Fatalf("add_rosteritem %s->%s: %d %s", a, b, code, body)
			}
		}
	}
}

func TestRealtimeChat(t *testing.T) {
	s := startServer(t, chatConfig)
	provision(t, s, "alice", "bob")

	t.Log("1. login: alice on phone and tablet, bob on phone (SCRAM-SHA-256)")
	alicePhone := login(t, s.c2s, "alice", "alicepw", "android-phone", "SCRAM-SHA-256")
	aliceTablet := login(t, s.c2s, "alice", "alicepw", "ios-tablet", "SCRAM-SHA-256")
	bob := login(t, s.c2s, "bob", "bobpw", "android-1", "SCRAM-SHA-256")

	t.Log("2. roster provisioned by the backend is visible to the app")
	r := bob.iq("roster", "get", "", "<query xmlns='jabber:iq:roster' ver=''/>")
	item := r.Child("query", domain.NSRoster).Child("item", "")
	if item == nil || item.Attr("jid") != "alice@localhost" || item.Attr("subscription") != "both" ||
		item.Attr("name") != "Alice" || item.ChildText("group", "") != "Friends" {
		t.Fatalf("bob roster: %s", r)
	}

	t.Log("3. initial presence: every party sees the others come online")
	alicePhone.send("<presence><priority>10</priority></presence>")
	aliceTablet.send("<presence><priority>0</priority></presence>")
	bob.send("<presence><priority>5</priority><status>Đang rảnh</status></presence>")
	bob.expectAll(5*time.Second, map[string]func(*domain.Element) bool{
		"alice phone online":  isPresence("alice@localhost/android-phone", "available"),
		"alice tablet online": isPresence("alice@localhost/ios-tablet", "available"),
	})
	p := alicePhone.expect("bob online", isPresence("bob@localhost/android-1", "available"))
	if p.ChildText("status", "") != "Đang rảnh" {
		t.Fatalf("status not propagated: %s", p)
	}
	aliceTablet.expect("bob online", isPresence("bob@localhost/android-1", "available"))

	t.Log("4. typing indicator and message alice -> bob, delivered in real time")
	start := time.Now()
	alicePhone.send("<message to='bob@localhost' type='chat' id='t1'><composing xmlns='http://jabber.org/protocol/chatstates'/></message>")
	bob.expect("composing", func(e *domain.Element) bool {
		return e.Name == "message" && e.Child("composing", domain.NSChatStates) != nil
	})
	alicePhone.send("<message to='bob@localhost' type='chat' id='m1'><body>Chào Bob 👋</body><active xmlns='http://jabber.org/protocol/chatstates'/><request xmlns='urn:xmpp:receipts'/></message>")
	m := bob.expect("message m1", isChat("alice@localhost/android-phone", "Chào Bob 👋"))
	t.Logf("   delivered in %v", time.Since(start))
	if m.Attr("id") != "m1" || m.Attr("to") != "bob@localhost" || m.Child("request", "urn:xmpp:receipts") == nil {
		t.Fatalf("message m1 altered: %s", m)
	}

	t.Log("5. delivery receipt bob -> alice (XEP-0184 passthrough)")
	bob.send("<message to='alice@localhost/android-phone' id='rc1'><received xmlns='urn:xmpp:receipts' id='m1'/></message>")
	rc := alicePhone.expect("receipt", func(e *domain.Element) bool {
		return e.Name == "message" && e.Child("received", "urn:xmpp:receipts") != nil
	})
	if rc.Child("received", "urn:xmpp:receipts").Attr("id") != "m1" {
		t.Fatalf("receipt: %s", rc)
	}

	t.Log("6. bob -> alice bare JID goes only to the highest priority device (phone)")
	bob.send("<message to='alice@localhost' type='chat' id='m2'><body>Phone only</body></message>")
	alicePhone.expect("m2 on phone", isChat("bob@localhost", "Phone only"))
	aliceTablet.silent(300*time.Millisecond, isChat("bob@localhost", "Phone only"))

	t.Log("7. tablet raises priority: both devices of alice receive")
	aliceTablet.send("<presence><priority>10</priority></presence>")
	bob.expect("tablet priority change", isPresence("alice@localhost/ios-tablet", "available"))
	bob.send("<message to='alice@localhost' type='chat' id='m3'><body>Both devices</body></message>")
	alicePhone.expect("m3 on phone", isChat("bob@localhost", "Both devices"))
	aliceTablet.expect("m3 on tablet", isChat("bob@localhost", "Both devices"))

	t.Log("8. full JID targets exactly one device")
	bob.send("<message to='alice@localhost/ios-tablet' type='chat' id='m4'><body>Tablet only</body></message>")
	aliceTablet.expect("m4 on tablet", isChat("bob@localhost", "Tablet only"))
	alicePhone.silent(300*time.Millisecond, isChat("bob@localhost", "Tablet only"))

	t.Log("9. burst of 300 messages keeps order")
	const burst = 300
	go func() {
		var b strings.Builder
		for i := 0; i < burst; i++ {
			fmt.Fprintf(&b, "<message to='bob@localhost' type='chat' id='b%d'><body>burst %d</body></message>", i, i)
		}
		alicePhone.send(b.String())
	}()
	start = time.Now()
	got := bob.messages(burst, 20*time.Second, func(e *domain.Element) bool {
		return strings.HasPrefix(e.ChildText("body", ""), "burst ")
	})
	for i, e := range got {
		if e.ChildText("body", "") != fmt.Sprintf("burst %d", i) {
			t.Fatalf("burst out of order at %d: %s", i, e)
		}
	}
	t.Logf("   %d messages in %v", burst, time.Since(start))

	t.Log("10. both directions at the same time (100 each)")
	const each = 100
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < each; i++ {
			alicePhone.send(fmt.Sprintf("<message to='bob@localhost' type='chat'><body>a2b %d</body></message>", i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < each; i++ {
			bob.send(fmt.Sprintf("<message to='alice@localhost/android-phone' type='chat'><body>b2a %d</body></message>", i))
		}
	}()
	var toBob, toAlice []*domain.Element
	var rw sync.WaitGroup
	rw.Add(2)
	go func() {
		defer rw.Done()
		toBob = bob.messages(each, 20*time.Second, func(e *domain.Element) bool { return strings.HasPrefix(e.ChildText("body", ""), "a2b ") })
	}()
	go func() {
		defer rw.Done()
		toAlice = alicePhone.messages(each, 20*time.Second, func(e *domain.Element) bool { return strings.HasPrefix(e.ChildText("body", ""), "b2a ") })
	}()
	wg.Wait()
	rw.Wait()
	for i := 0; i < each; i++ {
		if toBob[i].ChildText("body", "") != fmt.Sprintf("a2b %d", i) || toAlice[i].ChildText("body", "") != fmt.Sprintf("b2a %d", i) {
			t.Fatalf("concurrent order broken at %d", i)
		}
	}

	t.Log("11. bob loses connection: alice's devices see him go offline")
	bob.conn.Close()
	alicePhone.expect("bob offline", isPresence("bob@localhost/android-1", "unavailable"))
	aliceTablet.expect("bob offline", isPresence("bob@localhost/android-1", "unavailable"))

	t.Log("12. messages while bob is offline are queued")
	for i := 1; i <= 3; i++ {
		alicePhone.send(fmt.Sprintf("<message to='bob@localhost' type='chat' id='off%d'><body>offline %d</body></message>", i, i))
	}
	alicePhone.send("<message to='bob@localhost' type='chat' id='typing'><composing xmlns='http://jabber.org/protocol/chatstates'/></message>")
	deadline := time.Now().Add(3 * time.Second)
	for {
		_, body := api(t, s.http, "get_offline_count", map[string]any{"user": "bob", "host": "localhost"})
		if body == "3" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("offline count = %s, want 3 (typing notification must not be stored)", body)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if _, body := api(t, s.http, "user_resources", map[string]any{"user": "bob", "host": "localhost"}); body != "[]" {
		t.Fatalf("bob still has resources: %s", body)
	}

	t.Log("13. bob reconnects: nothing is pushed before initial presence")
	bob = login(t, s.c2s, "bob", "bobpw", "android-1", "SCRAM-SHA-256")
	bob.silent(300*time.Millisecond, func(e *domain.Element) bool { return e.Name == "message" })

	t.Log("14. after presence bob gets the queue in order with delay stamps")
	bob.send("<presence><priority>5</priority></presence>")
	queued := bob.messages(3, 5*time.Second, func(e *domain.Element) bool {
		return strings.HasPrefix(e.ChildText("body", ""), "offline ")
	})
	for i, e := range queued {
		d := e.Child("delay", domain.NSDelay)
		if e.ChildText("body", "") != fmt.Sprintf("offline %d", i+1) || d == nil || d.Text() != "Offline Storage" || d.Attr("stamp") == "" {
			t.Fatalf("queued message %d: %s", i, e)
		}
		if e.Attr("from") != "alice@localhost/android-phone" {
			t.Fatalf("queued message lost original sender: %s", e)
		}
	}
	alicePhone.expect("bob back online", isPresence("bob@localhost/android-1", "available"))
	if _, body := api(t, s.http, "get_offline_count", map[string]any{"user": "bob", "host": "localhost"}); body != "0" {
		t.Fatalf("offline queue not drained: %s", body)
	}

	t.Log("15. backend system notification via send_message reaches all alice devices")
	if code, body := api(t, s.http, "send_message", map[string]any{
		"type": "chat", "from": "bob@localhost", "to": "alice@localhost", "subject": "", "body": "Hệ thống: đơn hàng đã giao",
	}); body != "0" {
		t.Fatalf("send_message: %d %s", code, body)
	}
	alicePhone.expect("system msg phone", isChat("bob@localhost", "Hệ thống: đơn hàng đã giao"))
	aliceTablet.expect("system msg tablet", isChat("bob@localhost", "Hệ thống: đơn hàng đã giao"))

	t.Log("16. client and server ping keep the connection observable")
	if pong := bob.iq("ping-1", "get", "localhost", "<ping xmlns='urn:xmpp:ping'/>"); pong.Attr("type") != "result" {
		t.Fatalf("ping: %s", pong)
	}

	t.Log("17. tablet logs out: phone and bob see it leave, phone keeps chatting")
	aliceTablet.send("<presence type='unavailable'/></stream:stream>")
	bob.expect("tablet offline", isPresence("alice@localhost/ios-tablet", "unavailable"))
	alicePhone.expect("own tablet offline", isPresence("alice@localhost/ios-tablet", "unavailable"))
	bob.send("<message to='alice@localhost' type='chat' id='m9'><body>Still here?</body></message>")
	alicePhone.expect("m9", isChat("bob@localhost", "Still here?"))

	t.Log("18. admin kicks bob: session closed, alice notified")
	if _, body := api(t, s.http, "kick_user", map[string]any{"user": "bob", "host": "localhost"}); body != "1" {
		t.Fatalf("kick_user: %s", body)
	}
	bob.expect("kicked", func(e *domain.Element) bool {
		return e.Name == "error" && e.Child("policy-violation", domain.NSStreams) != nil
	})
	alicePhone.expect("bob kicked offline", isPresence("bob@localhost/android-1", "unavailable"))
}

func TestRealtimeFriendRequestFromApp(t *testing.T) {
	s := startServer(t, chatConfig)
	for _, u := range []string{"carol", "dave"} {
		api(t, s.http, "register", map[string]any{"user": u, "host": "localhost", "password": u + "pw"})
	}
	carol := login(t, s.c2s, "carol", "carolpw", "phone", "PLAIN")
	carol.send("<presence/>")
	carol.expect("own presence", isPresence("carol@localhost/phone", "available"))

	t.Log("carol invites dave while dave is offline")
	carol.send("<presence to='dave@localhost' type='subscribe'><status>Kết bạn nhé</status></presence>")
	carol.expect("pending push", func(e *domain.Element) bool {
		return isRosterPush("dave@localhost", "none")(e) && e.Child("query", domain.NSRoster).Child("item", "").Attr("ask") == "subscribe"
	})

	t.Log("dave logs in later and receives the stored invitation")
	dave := login(t, s.c2s, "dave", "davepw", "phone", "PLAIN")
	dave.send("<presence/>")
	inv := dave.expect("invitation", isPresence("carol@localhost", "subscribe"))
	if inv.ChildText("status", "") != "Kết bạn nhé" {
		t.Fatalf("invitation message lost: %s", inv)
	}

	t.Log("dave accepts and invites back; both become friends with presence")
	dave.send("<presence to='carol@localhost' type='subscribed'/>")
	dave.send("<presence to='carol@localhost' type='subscribe'/>")
	carol.expectAll(5*time.Second, map[string]func(*domain.Element) bool{
		"accepted":       isPresence("dave@localhost", "subscribed"),
		"push to":        isRosterPush("dave@localhost", "to"),
		"dave presence":  isPresence("dave@localhost/phone", "available"),
		"dave asks back": isPresence("dave@localhost", "subscribe"),
	})
	carol.send("<presence to='dave@localhost' type='subscribed'/>")
	dave.expectAll(5*time.Second, map[string]func(*domain.Element) bool{
		"push both":      isRosterPush("carol@localhost", "both"),
		"carol presence": isPresence("carol@localhost/phone", "available"),
	})

	_, body := api(t, s.http, "get_roster", map[string]any{"user": "carol", "host": "localhost"})
	if !strings.Contains(body, `"jid":"dave@localhost"`) || !strings.Contains(body, `"subscription":"both"`) {
		t.Fatalf("carol roster: %s", body)
	}

	t.Log("chat works between the new friends")
	dave.send("<message to='carol@localhost' type='chat'><body>Hi Carol</body></message>")
	carol.expect("hi", isChat("dave@localhost/phone", "Hi Carol"))

	t.Log("carol removes dave: dave is unsubscribed and sees carol offline")
	carol.iq("rm", "set", "", "<query xmlns='jabber:iq:roster'><item jid='dave@localhost' subscription='remove'/></query>")
	dave.expectAll(5*time.Second, map[string]func(*domain.Element) bool{
		"carol unavailable": isPresence("carol@localhost/phone", "unavailable"),
		"unsubscribe":       isPresence("carol@localhost", "unsubscribe"),
		"unsubscribed":      isPresence("carol@localhost", "unsubscribed"),
		"push none":         isRosterPush("carol@localhost", "none"),
	})
}
