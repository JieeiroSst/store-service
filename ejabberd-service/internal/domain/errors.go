package domain

import "errors"

var (
	ErrUserExists       = errors.New("user already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrNotAllowed       = errors.New("not allowed")
	ErrNotMyHost        = errors.New("host is not served by this server")
	ErrOfflineQueueFull = errors.New("offline queue full")
	ErrNotFound         = errors.New("not found")
)

type StanzaError struct {
	Type      string
	Condition string
	Text      string
	Lang      string
	Extra     []*Element
}

func (e StanzaError) Error() string {
	if e.Text != "" {
		return e.Condition + ": " + e.Text
	}
	return e.Condition
}

func (e StanzaError) Element() *Element {
	el := NewElement("error", "")
	el.SetAttr("type", e.Type)
	el.Append(NewElement(e.Condition, NSStanzas))
	if e.Text != "" {
		t := NewElement("text", NSStanzas)
		if e.Lang != "" {
			t.SetAttr("xml:lang", e.Lang)
		}
		t.AppendText(e.Text)
		el.Append(t)
	}
	for _, x := range e.Extra {
		el.Append(x.Copy())
	}
	return el
}

func stanzaErr(typ, cond, text string) StanzaError {
	return StanzaError{Type: typ, Condition: cond, Text: text}
}

func ErrBadRequest(text string) StanzaError { return stanzaErr("modify", "bad-request", text) }
func ErrConflict(text string) StanzaError   { return stanzaErr("cancel", "conflict", text) }
func ErrFeatureNotImplemented(text string) StanzaError {
	return stanzaErr("cancel", "feature-not-implemented", text)
}
func ErrForbidden(text string) StanzaError { return stanzaErr("auth", "forbidden", text) }
func ErrInternalServerError(text string) StanzaError {
	return stanzaErr("wait", "internal-server-error", text)
}
func ErrItemNotFound(text string) StanzaError { return stanzaErr("cancel", "item-not-found", text) }
func ErrJIDMalformed(text string) StanzaError { return stanzaErr("modify", "jid-malformed", text) }
func ErrNotAcceptable(text string) StanzaError {
	return stanzaErr("modify", "not-acceptable", text)
}
func ErrNotAllowedStanza(text string) StanzaError { return stanzaErr("cancel", "not-allowed", text) }
func ErrNotAuthorized(text string) StanzaError    { return stanzaErr("auth", "not-authorized", text) }
func ErrRemoteServerNotFound(text string) StanzaError {
	return stanzaErr("cancel", "remote-server-not-found", text)
}
func ErrResourceConstraint(text string) StanzaError {
	return stanzaErr("wait", "resource-constraint", text)
}
func ErrServiceUnavailable(text string) StanzaError {
	return stanzaErr("cancel", "service-unavailable", text)
}

func MakeError(stanza *Element, se StanzaError) *Element {
	out := stanza.Copy()
	from, to := out.Attr("from"), out.Attr("to")
	out.RemoveAttr("from").RemoveAttr("to")
	if to != "" {
		out.SetAttr("from", to)
	}
	if from != "" {
		out.SetAttr("to", from)
	}
	out.SetAttr("type", "error")
	if se.Lang == "" {
		se.Lang = Lang(stanza)
	}
	out.Append(se.Element())
	return out
}

type StreamError struct {
	Condition string
	Text      string
}

func (e StreamError) Error() string {
	if e.Text != "" {
		return e.Condition + ": " + e.Text
	}
	return e.Condition
}

func (e StreamError) Element() *Element {
	el := &Element{Name: "stream:error"}
	el.Append(NewElement(e.Condition, NSStreams))
	if e.Text != "" {
		el.Append(NewElement("text", NSStreams).AppendText(e.Text))
	}
	return el
}

func StreamErr(cond, text string) StreamError { return StreamError{Condition: cond, Text: text} }
