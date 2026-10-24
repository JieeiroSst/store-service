package webhook

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

var errBlockedAddress = errors.New("webhook target resolves to a non-public address")

func isPublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
	return addr.IsValid() &&
		addr.IsGlobalUnicast() &&
		!addr.IsPrivate() &&
		!addr.IsLoopback() &&
		!addr.IsLinkLocalUnicast() &&
		!netip.MustParsePrefix("100.64.0.0/10").Contains(addr) &&
		!netip.MustParsePrefix("198.18.0.0/15").Contains(addr) &&
		!netip.MustParsePrefix("fc00::/7").Contains(addr) &&
		!netip.MustParsePrefix("64:ff9b::/96").Contains(addr)
}

func publicOnlyControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return err
	}
	if !isPublicAddr(addr) {
		return fmt.Errorf("%w: %s", errBlockedAddress, addr)
	}
	return nil
}

func newPublicOnlyClient() *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, Control: publicOnlyControl}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		MaxIdleConns:          50,
		IdleConnTimeout:       90 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
