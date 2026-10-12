package domain

import (
	"strconv"
	"time"
)

const (
	NSClient       = "jabber:client"
	NSStream       = "http://etherx.jabber.org/streams"
	NSStanzas      = "urn:ietf:params:xml:ns:xmpp-stanzas"
	NSStreams      = "urn:ietf:params:xml:ns:xmpp-streams"
	NSTLS          = "urn:ietf:params:xml:ns:xmpp-tls"
	NSSASL         = "urn:ietf:params:xml:ns:xmpp-sasl"
	NSBind         = "urn:ietf:params:xml:ns:xmpp-bind"
	NSSession      = "urn:ietf:params:xml:ns:xmpp-session"
	NSRoster       = "jabber:iq:roster"
	NSRosterVer    = "urn:xmpp:features:rosterver"
	NSPing         = "urn:xmpp:ping"
	NSDiscoInfo    = "http://jabber.org/protocol/disco#info"
	NSDiscoItems   = "http://jabber.org/protocol/disco#items"
	NSVersion      = "jabber:iq:version"
	NSRegister     = "jabber:iq:register"
	NSFeatureIQReg = "http://jabber.org/features/iq-register"
	NSDelay        = "urn:xmpp:delay"
	NSOffline      = "http://jabber.org/protocol/offline"
	NSMsgOffline   = "msgoffline"
	NSHints        = "urn:xmpp:hints"
	NSXEvent       = "jabber:x:event"
	NSXExpire      = "jabber:x:expire"
	NSChatStates   = "http://jabber.org/protocol/chatstates"
	NSNick         = "http://jabber.org/protocol/nick"
	NSLast         = "jabber:iq:last"
	NSHTTPUpload   = "urn:xmpp:http:upload:0"
	NSXData        = "jabber:x:data"
	NSOOB          = "jabber:x:oob"
)

const (
	KindMessage  = "message"
	KindPresence = "presence"
	KindIQ       = "iq"
)

func IsStanza(el *Element) bool {
	switch el.Name {
	case KindMessage, KindPresence, KindIQ:
		return true
	}
	return false
}

func StanzaType(el *Element) string {
	t := el.Attr("type")
	if t != "" {
		return t
	}
	switch el.Name {
	case KindMessage:
		return "normal"
	case KindPresence:
		return "available"
	}
	return ""
}

func StanzaFrom(el *Element) (JID, bool) {
	return attrJID(el, "from")
}

func StanzaTo(el *Element) (JID, bool) {
	return attrJID(el, "to")
}

func attrJID(el *Element, name string) (JID, bool) {
	v := el.Attr(name)
	if v == "" {
		return JID{}, false
	}
	j, err := ParseJID(v)
	if err != nil {
		return JID{}, false
	}
	return j, true
}

func SetFromTo(el *Element, from, to JID) *Element {
	el.SetAttr("from", from.String())
	el.SetAttr("to", to.String())
	return el
}

func Lang(el *Element) string {
	return el.Attr("xml:lang")
}

func Priority(presence *Element) int {
	if presence == nil {
		return 0
	}
	v := presence.ChildText("priority", "")
	if v == "" {
		return 0
	}
	p, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	if p < -128 {
		return -128
	}
	if p > 127 {
		return 127
	}
	return p
}

func StatusText(presence *Element) string {
	return presence.ChildText("status", "")
}

func NewPresence(ptype string, from, to JID) *Element {
	el := NewElement(KindPresence, NSClient)
	if ptype != "" && ptype != "available" {
		el.SetAttr("type", ptype)
	}
	if !from.IsZero() {
		el.SetAttr("from", from.String())
	}
	if !to.IsZero() {
		el.SetAttr("to", to.String())
	}
	return el
}

func NewIQ(iqType, id string, from, to JID, payload *Element) *Element {
	el := NewElement(KindIQ, NSClient)
	el.SetAttr("type", iqType)
	el.SetAttr("id", id)
	if !from.IsZero() {
		el.SetAttr("from", from.String())
	}
	if !to.IsZero() {
		el.SetAttr("to", to.String())
	}
	if payload != nil {
		el.Append(payload)
	}
	return el
}

func NewMessage(msgType string, from, to JID, subject, body string) *Element {
	el := NewElement(KindMessage, NSClient)
	if msgType != "" {
		el.SetAttr("type", msgType)
	}
	el.SetAttr("from", from.String())
	el.SetAttr("to", to.String())
	if subject != "" {
		el.Append(NewElement("subject", "").AppendText(subject))
	}
	if body != "" {
		el.Append(NewElement("body", "").AppendText(body))
	}
	return el
}

func IQPayload(iq *Element) *Element {
	els := iq.Elements()
	if len(els) == 0 {
		return nil
	}
	return els[0]
}

func MakeIQResult(iq *Element, payload *Element) *Element {
	res := &Element{Name: KindIQ}
	if ns := iq.Namespace(); ns != "" {
		res.SetAttr("xmlns", ns)
	}
	res.SetAttr("type", "result")
	res.SetAttr("id", iq.Attr("id"))
	if v := iq.Attr("to"); v != "" {
		res.SetAttr("from", v)
	}
	if v := iq.Attr("from"); v != "" {
		res.SetAttr("to", v)
	}
	if payload != nil {
		res.Append(payload)
	}
	return res
}

func AddDelay(el *Element, from JID, ts time.Time, desc string) *Element {
	d := NewElement("delay", NSDelay)
	d.SetAttr("from", from.String())
	d.SetAttr("stamp", ts.UTC().Format("2006-01-02T15:04:05.000000Z"))
	d.AppendText(desc)
	el.Append(d)
	return el
}

func HasSubtag(el *Element, name, xmlns string) bool {
	return el.Child(name, xmlns) != nil
}
