package domain

import (
	"crypto/rand"
	"encoding/base64"
	"net"
	"strings"
	"testing"
	"time"
)

func TestParseJID(t *testing.T) {
	cases := []struct {
		in, user, server, resource string
		ok                         bool
	}{
		{"Alice@LocalHost/Phone", "alice", "localhost", "Phone", true},
		{"localhost", "", "localhost", "", true},
		{"localhost/res", "", "localhost", "res", true},
		{"@localhost", "", "", "", false},
		{"a@localhost/", "", "", "", false},
		{"a b@localhost", "", "", "", false},
		{"", "", "", "", false},
	}
	for _, c := range cases {
		j, err := ParseJID(c.in)
		if (err == nil) != c.ok {
			t.Fatalf("ParseJID(%q) err = %v", c.in, err)
		}
		if !c.ok {
			continue
		}
		if j.LUser != c.user || j.LServer != c.server || j.LResource != c.resource {
			t.Fatalf("ParseJID(%q) = %+v", c.in, j.LJID())
		}
	}
}

func TestElementRoundTrip(t *testing.T) {
	src := `<message xmlns="jabber:client" to="b@h" type="chat"><body>a &amp; b &lt;c&gt;</body><x xmlns="jabber:x:event"><offline/></x></message>`
	el, err := ParseElement(src)
	if err != nil {
		t.Fatal(err)
	}
	if el.ChildText("body", "") != "a & b <c>" {
		t.Fatalf("body = %q", el.ChildText("body", ""))
	}
	if el.Child("x", NSXEvent) == nil {
		t.Fatal("missing x:event child")
	}
	if got := el.String(); got != src {
		t.Fatalf("round trip:\n got %s\nwant %s", got, src)
	}
	if got := el.StringNS(NSClient); strings.Contains(got, `xmlns="jabber:client"`) {
		t.Fatalf("StringNS must omit inherited namespace: %s", got)
	}
}

func TestMakeError(t *testing.T) {
	el, _ := ParseElement(`<iq xmlns="jabber:client" type="get" id="1" from="a@h/r" to="h"><query xmlns="jabber:iq:version"/></iq>`)
	e := MakeError(el, ErrServiceUnavailable(""))
	if e.Attr("type") != "error" || e.Attr("to") != "a@h/r" || e.Attr("from") != "h" {
		t.Fatalf("bad error stanza: %s", e)
	}
	if e.Child("error", "").Child("service-unavailable", NSStanzas) == nil {
		t.Fatalf("missing condition: %s", e)
	}
}

func TestNeedToStore(t *testing.T) {
	p := OfflinePolicy{StoreEmptyBody: StoreEmptyBodyUnlessChatState}
	parse := func(s string) *Element {
		el, err := ParseElement(s)
		if err != nil {
			t.Fatal(err)
		}
		return el
	}
	cases := []struct {
		xml  string
		want bool
	}{
		{`<message type="chat"><body>hi</body></message>`, true},
		{`<message><body>hi</body></message>`, true},
		{`<message type="headline"><body>hi</body></message>`, false},
		{`<message type="groupchat"><body>hi</body></message>`, false},
		{`<message type="error"><body>hi</body></message>`, false},
		{`<message type="chat"><composing xmlns="http://jabber.org/protocol/chatstates"/></message>`, false},
		{`<message type="chat"><body>x</body><no-store xmlns="urn:xmpp:hints"/></message>`, false},
		{`<message type="headline"><store xmlns="urn:xmpp:hints"/></message>`, true},
	}
	for _, c := range cases {
		if got := NeedToStore(p, parse(c.xml)); got != c.want {
			t.Errorf("NeedToStore(%s) = %v, want %v", c.xml, got, c.want)
		}
	}
}

func TestCheckEvent(t *testing.T) {
	el, _ := ParseElement(`<message from="a@h/r" to="b@h" id="m1"><x xmlns="jabber:x:event"><offline/></x><body>x</body></message>`)
	d := CheckEvent(el)
	if !d.Store || d.Reply == nil || d.Reply.Attr("to") != "a@h/r" {
		t.Fatalf("unexpected decision %+v", d)
	}
	el, _ = ParseElement(`<message><x xmlns="jabber:x:event"><id>1</id><composing/></x></message>`)
	if CheckEvent(el).Store {
		t.Fatal("composing events must not be stored")
	}
}

func TestFindExpire(t *testing.T) {
	now := time.Unix(1000, 0)
	el, _ := ParseElement(`<message><x xmlns="jabber:x:expire" seconds="30"/></message>`)
	exp := FindExpire(now, el)
	if exp == nil || !exp.Equal(now.Add(30*time.Second)) {
		t.Fatalf("expire = %v", exp)
	}
}

func TestACLMatchRuleOrdered(t *testing.T) {
	m := &ACLMatcher{
		ACLs: map[string][]ACLSpec{
			"blocked": {{Kind: ACLUser, User: "spammer"}},
			"admin":   {{Kind: ACLUser, User: "admin", Server: "localhost"}},
		},
		AccessRules: map[string]AccessRule{
			"c2s": {
				{Allow: false, Specs: []ACLSpec{{Kind: ACLRef, Name: "blocked"}}},
				{Allow: true, Specs: []ACLSpec{{Kind: ACLRef, Name: "all"}}},
			},
		},
		ShaperRules: map[string]ShaperRule{
			"max_user_offline_messages": {
				{Value: "5000", Specs: []ACLSpec{{Kind: ACLRef, Name: "admin"}}},
				{Value: "100", Specs: []ACLSpec{{Kind: ACLRef, Name: "all"}}},
			},
		},
		IsMyHost: func(h string) bool { return h == "localhost" },
	}
	if m.MatchRule("c2s", MatchJID(MustJID("spammer", "localhost", ""))) {
		t.Fatal("blocked user must be denied")
	}
	if !m.MatchRule("c2s", MatchJID(MustJID("bob", "localhost", ""))) {
		t.Fatal("regular user must be allowed")
	}
	if m.MatchRule("unknown", MatchJID(MustJID("bob", "localhost", ""))) {
		t.Fatal("unknown rule must deny")
	}
	if v := m.MatchShaper("max_user_offline_messages", MatchJID(MustJID("admin", "localhost", ""))); v != "5000" {
		t.Fatalf("admin limit = %s", v)
	}
	if v := m.MatchShaper("max_user_offline_messages", MatchJID(MustJID("bob", "localhost", ""))); v != "100" {
		t.Fatalf("user limit = %s", v)
	}
	_, loop, _ := net.ParseCIDR("127.0.0.0/8")
	m.AccessRules["trusted"] = AccessRule{{Allow: true, Specs: []ACLSpec{{Kind: ACLIP, Net: loop}}}}
	if !m.MatchRule("trusted", MatchIP(net.ParseIP("127.0.0.1"))) || m.MatchRule("trusted", MatchIP(net.ParseIP("10.0.0.1"))) {
		t.Fatal("ip acl mismatch")
	}
}

type staticCreds struct {
	pass Password
}

func (c staticCreds) GetPassword(string) (Password, bool) { return c.pass, true }

func (c staticCreds) CheckPassword(_, _, p string) bool { return MatchPassword(p, c.pass) }

func scramClient(t *testing.T, h ScramHash, mech SASLMechanism, user, password string) error {
	t.Helper()
	cnonceRaw := make([]byte, 12)
	_, _ = rand.Read(cnonceRaw)
	cnonce := base64.RawStdEncoding.EncodeToString(cnonceRaw)
	clientFirstBare := "n=" + user + ",r=" + cnonce
	step, err := mech.Step([]byte("n,," + clientFirstBare))
	if err != nil {
		return err
	}
	attrs := parseScramAttrs(string(step.Challenge))
	salt, _ := base64.StdEncoding.DecodeString(attrs["s"])
	iter := 0
	for _, c := range attrs["i"] {
		iter = iter*10 + int(c-'0')
	}
	salted := SaltedPassword(h, password, salt, iter)
	clientKey := ClientKey(h, salted)
	stored := StoredKey(h, clientKey)
	finalNoProof := "c=" + base64.StdEncoding.EncodeToString([]byte("n,,")) + ",r=" + attrs["r"]
	authMsg := clientFirstBare + "," + string(step.Challenge) + "," + finalNoProof
	sig := HMAC(h, stored, []byte(authMsg))
	proof := make([]byte, len(clientKey))
	for i := range proof {
		proof[i] = clientKey[i] ^ sig[i]
	}
	step, err = mech.Step([]byte(finalNoProof + ",p=" + base64.StdEncoding.EncodeToString(proof)))
	if err != nil {
		return err
	}
	serverSig := HMAC(h, ServerKey(h, salted), []byte(authMsg))
	if string(step.Data) != "v="+base64.StdEncoding.EncodeToString(serverSig) {
		t.Fatalf("server signature mismatch")
	}
	if !step.Success {
		t.Fatal("expected success")
	}
	return nil
}

func TestSCRAMWithPlainAndScramStorage(t *testing.T) {
	for _, h := range []ScramHash{ScramSHA1, ScramSHA256, ScramSHA512} {
		mechName := map[ScramHash]string{ScramSHA1: "SCRAM-SHA-1", ScramSHA256: "SCRAM-SHA-256", ScramSHA512: "SCRAM-SHA-512"}[h]
		sc, err := PasswordToScram(h, "secret", ScramDefaultIterationCount)
		if err != nil {
			t.Fatal(err)
		}
		for _, stored := range []Password{{Plain: "secret"}, {Scram: sc}} {
			m, _ := NewSASLMechanism(mechName, staticCreds{pass: stored})
			if err := scramClient(t, h, m, "alice", "secret"); err != nil {
				t.Fatalf("%s scram=%v: %v", mechName, stored.IsScram(), err)
			}
			m, _ = NewSASLMechanism(mechName, staticCreds{pass: stored})
			if err := scramClient(t, h, m, "alice", "wrong"); err != ErrSASLNotAuthorized {
				t.Fatalf("%s wrong password err = %v", mechName, err)
			}
		}
	}
}

func TestPlainSASL(t *testing.T) {
	m, _ := NewSASLMechanism("PLAIN", staticCreds{pass: Password{Plain: "pw"}})
	step, err := m.Step([]byte("\x00alice\x00pw"))
	if err != nil || !step.Success || step.User != "alice" {
		t.Fatalf("plain = %+v %v", step, err)
	}
	m, _ = NewSASLMechanism("PLAIN", staticCreds{pass: Password{Plain: "pw"}})
	if _, err := m.Step([]byte("\x00alice\x00bad")); err != ErrSASLNotAuthorized {
		t.Fatalf("err = %v", err)
	}
}

func TestSASLMechanismsFollowStoreType(t *testing.T) {
	if got := strings.Join(SASLMechanisms(PasswordPlain, ScramSHA1), ","); got != "SCRAM-SHA-1,SCRAM-SHA-256,SCRAM-SHA-512,PLAIN" {
		t.Fatal(got)
	}
	if got := strings.Join(SASLMechanisms(PasswordScram, ScramSHA256), ","); got != "SCRAM-SHA-256,PLAIN" {
		t.Fatal(got)
	}
}

func TestShaper(t *testing.T) {
	s := NewShaper(ShaperSpec{Rate: 1000, Burst: 1000})
	now := time.Unix(0, 0)
	s.now = func() time.Time { return now }
	s.lastTime = now
	if d := s.Update(500); d != 0 {
		t.Fatalf("pause = %v", d)
	}
	if d := s.Update(1000); d != 500*time.Millisecond {
		t.Fatalf("pause = %v", d)
	}
}

func TestUploadSlotSignature(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	slot, err := NewUploadSlot(MustJID("alice", "localhost", "r"), "my photo (1).png", 1234, now)
	if err != nil {
		t.Fatal(err)
	}
	if slot.FileName != "my_photo__1_.png" || len(slot.Secret) != 40 || len(slot.UserStr) != 40 {
		t.Fatalf("slot = %+v", slot)
	}
	key := []byte("k")
	sig := SignSlot(key, slot.Path(), slot.Size, slot.Expires)
	if !VerifySlot(key, slot.Path(), slot.Size, slot.Expires, sig, now) {
		t.Fatal("valid signature rejected")
	}
	if VerifySlot(key, slot.Path(), slot.Size+1, slot.Expires, sig, now) {
		t.Fatal("size is not covered by the signature")
	}
	if VerifySlot([]byte("other"), slot.Path(), slot.Size, slot.Expires, sig, now) {
		t.Fatal("signature accepted with another key")
	}
	if VerifySlot(key, slot.Path(), slot.Size, slot.Expires, sig, now.Add(SlotTimeout+time.Second)) {
		t.Fatal("expired slot accepted")
	}
}
