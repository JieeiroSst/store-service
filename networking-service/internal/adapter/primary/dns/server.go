package dns

import (
	"context"
	"encoding/hex"
	"log"
	"math/rand/v2"
	"net"
	"strings"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/application"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/miekg/dns"
)

type Handler struct {
	catalog     *application.CatalogService
	domain      string
	dc          string
	ttl         uint32
	onlyPassing bool
}

func NewHandler(catalog *application.CatalogService, cfg *config.Config) *Handler {
	return &Handler{
		catalog:     catalog,
		domain:      dns.Fqdn(cfg.DNS.Domain),
		dc:          cfg.Agent.Datacenter,
		ttl:         uint32(cfg.DNS.TTL / time.Second),
		onlyPassing: cfg.DNS.OnlyPassing,
	}
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Authoritative = true
	m.RecursionAvailable = false
	if len(req.Question) != 1 {
		m.Rcode = dns.RcodeFormatError
		_ = w.WriteMsg(m)
		return
	}
	q := req.Question[0]
	name := strings.ToLower(q.Name)
	if !dns.IsSubDomain(h.domain, name) {
		m.Authoritative = false
		m.Rcode = dns.RcodeRefused
		_ = w.WriteMsg(m)
		return
	}

	labels := dns.SplitDomainName(strings.TrimSuffix(name, h.domain))
	if n := len(labels); n > 0 && labels[n-1] == h.dc {
		labels = labels[:n-1]
	} else if n > 1 && (labels[n-2] == "service" || labels[n-2] == "node" || labels[n-2] == "addr") {
		h.nxdomain(m)
		_ = w.WriteMsg(m)
		return
	}

	switch {
	case len(labels) >= 2 && labels[len(labels)-1] == "service":
		h.service(m, q, labels[:len(labels)-1])
	case len(labels) == 2 && labels[1] == "node":
		h.node(m, q, labels[0])
	case len(labels) == 2 && labels[1] == "addr":
		h.addr(m, q, labels[0])
	default:
		h.nxdomain(m)
	}

	size := dns.MinMsgSize
	if opt := req.IsEdns0(); opt != nil {
		size = int(opt.UDPSize())
	}
	if _, isUDP := w.RemoteAddr().(*net.UDPAddr); isUDP {
		m.Truncate(size)
	}
	_ = w.WriteMsg(m)
}

func (h *Handler) service(m *dns.Msg, q dns.Question, parts []string) {
	var name, tag string
	switch {
	case len(parts) == 2 && strings.HasPrefix(parts[0], "_") && strings.HasPrefix(parts[1], "_"):
		name, tag = parts[0][1:], parts[1][1:]
		if tag == "tcp" || tag == "udp" {
			tag = ""
		}
	case len(parts) == 2:
		tag, name = parts[0], parts[1]
	case len(parts) == 1:
		name = parts[0]
	default:
		h.nxdomain(m)
		return
	}
	var tags []string
	if tag != "" {
		tags = []string{tag}
	}
	entries, _, err := h.catalog.ServiceNodes(context.Background(), application.QueryOptions{}, name, tags, false)
	if err != nil {
		log.Printf("dns: %v", err)
		m.Rcode = dns.RcodeServerFailure
		return
	}
	healthy := entries[:0]
	for _, e := range entries {
		st := domain.AggregateStatus(e.Checks)
		if st == domain.HealthCritical || (h.onlyPassing && st != domain.HealthPassing) {
			continue
		}
		healthy = append(healthy, e)
	}
	if len(healthy) == 0 {
		h.nxdomain(m)
		return
	}
	rand.Shuffle(len(healthy), func(i, j int) { healthy[i], healthy[j] = healthy[j], healthy[i] })

	for _, e := range healthy {
		addr := e.Service.Address
		if addr == "" {
			addr = e.Node.Address
		}
		switch q.Qtype {
		case dns.TypeSRV:
			target := e.Node.Name + ".node." + h.dc + "." + h.domain
			if e.Service.Address != "" && e.Service.Address != e.Node.Address {
				if ip := net.ParseIP(addr); ip != nil {
					target = hexIP(ip) + ".addr." + h.dc + "." + h.domain
				} else {
					target = dns.Fqdn(addr)
				}
			}
			m.Answer = append(m.Answer, &dns.SRV{
				Hdr:      h.hdr(q.Name, dns.TypeSRV),
				Priority: 1, Weight: 1, Port: uint16(e.Service.Port), Target: target,
			})
			if rr := h.addrRecord(target, addr, 0); rr != nil {
				m.Extra = append(m.Extra, rr)
			}
		default:
			if rr := h.addrRecord(q.Name, addr, q.Qtype); rr != nil {
				m.Answer = append(m.Answer, rr)
			}
		}
	}
	if len(m.Answer) == 0 {
		m.Ns = append(m.Ns, h.soa())
	}
}

func (h *Handler) node(m *dns.Msg, q dns.Question, name string) {
	n, _, _, err := h.catalog.Node(context.Background(), application.QueryOptions{}, name)
	if err != nil {
		log.Printf("dns: %v", err)
		m.Rcode = dns.RcodeServerFailure
		return
	}
	if n == nil {
		h.nxdomain(m)
		return
	}
	if rr := h.addrRecord(q.Name, n.Address, q.Qtype); rr != nil {
		m.Answer = append(m.Answer, rr)
	} else {
		m.Ns = append(m.Ns, h.soa())
	}
}

func (h *Handler) addr(m *dns.Msg, q dns.Question, label string) {
	b, err := hex.DecodeString(label)
	if err != nil || (len(b) != net.IPv4len && len(b) != net.IPv6len) {
		h.nxdomain(m)
		return
	}
	if rr := h.addrRecord(q.Name, net.IP(b).String(), q.Qtype); rr != nil {
		m.Answer = append(m.Answer, rr)
	} else {
		m.Ns = append(m.Ns, h.soa())
	}
}

func hexIP(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		return hex.EncodeToString(v4)
	}
	return hex.EncodeToString(ip.To16())
}

func (h *Handler) addrRecord(name, addr string, qtype uint16) dns.RR {
	ip := net.ParseIP(addr)
	switch {
	case ip == nil:
		return &dns.CNAME{Hdr: h.hdr(name, dns.TypeCNAME), Target: dns.Fqdn(addr)}
	case ip.To4() != nil && (qtype == 0 || qtype == dns.TypeA || qtype == dns.TypeANY):
		return &dns.A{Hdr: h.hdr(name, dns.TypeA), A: ip.To4()}
	case ip.To4() == nil && (qtype == 0 || qtype == dns.TypeAAAA || qtype == dns.TypeANY):
		return &dns.AAAA{Hdr: h.hdr(name, dns.TypeAAAA), AAAA: ip}
	}
	return nil
}

func (h *Handler) hdr(name string, t uint16) dns.RR_Header {
	return dns.RR_Header{Name: name, Rrtype: t, Class: dns.ClassINET, Ttl: h.ttl}
}

func (h *Handler) soa() dns.RR {
	return &dns.SOA{
		Hdr: dns.RR_Header{Name: h.domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET},
		Ns:  "ns." + h.domain, Mbox: "hostmaster." + h.domain,
		Serial: uint32(time.Now().Unix()), Refresh: 3600, Retry: 600, Expire: 86400, Minttl: 0,
	}
}

func (h *Handler) nxdomain(m *dns.Msg) {
	m.Rcode = dns.RcodeNameError
	m.Ns = append(m.Ns, h.soa())
}

type Server struct {
	addr string
	udp  *dns.Server
	tcp  *dns.Server
}

func NewServer(h *Handler, cfg *config.Config) *Server {
	s := &Server{}
	if cfg.DNS.Port == "" {
		return s
	}
	s.addr = ":" + cfg.DNS.Port
	s.udp = &dns.Server{Addr: s.addr, Net: "udp", Handler: h}
	s.tcp = &dns.Server{Addr: s.addr, Net: "tcp", Handler: h}
	return s
}

func (s *Server) Start() error {
	if s.udp == nil {
		return nil
	}
	pc, err := net.ListenPacket("udp", s.addr)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		pc.Close()
		return err
	}
	s.udp.PacketConn, s.tcp.Listener = pc, ln
	for _, srv := range []*dns.Server{s.udp, s.tcp} {
		go func() {
			if err := srv.ActivateAndServe(); err != nil {
				log.Printf("dns %s server error: %v", srv.Net, err)
			}
		}()
	}
	log.Printf("dns listening on %s (udp+tcp)", s.addr)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.udp == nil {
		return nil
	}
	_ = s.udp.ShutdownContext(ctx)
	return s.tcp.ShutdownContext(ctx)
}

func (s *Server) Addr() string {
	if s.udp == nil || s.udp.PacketConn == nil {
		return ""
	}
	return s.udp.PacketConn.LocalAddr().String()
}
