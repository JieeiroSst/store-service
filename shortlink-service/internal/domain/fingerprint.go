package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
	"time"
)

type FingerprintData struct {
	IPAddress       string
	UserAgent       string
	Timezone        string
	Language        string
	ScreenWidth     *int
	ScreenHeight    *int
	Platform        string
	PlatformVersion string
}

type FingerprintMatch struct {
	ClickID         string
	LinkID          string
	ConfidenceScore int
	MatchedFactors  []string
	ClickedAt       time.Time
}

const (
	WeightIPAddress               = 40
	WeightUserAgent               = 30
	WeightTimezone                = 10
	WeightLanguage                = 10
	WeightScreenResolution        = 10
	DefaultAttributionWindowHours = 168
	MaxAttributionWindowHours     = 2160
	ConfidenceThreshold           = 70
)

func GenerateFingerprintHash(d FingerprintData) string {
	components := []string{
		d.IPAddress,
		d.UserAgent,
		d.Timezone,
		d.Language,
		intPtrToString(d.ScreenWidth),
		intPtrToString(d.ScreenHeight),
		d.Platform,
		d.PlatformVersion,
	}
	concatenated := strings.Join(components, "|")
	sum := sha256.Sum256([]byte(concatenated))
	return hex.EncodeToString(sum[:])
}

func intPtrToString(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}

var nonAttributableV4 = []struct {
	base   string
	prefix int
}{
	{"0.0.0.0", 8},
	{"10.0.0.0", 8},
	{"100.64.0.0", 10},  // CGNAT (RFC 6598)
	{"127.0.0.0", 8},    // loopback
	{"169.254.0.0", 16}, // link-local
	{"172.16.0.0", 12},  // private
	{"192.0.0.0", 24},   // IETF protocol assignments
	{"192.168.0.0", 16}, // private
	{"198.18.0.0", 15},  // benchmarking
}

func IsAttributableIP(ip string) bool {
	if ip == "" {
		return false
	}
	addr := strings.TrimSpace(ip)
	addr = strings.TrimPrefix(addr, "::ffff:")

	if strings.Contains(addr, ".") && !strings.Contains(addr, ":") {
		parsed := net.ParseIP(addr).To4()
		if parsed == nil {
			return false
		}
		for _, r := range nonAttributableV4 {
			_, cidr, err := net.ParseCIDR(r.base + "/" + strconv.Itoa(r.prefix))
			if err != nil {
				continue
			}
			if cidr.Contains(parsed) {
				return false
			}
		}
		return true
	}

	if strings.Contains(addr, ":") {
		low := strings.ToLower(addr)
		if low == "::" || low == "::1" {
			return false
		}
		if strings.HasPrefix(low, "fc") || strings.HasPrefix(low, "fd") {
			return false // fc00::/7 ULA
		}
		if strings.HasPrefix(low, "fe8") || strings.HasPrefix(low, "fe9") ||
			strings.HasPrefix(low, "fea") || strings.HasPrefix(low, "feb") {
			return false // fe80::/10 link-local
		}
		return true
	}

	return false
}

func normalizeIP(ip string) string {
	if ip == "" {
		return ""
	}
	if strings.Contains(ip, ".") {
		parts := strings.Split(ip, ".")
		if len(parts) > 3 {
			parts = parts[:3]
		}
		return strings.Join(parts, ".")
	}
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		if len(parts) > 4 {
			parts = parts[:4]
		}
		return strings.Join(parts, ":")
	}
	return ip
}

var (
	uaPlatformNames = []string{"iPhone", "iPad", "Android", "Windows", "Macintosh", "Linux"}
	uaBrowserNames  = []string{"Chrome", "Safari", "Firefox", "Edge", "Opera"}
)


func normalizeUserAgent(ua string) string {
	if ua == "" {
		return ""
	}
	lower := strings.ToLower(ua)
	platform := ""
	for _, p := range uaPlatformNames {
		if strings.Contains(lower, strings.ToLower(p)) {
			platform = p
			break
		}
	}
	browser := ""
	for _, b := range uaBrowserNames {
		if strings.Contains(lower, strings.ToLower(b)) {
			browser = b
			break
		}
	}
	return strings.ToLower(platform + "|" + browser)
}

func CalculateConfidenceScore(a, b FingerprintData) (score int, matchedFactors []string) {
	matchedFactors = []string{}

	if a.IPAddress != "" && b.IPAddress != "" && IsAttributableIP(a.IPAddress) && IsAttributableIP(b.IPAddress) {
		if normalizeIP(a.IPAddress) == normalizeIP(b.IPAddress) {
			score += WeightIPAddress
			matchedFactors = append(matchedFactors, "ip")
		}
	}

	if a.UserAgent != "" && b.UserAgent != "" {
		if normalizeUserAgent(a.UserAgent) == normalizeUserAgent(b.UserAgent) {
			score += WeightUserAgent
			matchedFactors = append(matchedFactors, "user_agent")
		}
	}

	if a.Timezone != "" && b.Timezone != "" && a.Timezone == b.Timezone {
		score += WeightTimezone
		matchedFactors = append(matchedFactors, "timezone")
	}

	if a.Language != "" && b.Language != "" {
		la, lb := firstTwoLower(a.Language), firstTwoLower(b.Language)
		if la == lb {
			score += WeightLanguage
			matchedFactors = append(matchedFactors, "language")
		}
	}

	if a.ScreenWidth != nil && a.ScreenHeight != nil && b.ScreenWidth != nil && b.ScreenHeight != nil {
		if *a.ScreenWidth == *b.ScreenWidth && *a.ScreenHeight == *b.ScreenHeight {
			score += WeightScreenResolution
			matchedFactors = append(matchedFactors, "screen")
		}
	}

	return score, matchedFactors
}

func firstTwoLower(s string) string {
	s = strings.ToLower(s)
	if len(s) <= 2 {
		return s
	}
	return s[:2]
}
