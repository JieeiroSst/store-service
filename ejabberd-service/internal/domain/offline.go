package domain

import (
	"strconv"
	"time"
)

type StoreEmptyBody string

const (
	StoreEmptyBodyTrue            StoreEmptyBody = "true"
	StoreEmptyBodyFalse           StoreEmptyBody = "false"
	StoreEmptyBodyUnlessChatState StoreEmptyBody = "unless_chat_state"
)

type OfflinePolicy struct {
	StoreGroupchat bool
	StoreEmptyBody StoreEmptyBody
}

type OfflineMessage struct {
	LUser     string
	LServer   string
	Timestamp time.Time
	Expire    *time.Time
	From      JID
	To        JID
	Packet    *Element
}

type StoreHint int

const (
	HintNone StoreHint = iota
	HintStore
	HintNoStore
)

func CheckStoreHint(msg *Element) StoreHint {
	if msg.Child("store", NSHints) != nil {
		return HintStore
	}
	if msg.Child("no-store", NSHints) != nil || msg.Child("no-storage", NSHints) != nil {
		return HintNoStore
	}
	return HintNone
}

func NeedToStore(policy OfflinePolicy, msg *Element) bool {
	msgType := StanzaType(msg)
	if msgType == "error" {
		return false
	}
	if msg.Child("offline", NSOffline) != nil {
		return false
	}
	switch CheckStoreHint(msg) {
	case HintStore:
		return true
	case HintNoStore:
		return false
	}
	store := true
	switch msgType {
	case "groupchat":
		store = policy.StoreGroupchat
	case "headline":
		store = false
	}
	if !store {
		return false
	}
	switch policy.StoreEmptyBody {
	case StoreEmptyBodyTrue:
		return true
	case StoreEmptyBodyFalse:
		return len(msg.ChildrenNamed("body", "")) > 0
	default:
		return !IsStandaloneChatState(msg)
	}
}

func IsStandaloneChatState(msg *Element) bool {
	hasState := false
	for _, el := range msg.Elements() {
		switch el.Namespace() {
		case NSChatStates:
			hasState = true
		case NSHints, NSDelay, "jabber:x:delay", "urn:xmpp:sid:0", "urn:xmpp:receipts", "urn:xmpp:chat-markers:0":
		default:
			if el.Name != "thread" {
				return false
			}
		}
	}
	return hasState
}

type EventDecision struct {
	Store bool
	Reply *Element
}

func CheckEvent(msg *Element) EventDecision {
	x := msg.Child("x", NSXEvent)
	if x == nil {
		return EventDecision{Store: true}
	}
	hasID := x.Child("id", "") != nil
	offline := x.Child("offline", "") != nil
	composing := x.Child("composing", "") != nil
	delivered := x.Child("delivered", "") != nil
	displayed := x.Child("displayed", "") != nil
	if !hasID {
		if !offline {
			return EventDecision{Store: true}
		}
		from, _ := StanzaFrom(msg)
		to, _ := StanzaTo(msg)
		reply := NewElement(KindMessage, NSClient)
		reply.SetAttr("from", to.String())
		reply.SetAttr("to", from.String())
		if v := msg.Attr("id"); v != "" {
			reply.SetAttr("id", v)
		}
		if v := msg.Attr("type"); v != "" {
			reply.SetAttr("type", v)
		}
		ev := NewElement("x", NSXEvent)
		ev.Append(NewElement("id", "").AppendText(msg.Attr("id")))
		ev.Append(NewElement("offline", ""))
		reply.Append(ev)
		return EventDecision{Store: true, Reply: reply}
	}
	if composing {
		return EventDecision{Store: false}
	}
	if !delivered && !displayed && !offline {
		return EventDecision{Store: false}
	}
	return EventDecision{Store: true}
}

func FindExpire(now time.Time, msg *Element) *time.Time {
	x := msg.Child("x", NSXExpire)
	if x == nil {
		return nil
	}
	secs, err := strconv.Atoi(x.Attr("seconds"))
	if err != nil {
		return nil
	}
	t := now.Add(time.Duration(secs) * time.Second)
	return &t
}

func (m OfflineMessage) Expired(now time.Time) bool {
	exp := m.Expire
	if exp == nil {
		exp = FindExpire(now, m.Packet)
	}
	return exp != nil && !exp.After(now)
}
