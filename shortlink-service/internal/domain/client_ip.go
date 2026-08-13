package domain

import "strings"

type TrustProxy struct {
	Enabled  bool
	HopCount int
}

func ParseTrustProxyEnv(v string) TrustProxy {
	v = strings.TrimSpace(v)
	if v == "" {
		return TrustProxy{}
	}
	if v == "1" || strings.EqualFold(v, "true") {
		return TrustProxy{Enabled: true, HopCount: 0}
	}
	n := 0
	parsed := true
	for _, c := range v {
		if c < '0' || c > '9' {
			parsed = false
			break
		}
	}
	if parsed {
		for _, c := range v {
			n = n*10 + int(c-'0')
		}
		return TrustProxy{Enabled: true, HopCount: n}
	}
	return TrustProxy{}
}

func normalizeIPMapped(ip string) string {
	return strings.TrimPrefix(ip, "::ffff:")
}

func ResolveClientIP(tp TrustProxy, remoteAddr, xForwardedFor string) string {
	if !tp.Enabled || xForwardedFor == "" {
		return normalizeIPMapped(remoteAddr)
	}

	hops := splitAndTrim(xForwardedFor)
	if len(hops) == 0 {
		return normalizeIPMapped(remoteAddr)
	}

	if tp.HopCount <= 0 {
		return normalizeIPMapped(hops[0])
	}

	idx := len(hops) - tp.HopCount
	if idx < 0 {
		idx = 0
	}
	return normalizeIPMapped(hops[idx])
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
