package domain

import (
	"crypto/sha1"
	"encoding/hex"
	"slices"
	"strings"
)

type Subscription string

const (
	SubNone   Subscription = "none"
	SubTo     Subscription = "to"
	SubFrom   Subscription = "from"
	SubBoth   Subscription = "both"
	SubRemove Subscription = "remove"
)

type Ask string

const (
	AskNone        Ask = "none"
	AskIn          Ask = "in"
	AskOut         Ask = "out"
	AskBoth        Ask = "both"
	AskSubscribe   Ask = "subscribe"
	AskUnsubscribe Ask = "unsubscribe"
)

const (
	PresSubscribe    = "subscribe"
	PresSubscribed   = "subscribed"
	PresUnsubscribe  = "unsubscribe"
	PresUnsubscribed = "unsubscribed"
)

func IsSubscriptionPresence(ptype string) bool {
	switch ptype {
	case PresSubscribe, PresSubscribed, PresUnsubscribe, PresUnsubscribed:
		return true
	}
	return false
}

type RosterItem struct {
	LUser        string
	LServer      string
	JID          LJID
	Name         string
	Subscription Subscription
	Ask          Ask
	Groups       []string
	AskMessage   string
	Xs           []*Element
}

func NewRosterItem(luser, lserver string, jid LJID) RosterItem {
	return RosterItem{LUser: luser, LServer: lserver, JID: jid.Bare(), Subscription: SubNone, Ask: AskNone}
}

func (r RosterItem) IsFrom() bool { return r.Subscription == SubBoth || r.Subscription == SubFrom }
func (r RosterItem) IsTo() bool   { return r.Subscription == SubBoth || r.Subscription == SubTo }

func (r RosterItem) IsPendingIn() bool {
	return r.Subscription == SubNone && r.Ask == AskIn
}

func AskToPending(a Ask) Ask {
	switch a {
	case AskSubscribe:
		return AskOut
	case AskUnsubscribe, "":
		return AskNone
	}
	return a
}

type stateKey struct {
	sub   Subscription
	ask   Ask
	ptype string
}

type subState struct {
	sub Subscription
	ask Ask
}

var inStateTable = map[stateKey]subState{
	{SubNone, AskNone, PresSubscribe}:    {SubNone, AskIn},
	{SubNone, AskOut, PresSubscribe}:     {SubNone, AskBoth},
	{SubNone, AskOut, PresSubscribed}:    {SubTo, AskNone},
	{SubNone, AskOut, PresUnsubscribed}:  {SubNone, AskNone},
	{SubNone, AskIn, PresUnsubscribe}:    {SubNone, AskNone},
	{SubNone, AskBoth, PresSubscribed}:   {SubTo, AskIn},
	{SubNone, AskBoth, PresUnsubscribe}:  {SubNone, AskOut},
	{SubNone, AskBoth, PresUnsubscribed}: {SubNone, AskIn},
	{SubTo, AskNone, PresSubscribe}:      {SubTo, AskIn},
	{SubTo, AskNone, PresUnsubscribed}:   {SubNone, AskNone},
	{SubTo, AskIn, PresUnsubscribe}:      {SubTo, AskNone},
	{SubTo, AskIn, PresUnsubscribed}:     {SubNone, AskIn},
	{SubFrom, AskNone, PresSubscribed}:   {SubBoth, AskNone},
	{SubFrom, AskNone, PresUnsubscribe}:  {SubNone, AskNone},
	{SubFrom, AskOut, PresSubscribed}:    {SubBoth, AskNone},
	{SubFrom, AskOut, PresUnsubscribe}:   {SubNone, AskOut},
	{SubFrom, AskOut, PresUnsubscribed}:  {SubFrom, AskNone},
	{SubTo, AskOut, PresSubscribe}:       {SubTo, AskIn},
	{SubTo, AskOut, PresUnsubscribed}:    {SubNone, AskNone},
	{SubTo, AskBoth, PresUnsubscribe}:    {SubTo, AskNone},
	{SubTo, AskBoth, PresUnsubscribed}:   {SubNone, AskIn},
	{SubFrom, AskIn, PresSubscribed}:     {SubBoth, AskNone},
	{SubFrom, AskIn, PresUnsubscribe}:    {SubNone, AskNone},
	{SubFrom, AskBoth, PresSubscribed}:   {SubBoth, AskNone},
	{SubFrom, AskBoth, PresUnsubscribe}:  {SubNone, AskOut},
	{SubFrom, AskBoth, PresUnsubscribed}: {SubFrom, AskNone},
}

var outStateTable = map[stateKey]subState{
	{SubNone, AskNone, PresSubscribe}:    {SubNone, AskOut},
	{SubNone, AskOut, PresSubscribe}:     {SubNone, AskOut},
	{SubNone, AskOut, PresUnsubscribe}:   {SubNone, AskNone},
	{SubNone, AskIn, PresSubscribe}:      {SubNone, AskBoth},
	{SubNone, AskIn, PresSubscribed}:     {SubFrom, AskNone},
	{SubNone, AskIn, PresUnsubscribed}:   {SubNone, AskNone},
	{SubNone, AskBoth, PresSubscribed}:   {SubFrom, AskOut},
	{SubNone, AskBoth, PresUnsubscribe}:  {SubNone, AskIn},
	{SubNone, AskBoth, PresUnsubscribed}: {SubNone, AskOut},
	{SubTo, AskNone, PresSubscribed}:     {SubBoth, AskNone},
	{SubTo, AskNone, PresUnsubscribe}:    {SubNone, AskNone},
	{SubTo, AskIn, PresSubscribed}:       {SubBoth, AskNone},
	{SubTo, AskIn, PresUnsubscribe}:      {SubNone, AskIn},
	{SubTo, AskIn, PresUnsubscribed}:     {SubTo, AskNone},
	{SubFrom, AskNone, PresSubscribe}:    {SubFrom, AskOut},
	{SubFrom, AskNone, PresUnsubscribed}: {SubNone, AskNone},
	{SubFrom, AskOut, PresUnsubscribe}:   {SubFrom, AskNone},
	{SubFrom, AskOut, PresUnsubscribed}:  {SubNone, AskOut},
	{SubTo, AskOut, PresSubscribed}:      {SubBoth, AskNone},
	{SubTo, AskOut, PresUnsubscribe}:     {SubNone, AskNone},
	{SubTo, AskBoth, PresSubscribed}:     {SubBoth, AskNone},
	{SubTo, AskBoth, PresUnsubscribe}:    {SubNone, AskIn},
	{SubTo, AskBoth, PresUnsubscribed}:   {SubTo, AskNone},
	{SubFrom, AskIn, PresSubscribe}:      {SubFrom, AskOut},
	{SubFrom, AskIn, PresUnsubscribed}:   {SubNone, AskNone},
	{SubFrom, AskBoth, PresUnsubscribe}:  {SubFrom, AskNone},
	{SubFrom, AskBoth, PresUnsubscribed}: {SubNone, AskOut},
}

func InStateChange(sub Subscription, ask Ask, ptype string) (Subscription, Ask, bool) {
	ask = AskToPending(ask)
	if sub == SubBoth {
		switch ptype {
		case PresUnsubscribe:
			return SubTo, AskNone, true
		case PresUnsubscribed:
			return SubFrom, AskNone, true
		}
		return "", "", false
	}
	s, ok := inStateTable[stateKey{sub, ask, ptype}]
	return s.sub, s.ask, ok
}

func OutStateChange(sub Subscription, ask Ask, ptype string) (Subscription, Ask, bool) {
	ask = AskToPending(ask)
	if sub == SubBoth {
		switch ptype {
		case PresUnsubscribe:
			return SubFrom, AskNone, true
		case PresUnsubscribed:
			return SubTo, AskNone, true
		}
		return "", "", false
	}
	s, ok := outStateTable[stateKey{sub, ask, ptype}]
	return s.sub, s.ask, ok
}

func InAutoReply(sub Subscription, ask Ask, ptype string) string {
	ask = AskToPending(ask)
	switch ptype {
	case PresSubscribe:
		switch {
		case sub == SubFrom && ask == AskNone,
			sub == SubFrom && ask == AskOut,
			sub == SubBoth && ask == AskNone:
			return PresSubscribed
		}
	case PresUnsubscribe:
		switch {
		case sub == SubNone && ask == AskIn,
			sub == SubNone && ask == AskBoth,
			sub == SubTo && ask == AskIn,
			sub == SubFrom && ask == AskNone,
			sub == SubFrom && ask == AskOut,
			sub == SubBoth && ask == AskNone:
			return PresUnsubscribed
		}
	}
	return ""
}

func (r RosterItem) Encode() *Element {
	el := NewElement("item", "")
	el.SetAttr("jid", r.JID.String())
	if r.Name != "" {
		el.SetAttr("name", r.Name)
	}
	sub := r.Subscription
	if sub == "" {
		sub = SubNone
	}
	el.SetAttr("subscription", string(sub))
	switch AskToPending(r.Ask) {
	case AskOut, AskBoth:
		el.SetAttr("ask", "subscribe")
	}
	for _, g := range r.Groups {
		el.Append(NewElement("group", "").AppendText(g))
	}
	return el
}

func FilterRoster(items []RosterItem) []RosterItem {
	out := make([]RosterItem, 0, len(items))
	for _, it := range items {
		if !it.IsPendingIn() {
			out = append(out, it)
		}
	}
	return out
}

func RosterHash(items []RosterItem) string {
	lines := make([]string, 0, len(items))
	for _, it := range items {
		groups := slices.Clone(it.Groups)
		slices.Sort(groups)
		lines = append(lines, it.Encode().String()+strings.Join(groups, "\x00"))
	}
	slices.Sort(lines)
	sum := sha1.Sum([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

func HasDuplicatedGroups(groups []string) bool {
	seen := make(map[string]struct{}, len(groups))
	for _, g := range groups {
		if _, ok := seen[g]; ok {
			return true
		}
		seen[g] = struct{}{}
	}
	return false
}

func NickFromSubEls(els []*Element, def string) string {
	for _, el := range els {
		if el.Name == "nick" && el.Namespace() == NSNick {
			return el.Text()
		}
	}
	return def
}
