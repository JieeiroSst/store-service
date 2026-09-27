package security

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	High   = "high"
	Medium = "medium"
	Low    = "low"
	Info   = "info"
)

type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

type Options struct {
	ProtectedPaths []string
	SkipRateLimit  bool
}

var rank = map[string]int{High: 0, Medium: 1, Low: 2, Info: 3}

func Scan(ctx context.Context, c *http.Client, baseURL string, opt Options) []Finding {
	base := strings.TrimRight(baseURL, "/")
	if _, err := url.Parse(base); err != nil {
		return []Finding{{ID: "SEC-000", Severity: High, Title: "invalid base URL", Detail: err.Error()}}
	}
	var out []Finding
	add := func(id, sev, title, detail, fix string) { out = append(out, Finding{id, sev, title, detail, fix}) }

	resp, _, err := do(ctx, c, http.MethodGet, base+"/", nil)
	if err != nil {
		add("SEC-001", High, "target not reachable or TLS validation failed", err.Error(), "Check the URL, certificate chain and hostname.")
		return out
	}
	final := resp.Request.URL
	local := isLocal(final.Hostname())

	if final.Scheme != "https" {
		sev := Medium
		if local {
			sev = Info
		}
		add("SEC-TLS-001", sev, "traffic is not encrypted (HTTP)", "Final URL: "+final.String(), "Serve the site over HTTPS and redirect HTTP to HTTPS.")
	}
	if resp.TLS != nil {
		if resp.TLS.Version < tls.VersionTLS12 {
			add("SEC-TLS-002", High, "TLS version below 1.2", tlsName(resp.TLS.Version), "Disable TLS 1.0/1.1.")
		}
		if certs := resp.TLS.PeerCertificates; len(certs) > 0 {
			days := int(time.Until(certs[0].NotAfter).Hours() / 24)
			switch {
			case days < 0:
				add("SEC-TLS-003", High, "certificate expired", certs[0].NotAfter.Format(time.RFC3339), "Renew the certificate.")
			case days < 14:
				add("SEC-TLS-003", Medium, fmt.Sprintf("certificate expires in %d days", days), "", "Renew the certificate; automate renewal.")
			case days < 30:
				add("SEC-TLS-003", Low, fmt.Sprintf("certificate expires in %d days", days), "", "Make sure renewal is automated.")
			}
		}
	}

	h := resp.Header
	if final.Scheme == "https" && h.Get("Strict-Transport-Security") == "" {
		add("SEC-HDR-001", Medium, "missing Strict-Transport-Security", "", "Add: Strict-Transport-Security: max-age=31536000; includeSubDomains")
	}
	if !strings.EqualFold(h.Get("X-Content-Type-Options"), "nosniff") {
		add("SEC-HDR-002", Low, "missing X-Content-Type-Options: nosniff", "", "Add the header.")
	}
	csp := h.Get("Content-Security-Policy")

	isHTML := strings.Contains(strings.ToLower(h.Get("Content-Type")), "html")
	if isHTML && csp == "" {
		add("SEC-HDR-003", Medium, "missing Content-Security-Policy", "", "Define a CSP that limits script and frame sources.")
	}
	if isHTML && h.Get("X-Frame-Options") == "" && !strings.Contains(strings.ToLower(csp), "frame-ancestors") {
		add("SEC-HDR-004", Medium, "page can be framed (clickjacking)", "No X-Frame-Options or CSP frame-ancestors", "Add X-Frame-Options: DENY or CSP frame-ancestors 'none'.")
	}
	if h.Get("Referrer-Policy") == "" {
		add("SEC-HDR-005", Low, "missing Referrer-Policy", "", "Add Referrer-Policy: strict-origin-when-cross-origin")
	}
	for _, k := range []string{"Server", "X-Powered-By", "X-AspNet-Version"} {
		if v := h.Get(k); v != "" && strings.ContainsAny(v, "0123456789/") {
			add("SEC-HDR-006", Low, "software/version disclosed in "+k, k+": "+v, "Remove or genericize the header.")
		}
	}

	for _, ck := range resp.Cookies() {
		var miss []string
		if final.Scheme == "https" && !ck.Secure {
			miss = append(miss, "Secure")
		}
		if !ck.HttpOnly {
			miss = append(miss, "HttpOnly")
		}
		if ck.SameSite == http.SameSiteDefaultMode {
			miss = append(miss, "SameSite")
		}
		if len(miss) > 0 {
			add("SEC-CK-001", Medium, fmt.Sprintf("cookie %q lacks %s", ck.Name, strings.Join(miss, ", ")), "", "Set Secure, HttpOnly and SameSite on session cookies.")
		}
	}

	if r, _, err := do(ctx, c, http.MethodGet, base+"/", map[string]string{"Origin": "https://evil.example"}); err == nil {
		acao, acac := r.Header.Get("Access-Control-Allow-Origin"), r.Header.Get("Access-Control-Allow-Credentials")
		switch {
		case acao == "https://evil.example" && strings.EqualFold(acac, "true"):
			add("SEC-CORS-001", High, "CORS reflects any origin with credentials", "Origin https://evil.example was allowed with credentials", "Use an explicit allow-list of origins.")
		case acao == "https://evil.example":
			add("SEC-CORS-002", Medium, "CORS reflects arbitrary origins", "", "Use an explicit allow-list of origins.")
		case acao == "*":
			add("SEC-CORS-003", Info, "CORS allows any origin (*)", "Fine for public APIs, wrong for authenticated ones.", "")
		}
	}

	probe := randHex()
	if r, b, err := do(ctx, c, "TRACE", base+"/", map[string]string{"X-Trace-Probe": probe}); err == nil && r.StatusCode == http.StatusOK && strings.Contains(string(b), probe) {
		add("SEC-MTH-001", Medium, "HTTP TRACE is enabled", "", "Disable TRACE.")
	}

	_, soft, _ := do(ctx, c, http.MethodGet, base+"/"+randHex()+"-not-found", nil)
	for _, p := range exposed {
		r, b, err := do(ctx, c, http.MethodGet, base+p.path, nil)
		if err != nil || r.StatusCode != http.StatusOK || string(b) == string(soft) {
			continue
		}
		if strings.Contains(string(b), p.sig) {
			add("SEC-EXP-"+p.id, p.sev, "exposed "+p.path, "Matched signature "+fmt.Sprintf("%q", p.sig), "Remove or restrict access to this path.")
		}
	}

	if _, b, err := do(ctx, c, http.MethodGet, base+"/"+randHex()+"?q=%27", nil); err == nil {
		low := strings.ToLower(string(b))
		for _, s := range []string{"traceback (most recent", "goroutine ", "panic:", "at java.", "sqlstate", "exception in thread", "stack trace", "unhandled exception"} {
			if strings.Contains(low, s) {
				add("SEC-ERR-001", Medium, "error page leaks internals", "Response contains "+fmt.Sprintf("%q", s), "Return generic error pages; log details server-side.")
				break
			}
		}
	}

	for _, p := range opt.ProtectedPaths {
		path := "/" + strings.TrimLeft(p, "/")
		r, b, err := do(ctx, c, http.MethodGet, base+path, nil)
		if err == nil && r.StatusCode >= 200 && r.StatusCode < 300 && string(b) != string(soft) {
			add("SEC-AUTH-001", High, "protected endpoint answers without credentials", fmt.Sprintf("GET %s returned %d anonymously", path, r.StatusCode), "Enforce authentication and authorization on this endpoint.")
		}
	}

	if !opt.SkipRateLimit {
		if !rateLimited(ctx, c, base+"/") {
			add("SEC-RL-001", Info, "no rate limiting observed", "30 rapid requests were all served with no 429/Retry-After/X-RateLimit headers", "Consider rate limiting, especially on login and payment endpoints.")
		}
	}

	sort.SliceStable(out, func(i, j int) bool { return rank[out[i].Severity] < rank[out[j].Severity] })
	return out
}

var exposed = []struct{ id, path, sig, sev string }{
	{"ENV", "/.env", "=", High},
	{"GIT", "/.git/config", "[core]", High},
	{"GITHEAD", "/.git/HEAD", "ref:", High},
	{"ACTENV", "/actuator/env", "propertySources", High},
	{"PPROF", "/debug/pprof/", "profiles", Medium},
	{"PHPINFO", "/phpinfo.php", "phpinfo()", Medium},
	{"STATUS", "/server-status", "Apache Server Status", Medium},
	{"SWAGGER", "/swagger.json", "swagger", Info},
	{"OPENAPI", "/openapi.json", "openapi", Info},
	{"METRICS", "/metrics", "# HELP", Info},
}

func rateLimited(ctx context.Context, c *http.Client, u string) bool {
	var mu sync.Mutex
	limited := false
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	for i := 0; i < 30; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			r, _, err := do(ctx, c, http.MethodGet, u, nil)
			if err != nil {
				return
			}
			if r.StatusCode == http.StatusTooManyRequests || r.Header.Get("Retry-After") != "" || r.Header.Get("X-RateLimit-Limit") != "" || r.Header.Get("RateLimit-Limit") != "" {
				mu.Lock()
				limited = true
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return limited
}

func do(ctx context.Context, c *http.Client, method, u string, hdr map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	return resp, b, nil
}

func randHex() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func isLocal(host string) bool {
	return host == "localhost" || host == "::1" || strings.HasPrefix(host, "127.") || strings.HasSuffix(host, ".local")
}

func tlsName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	}
	return fmt.Sprintf("0x%x", v)
}
