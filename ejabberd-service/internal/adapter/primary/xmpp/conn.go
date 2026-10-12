package xmpp

import (
	"bufio"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

var errStanzaTooBig = errors.New("stanza too big")

type meteredReader struct {
	r      io.Reader
	limit  int
	n      int
	shaper *domain.Shaper
}

func (m *meteredReader) Read(p []byte) (int, error) {
	n, err := m.r.Read(p)
	if n > 0 {
		m.n += n
		if pause := m.shaper.Update(n); pause > 0 {
			time.Sleep(pause)
		}
		if m.limit > 0 && m.n > m.limit {
			return n, errStanzaTooBig
		}
	}
	return n, err
}

func (m *meteredReader) reset() { m.n = 0 }

type streamConn struct {
	raw       net.Conn
	conn      net.Conn
	encrypted bool
	writeMu   sync.Mutex
	closeOnce sync.Once
	closed    chan struct{}
	reader    *meteredReader
	decoder   *xml.Decoder
}

var _ port.StreamConn = (*streamConn)(nil)

func newStreamConn(c net.Conn, maxStanza int) *streamConn {
	_, isTLS := c.(*tls.Conn)
	sc := &streamConn{raw: c, conn: c, encrypted: isTLS, closed: make(chan struct{})}
	sc.reader = &meteredReader{r: c, limit: maxStanza}
	sc.resetDecoder()
	return sc
}

func (s *streamConn) resetDecoder() {
	s.reader.r = s.conn
	s.decoder = xml.NewDecoder(bufio.NewReader(s.reader))
	s.decoder.Strict = true
}

func (s *streamConn) startTLS(cfg *tls.Config) error {
	t := tls.Server(s.conn, cfg)
	_ = t.SetDeadline(time.Now().Add(30 * time.Second))
	if err := t.Handshake(); err != nil {
		return err
	}
	_ = t.SetDeadline(time.Time{})
	s.conn = t
	s.encrypted = true
	s.resetDecoder()
	return nil
}

func (s *streamConn) writeRaw(data string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	select {
	case <-s.closed:
		return net.ErrClosed
	default:
	}
	_ = s.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	_, err := io.WriteString(s.conn, data)
	return err
}

func (s *streamConn) Send(el *domain.Element) {
	if err := s.writeRaw(el.StringNS(domain.NSClient)); err != nil && !errors.Is(err, net.ErrClosed) {
		slog.Debug("c2s write failed", "remote", s.RemoteIP(), "err", err)
		s.Close(nil)
	}
}

func (s *streamConn) Close(serr *domain.StreamError) {
	s.closeOnce.Do(func() {
		if serr != nil {
			_ = s.writeRaw(serr.Element().String())
		}
		_ = s.writeRaw("</stream:stream>")
		s.writeMu.Lock()
		close(s.closed)
		s.writeMu.Unlock()
		_ = s.conn.Close()
	})
}

func (s *streamConn) RemoteIP() string {
	if addr, ok := s.raw.RemoteAddr().(*net.TCPAddr); ok {
		return addr.IP.String()
	}
	host, _, _ := net.SplitHostPort(s.raw.RemoteAddr().String())
	return host
}

func (s *streamConn) IP() net.IP {
	return net.ParseIP(s.RemoteIP())
}

func (s *streamConn) Encrypted() bool { return s.encrypted }
