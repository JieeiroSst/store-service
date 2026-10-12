package xmpp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"sync"

	"github.com/JIeeiroSst/ejabberd-service/internal/application"
	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"go.uber.org/fx"
)

type Handler struct {
	cfg     *config.Config
	matcher *domain.ACLMatcher
	auth    *application.AuthService
	mods    *application.Modules
	c2s     *application.C2SService
	tls     *tls.Config
}

type Params struct {
	fx.In

	Config  *config.Config
	Matcher *domain.ACLMatcher
	Auth    *application.AuthService
	Modules *application.Modules
	C2S     *application.C2SService
	TLS     *tls.Config `optional:"true"`
}

func NewHandler(p Params) *Handler {
	return &Handler{cfg: p.Config, matcher: p.Matcher, auth: p.Auth, mods: p.Modules, c2s: p.C2S, tls: p.TLS}
}

func (h *Handler) shaperFor(l config.Listener, jid domain.JID) *domain.Shaper {
	name := h.matcher.MatchShaper(l.Shaper, domain.MatchJID(jid))
	spec, ok := h.cfg.Shapers[name]
	if !ok {
		return nil
	}
	return domain.NewShaper(spec)
}

func (h *Handler) Serve(ctx context.Context, l config.Listener, c net.Conn) {
	s := &stream{h: h, listener: l, conn: newStreamConn(c, l.MaxStanzaSize)}
	s.run(ctx)
}

var Module = fx.Options(
	fx.Provide(NewHandler),
	fx.Invoke(startListeners),
)

func listenAddr(l config.Listener) string {
	ip := l.IP
	if ip == "" || ip == "::" || ip == "0.0.0.0" {
		ip = ""
	}
	return net.JoinHostPort(ip, strconv.Itoa(l.Port))
}

func startListeners(lc fx.Lifecycle, cfg *config.Config, h *Handler, c2s *application.C2SService) {
	ctx, cancel := context.WithCancel(context.Background())
	var (
		wg        sync.WaitGroup
		listeners []net.Listener
		connsMu   sync.Mutex
		conns     = map[net.Conn]struct{}{}
	)
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			for _, l := range cfg.Listen {
				if l.Module != "ejabberd_c2s" {
					continue
				}
				if l.Transport != "" && l.Transport != "tcp" {
					slog.Warn("unsupported c2s transport, skipping", "port", l.Port, "transport", l.Transport)
					continue
				}
				ln, err := net.Listen("tcp", listenAddr(l))
				if err != nil {
					return fmt.Errorf("listen c2s %s: %w", listenAddr(l), err)
				}
				if l.TLS {
					if h.tls == nil {
						_ = ln.Close()
						return fmt.Errorf("listener on port %d requires tls but no certificate is available", l.Port)
					}
					ln = tls.NewListener(ln, h.tls)
				}
				listeners = append(listeners, ln)
				slog.Info("c2s listener started", "addr", ln.Addr().String(), "tls", l.TLS, "starttls", l.StartTLS, "starttls_required", l.StartTLSRequired)
				wg.Add(1)
				go func(l config.Listener, ln net.Listener) {
					defer wg.Done()
					for {
						c, err := ln.Accept()
						if err != nil {
							if errors.Is(err, net.ErrClosed) {
								return
							}
							slog.Warn("c2s accept failed", "err", err)
							continue
						}
						connsMu.Lock()
						conns[c] = struct{}{}
						connsMu.Unlock()
						wg.Add(1)
						go func() {
							defer wg.Done()
							defer func() {
								connsMu.Lock()
								delete(conns, c)
								connsMu.Unlock()
							}()
							h.Serve(ctx, l, c)
						}()
					}
				}(l, ln)
			}
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			for _, ln := range listeners {
				_ = ln.Close()
			}
			connsMu.Lock()
			for c := range conns {
				_ = c.Close()
			}
			connsMu.Unlock()
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-stopCtx.Done():
			}
			c2s.Wait(stopCtx)
			return nil
		},
	})
}
