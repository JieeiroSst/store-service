package domain

import (
	"fmt"
	"strings"
	"testing"
)

const erlangInStateChange = `
none none subscribe {none,in}
none none subscribed none
none none unsubscribe none
none none unsubscribed none
none out subscribe {none,both}
none out subscribed {to,none}
none out unsubscribe none
none out unsubscribed {none,none}
none in subscribe none
none in subscribed none
none in unsubscribe {none,none}
none in unsubscribed none
none both subscribe none
none both subscribed {to,in}
none both unsubscribe {none,out}
none both unsubscribed {none,in}
to none subscribe {to,in}
to none subscribed none
to none unsubscribe none
to none unsubscribed {none,none}
to in subscribe none
to in subscribed none
to in unsubscribe {to,none}
to in unsubscribed {none,in}
from none subscribe none
from none subscribed {both,none}
from none unsubscribe {none,none}
from none unsubscribed none
from out subscribe none
from out subscribed {both,none}
from out unsubscribe {none,out}
from out unsubscribed {from,none}
both none subscribe none
both none subscribed none
both none unsubscribe {to,none}
both none unsubscribed {from,none}
to out subscribe {to,in}
to out subscribed none
to out unsubscribe none
to out unsubscribed {none,none}
to both subscribe none
to both subscribed none
to both unsubscribe {to,none}
to both unsubscribed {none,in}
from in subscribe none
from in subscribed {both,none}
from in unsubscribe {none,none}
from in unsubscribed none
from both subscribe none
from both subscribed {both,none}
from both unsubscribe {none,out}
from both unsubscribed {from,none}
both in subscribe none
both out subscribed none
both both unsubscribe {to,none}
both out unsubscribed {from,none}
`

const erlangOutStateChange = `
none none subscribe {none,out}
none none subscribed none
none none unsubscribe none
none none unsubscribed none
none out subscribe {none,out}
none out subscribed none
none out unsubscribe {none,none}
none out unsubscribed none
none in subscribe {none,both}
none in subscribed {from,none}
none in unsubscribe none
none in unsubscribed {none,none}
none both subscribe none
none both subscribed {from,out}
none both unsubscribe {none,in}
none both unsubscribed {none,out}
to none subscribe none
to none subscribed {both,none}
to none unsubscribe {none,none}
to none unsubscribed none
to in subscribe none
to in subscribed {both,none}
to in unsubscribe {none,in}
to in unsubscribed {to,none}
from none subscribe {from,out}
from none subscribed none
from none unsubscribe none
from none unsubscribed {none,none}
from out subscribe none
from out subscribed none
from out unsubscribe {from,none}
from out unsubscribed {none,out}
both none subscribe none
both none subscribed none
both none unsubscribe {from,none}
both none unsubscribed {to,none}
to out subscribe none
to out subscribed {both,none}
to out unsubscribe {none,none}
to out unsubscribed none
to both subscribe none
to both subscribed {both,none}
to both unsubscribe {none,in}
to both unsubscribed {to,none}
from in subscribe {from,out}
from in subscribed none
from in unsubscribe none
from in unsubscribed {none,none}
from both subscribe none
from both subscribed none
from both unsubscribe {from,none}
from both unsubscribed {none,out}
both in unsubscribe {from,none}
both both unsubscribed {to,none}
`

func checkTable(t *testing.T, name, table string, fn func(Subscription, Ask, string) (Subscription, Ask, bool)) {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(table), "\n") {
		f := strings.Fields(line)
		sub, ask, ptype, want := Subscription(f[0]), Ask(f[1]), f[2], f[3]
		s, a, ok := fn(sub, ask, ptype)
		got := "none"
		if ok {
			got = fmt.Sprintf("{%s,%s}", s, a)
		}
		if got != want {
			t.Errorf("%s(%s, %s, %s) = %s, want %s", name, sub, ask, ptype, got, want)
		}
	}
}

func TestInStateChangeMatchesErlang(t *testing.T) {
	checkTable(t, "in_state_change", erlangInStateChange, InStateChange)
}

func TestOutStateChangeMatchesErlang(t *testing.T) {
	checkTable(t, "out_state_change", erlangOutStateChange, OutStateChange)
}

func TestInAutoReply(t *testing.T) {
	cases := []struct {
		sub   Subscription
		ask   Ask
		ptype string
		want  string
	}{
		{SubFrom, AskNone, PresSubscribe, PresSubscribed},
		{SubFrom, AskOut, PresSubscribe, PresSubscribed},
		{SubBoth, AskNone, PresSubscribe, PresSubscribed},
		{SubNone, AskIn, PresUnsubscribe, PresUnsubscribed},
		{SubNone, AskBoth, PresUnsubscribe, PresUnsubscribed},
		{SubTo, AskIn, PresUnsubscribe, PresUnsubscribed},
		{SubFrom, AskNone, PresUnsubscribe, PresUnsubscribed},
		{SubFrom, AskOut, PresUnsubscribe, PresUnsubscribed},
		{SubBoth, AskNone, PresUnsubscribe, PresUnsubscribed},
		{SubNone, AskNone, PresSubscribe, ""},
		{SubTo, AskNone, PresSubscribe, ""},
		{SubBoth, AskIn, PresSubscribe, ""},
	}
	for _, c := range cases {
		if got := InAutoReply(c.sub, c.ask, c.ptype); got != c.want {
			t.Errorf("in_auto_reply(%s, %s, %s) = %q, want %q", c.sub, c.ask, c.ptype, got, c.want)
		}
	}
}

func TestEncodeItemAsk(t *testing.T) {
	it := RosterItem{JID: LJID{User: "bob", Server: "localhost"}, Subscription: SubNone, Ask: AskBoth, Groups: []string{"Friends"}}
	got := it.Encode().String()
	want := `<item jid="bob@localhost" subscription="none" ask="subscribe"><group>Friends</group></item>`
	if got != want {
		t.Fatalf("encode = %s, want %s", got, want)
	}
	it.Ask = AskIn
	if it.Encode().HasAttr("ask") {
		t.Fatal("pending-in item must not carry ask attribute")
	}
}

func TestFilterRosterDropsPendingIn(t *testing.T) {
	items := []RosterItem{
		{JID: LJID{User: "a", Server: "h"}, Subscription: SubNone, Ask: AskIn},
		{JID: LJID{User: "b", Server: "h"}, Subscription: SubBoth, Ask: AskNone},
	}
	out := FilterRoster(items)
	if len(out) != 1 || out[0].JID.User != "b" {
		t.Fatalf("filter = %+v", out)
	}
}
