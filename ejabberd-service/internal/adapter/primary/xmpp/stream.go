package xmpp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/application"
	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

type streamState int

const (
	stateWaitTLS streamState = iota
	stateWaitAuth
	stateWaitBind
	stateEstablished
)

type stream struct {
	h        *Handler
	listener config.Listener
	conn     *streamConn
	state    streamState
	server   string
	lang     string
	user     string
	sasl     domain.SASLMechanism
	saslMech string
	session  *application.C2SSession
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *stream) tlsEnabled() bool {
	return s.listener.StartTLS && s.h.tls != nil && !s.conn.Encrypted()
}

func (s *stream) run(ctx context.Context) {
	defer func() {
		if s.session != nil {
			s.session.Terminate("connection closed")
			<-s.session.Done()
		}
		s.conn.Close(nil)
	}()
	if s.conn.Encrypted() {
		s.state = stateWaitAuth
	} else if s.tlsEnabled() {
		s.state = stateWaitTLS
	} else {
		s.state = stateWaitAuth
	}
	timeout := s.h.cfg.NegotiationTimeout
	for {
		if s.state != stateEstablished && timeout > 0 {
			_ = s.conn.conn.SetReadDeadline(time.Now().Add(timeout))
		} else {
			_ = s.conn.conn.SetReadDeadline(time.Time{})
		}
		restart, err := s.serveStream(ctx)
		if err != nil {
			s.handleReadError(err)
			return
		}
		if !restart {
			return
		}
		s.conn.resetDecoder()
	}
}

func (s *stream) handleReadError(err error) {
	var serr domain.StreamError
	switch {
	case errors.As(err, &serr):
		s.conn.Close(&serr)
	case errors.Is(err, errStanzaTooBig):
		e := domain.StreamErr("policy-violation", "XML stanza is too big")
		s.conn.Close(&e)
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
	default:
		var netErr interface{ Timeout() bool }
		if errors.As(err, &netErr) && netErr.Timeout() {
			e := domain.StreamErr("connection-timeout", "")
			s.conn.Close(&e)
			return
		}
		var synErr *xml.SyntaxError
		if errors.As(err, &synErr) {
			e := domain.StreamErr("not-well-formed", "")
			s.conn.Close(&e)
			return
		}
		slog.Debug("c2s stream closed", "remote", s.conn.RemoteIP(), "err", err)
	}
}

func (s *stream) serveStream(ctx context.Context) (bool, error) {
	if err := s.readStreamHeader(); err != nil {
		return false, err
	}
	for {
		s.conn.reader.reset()
		tok, err := s.conn.decoder.Token()
		if err != nil {
			return false, err
		}
		switch t := tok.(type) {
		case xml.EndElement:
			return false, nil
		case xml.StartElement:
			el, err := domain.ReadElement(s.conn.decoder, t, domain.NSClient)
			if err != nil {
				return false, err
			}
			restart, err := s.handleElement(ctx, el)
			if err != nil || restart {
				return restart, err
			}
		}
	}
}

func (s *stream) readStreamHeader() error {
	for {
		tok, err := s.conn.decoder.Token()
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "stream" || se.Name.Space != domain.NSStream {
			s.sendHeader(s.h.cfg.MyName(), newID())
			return domain.StreamErr("invalid-namespace", "")
		}
		var to, version string
		defaultNS := ""
		for _, a := range se.Attr {
			switch {
			case a.Name.Local == "to" && a.Name.Space == "":
				to = a.Value
			case a.Name.Local == "version" && a.Name.Space == "":
				version = a.Value
			case a.Name.Local == "lang":
				s.lang = a.Value
			case a.Name.Space == "" && a.Name.Local == "xmlns":
				defaultNS = a.Value
			}
		}
		if s.lang == "" {
			s.lang = s.h.cfg.Language
		}
		server, okHost := domain.Nameprep(to)
		if s.server != "" {
			server = s.server
		}
		if defaultNS != "" && defaultNS != domain.NSClient {
			s.sendHeader(s.h.cfg.MyName(), newID())
			return domain.StreamErr("invalid-namespace", "")
		}
		if to == "" || !okHost || !s.h.cfg.IsMyHost(server) {
			s.sendHeader(s.h.cfg.MyName(), newID())
			return domain.StreamErr("host-unknown", "")
		}
		s.server = server
		s.sendHeader(server, newID())
		if version != "1.0" {
			return domain.StreamErr("unsupported-version", "")
		}
		if s.state == stateWaitTLS || s.state == stateWaitAuth {
			s.conn.reader.shaper = s.h.shaperFor(s.listener, domain.MustJID("", server, ""))
		}
		s.sendFeatures()
		return nil
	}
}

func (s *stream) sendHeader(from, id string) {
	_ = s.conn.writeRaw(fmt.Sprintf(
		`<?xml version='1.0'?><stream:stream xmlns='%s' xmlns:stream='%s' id='%s' from='%s' version='1.0' xml:lang='%s'>`,
		domain.NSClient, domain.NSStream, id, domain.EscapeAttr(from), domain.EscapeAttr(s.lang)))
}

func (s *stream) sendFeatures() {
	f := &domain.Element{Name: "stream:features"}
	switch s.state {
	case stateWaitTLS, stateWaitAuth:
		if s.tlsEnabled() {
			st := domain.NewElement("starttls", domain.NSTLS)
			if s.listener.StartTLSRequired {
				st.Append(domain.NewElement("required", ""))
			}
			f.Append(st)
		}
		if !(s.listener.StartTLSRequired && !s.conn.Encrypted()) {
			mechs := domain.NewElement("mechanisms", domain.NSSASL)
			for _, m := range s.h.auth.Mechanisms() {
				mechs.Append(domain.NewElement("mechanism", "").AppendText(m))
			}
			f.Append(mechs)
			if s.h.mods.RegisterFeatureEnabled() {
				f.Append(domain.NewElement("register", domain.NSFeatureIQReg))
			}
		}
	case stateWaitBind:
		f.Append(domain.NewElement("bind", domain.NSBind))
		sess := domain.NewElement("session", domain.NSSession)
		sess.Append(domain.NewElement("optional", ""))
		f.Append(sess)
		if s.h.cfg.Modules.Roster.Enabled && s.h.cfg.Modules.Roster.Versioning {
			f.Append(domain.NewElement("ver", domain.NSRosterVer))
		}
	}
	_ = s.conn.writeRaw(f.String())
}

func (s *stream) handleElement(ctx context.Context, el *domain.Element) (bool, error) {
	ns := el.Namespace()
	switch s.state {
	case stateWaitTLS, stateWaitAuth:
		switch {
		case el.Name == "starttls" && ns == domain.NSTLS:
			return s.handleStartTLS()
		case (el.Name == "auth" || el.Name == "response" || el.Name == "abort") && ns == domain.NSSASL:
			if s.listener.StartTLSRequired && !s.conn.Encrypted() {
				return false, domain.StreamErr("policy-violation", "Use of STARTTLS required")
			}
			return s.handleSASL(ctx, el)
		case el.Name == domain.KindIQ && s.h.mods.RegisterFeatureEnabled() && isRegisterIQ(el):
			if s.listener.StartTLSRequired && !s.conn.Encrypted() {
				return false, domain.StreamErr("policy-violation", "Use of STARTTLS required")
			}
			s.handleUnauthenticatedRegister(ctx, el)
			return false, nil
		}
		return false, domain.StreamErr("not-authorized", "")
	case stateWaitBind:
		if el.Name == domain.KindIQ {
			if b := el.Child("bind", domain.NSBind); b != nil && domain.StanzaType(el) == "set" {
				s.handleBind(ctx, el, b)
				return false, nil
			}
			if domain.StanzaType(el) == "get" || domain.StanzaType(el) == "set" {
				s.conn.Send(domain.MakeError(el, domain.ErrNotAuthorized("")))
				return false, nil
			}
		}
		return false, domain.StreamErr("not-authorized", "")
	case stateEstablished:
		if !domain.IsStanza(el) {
			return false, nil
		}
		if el.Namespace() == "" {
			el.SetAttr("xmlns", domain.NSClient)
		}
		s.session.HandleClientPacket(el)
		return false, nil
	}
	return false, nil
}

func isRegisterIQ(el *domain.Element) bool {
	t := domain.StanzaType(el)
	if t != "get" && t != "set" {
		return false
	}
	els := el.Elements()
	return len(els) == 1 && els[0].Name == "query" && els[0].Namespace() == domain.NSRegister
}

func (s *stream) handleStartTLS() (bool, error) {
	if !s.tlsEnabled() {
		return false, domain.StreamErr("policy-violation", "")
	}
	_ = s.conn.writeRaw(domain.NewElement("proceed", domain.NSTLS).String())
	if err := s.conn.startTLS(s.h.tls); err != nil {
		slog.Warn("failed to secure c2s connection", "remote", s.conn.RemoteIP(), "err", err)
		return false, err
	}
	s.state = stateWaitAuth
	return true, nil
}

func (s *stream) saslFailure(cond, text string) {
	f := domain.NewElement("failure", domain.NSSASL)
	f.Append(domain.NewElement(cond, ""))
	if text != "" {
		f.Append(domain.NewElement("text", "").SetAttr("xml:lang", s.lang).AppendText(text))
	}
	_ = s.conn.writeRaw(f.String())
}

func (s *stream) handleSASL(ctx context.Context, el *domain.Element) (bool, error) {
	var input []byte
	switch el.Name {
	case "abort":
		s.sasl = nil
		s.saslFailure("aborted", "")
		return false, nil
	case "auth":
		mech := el.Attr("mechanism")
		m, err := s.h.auth.NewSASL(ctx, mech, s.server)
		if err != nil {
			s.saslFailure("invalid-mechanism", "")
			return false, nil
		}
		s.sasl, s.saslMech = m, mech
	case "response":
		if s.sasl == nil {
			s.saslFailure("malformed-request", "")
			return false, nil
		}
	}
	text := el.Text()
	if text != "" && text != "=" {
		var err error
		input, err = base64.StdEncoding.DecodeString(text)
		if err != nil {
			s.sasl = nil
			s.saslFailure("incorrect-encoding", "")
			return false, nil
		}
	}
	if el.Name == "auth" && text == "" {
		_ = s.conn.writeRaw(domain.NewElement("challenge", domain.NSSASL).String())
		return false, nil
	}
	step, err := s.sasl.Step(input)
	if err != nil {
		cond := err.Error()
		switch {
		case errors.Is(err, domain.ErrSASLNotAuthorized), errors.Is(err, domain.ErrSASLMalformed),
			errors.Is(err, domain.ErrSASLBadProtocol), errors.Is(err, domain.ErrSASLInvalidMech):
		default:
			cond = "temporary-auth-failure"
		}
		slog.Warn("failed c2s authentication", "mech", s.saslMech, "user", step.User, "server", s.server,
			"ip", s.conn.RemoteIP(), "reason", cond)
		s.sasl = nil
		if cond == domain.ErrSASLNotAuthorized.Error() {
			s.saslFailure(cond, "Invalid username or password")
		} else {
			s.saslFailure(cond, "")
		}
		return false, nil
	}
	if !step.Success {
		ch := domain.NewElement("challenge", domain.NSSASL)
		ch.AppendText(base64.StdEncoding.EncodeToString(step.Challenge))
		_ = s.conn.writeRaw(ch.String())
		return false, nil
	}
	succ := domain.NewElement("success", domain.NSSASL)
	if len(step.Data) > 0 {
		succ.AppendText(base64.StdEncoding.EncodeToString(step.Data))
	}
	_ = s.conn.writeRaw(succ.String())
	s.user = step.User
	s.sasl = nil
	s.state = stateWaitBind
	slog.Info("accepted c2s authentication", "mech", s.saslMech, "user", s.user, "server", s.server,
		"backend", s.h.cfg.AuthMethod, "ip", s.conn.RemoteIP())
	return true, nil
}

func (s *stream) handleUnauthenticatedRegister(ctx context.Context, el *domain.Element) {
	iq := el.Copy()
	if iq.Namespace() == "" {
		iq.SetAttr("xmlns", domain.NSClient)
	}
	iq.RemoveAttr("from")
	iq.SetAttr("to", s.server)
	res := s.h.mods.ProcessRegister(ctx, iq, s.conn.IP())
	if res == nil {
		return
	}
	res.RemoveAttr("from").RemoveAttr("to")
	s.conn.Send(res)
}

func (s *stream) handleBind(ctx context.Context, iq, bind *domain.Element) {
	resource := bind.ChildText("resource", "")
	if resource != "" {
		if _, ok := domain.Resourceprep(resource); !ok {
			s.conn.Send(domain.MakeError(iq, domain.ErrBadRequest("Malformed resource")))
			return
		}
	}
	sess, serr := s.h.c2s.Bind(ctx, application.BindParams{
		User:       s.user,
		Server:     s.server,
		Resource:   resource,
		IP:         s.conn.IP(),
		Conn:       s.connType(),
		AuthModule: s.h.cfg.AuthMethod,
		Access:     s.listener.Access,
		Lang:       s.lang,
		Stream:     s.conn,
	})
	if serr != nil {
		serr.Lang = s.lang
		s.conn.Send(domain.MakeError(iq, *serr))
		return
	}
	s.session = sess
	s.state = stateEstablished
	s.conn.reader.shaper = s.h.shaperFor(s.listener, sess.JID())
	res := domain.NewElement("bind", domain.NSBind)
	res.Append(domain.NewElement("jid", "").AppendText(sess.JID().String()))
	s.conn.Send(domain.MakeIQResult(iq, res))
}

func (s *stream) connType() string {
	if s.conn.Encrypted() {
		return "c2s_tls"
	}
	return "c2s"
}
