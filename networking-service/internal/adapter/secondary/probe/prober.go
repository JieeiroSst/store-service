package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

const maxOutput = 4 << 10

type Prober struct {
	client   *http.Client
	insecure *http.Client
	dialer   net.Dialer
}

func NewProber() *Prober {
	return &Prober{
		client:   newClient(false),
		insecure: newClient(true),
	}
}

func newClient(skipVerify bool) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DisableKeepAlives = true
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipVerify}
	return &http.Client{Transport: tr}
}

func (p *Prober) Probe(ctx context.Context, c domain.Check) (domain.HealthStatus, string) {
	switch c.Type {
	case domain.CheckHTTP:
		return p.http(ctx, c)
	case domain.CheckTCP:
		return p.tcp(ctx, c)
	}
	return domain.HealthCritical, fmt.Sprintf("check type %q cannot be probed", c.Type)
}

func (p *Prober) http(ctx context.Context, c domain.Check) (domain.HealthStatus, string) {
	method := c.Method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader
	if c.Body != "" {
		body = strings.NewReader(c.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.HTTP, body)
	if err != nil {
		return domain.HealthCritical, err.Error()
	}
	for k, vs := range c.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Consul Health Check")
	}
	req.Header.Set("Accept", "text/plain, text/*, */*")

	client := p.client
	if c.TLSSkipVerify {
		client = p.insecure
	}
	resp, err := client.Do(req)
	if err != nil {
		return domain.HealthCritical, err.Error()
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(io.LimitReader(resp.Body, maxOutput))
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	output := fmt.Sprintf("HTTP %s %s: %s Output: %s", method, c.HTTP, resp.Status, strings.TrimSpace(string(out)))
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode <= 299:
		return domain.HealthPassing, output
	case resp.StatusCode == http.StatusTooManyRequests:
		return domain.HealthWarning, output
	}
	return domain.HealthCritical, output
}

func (p *Prober) tcp(ctx context.Context, c domain.Check) (domain.HealthStatus, string) {
	conn, err := p.dialer.DialContext(ctx, "tcp", c.TCP)
	if err != nil {
		return domain.HealthCritical, fmt.Sprintf("dial tcp %s: %v", c.TCP, err)
	}
	_ = conn.Close()
	return domain.HealthPassing, fmt.Sprintf("TCP connect %s: Success", c.TCP)
}
