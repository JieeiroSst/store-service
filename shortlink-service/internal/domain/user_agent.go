package domain

import (
	"regexp"
	"strings"
)

type DeviceType string

const (
	DeviceIOS     DeviceType = "ios"
	DeviceAndroid DeviceType = "android"
	DeviceWeb     DeviceType = "web"
)

func DetectDevice(userAgent string) DeviceType {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		return DeviceIOS
	}
	if strings.Contains(ua, "android") {
		return DeviceAndroid
	}
	return DeviceWeb
}

type ParsedUserAgent struct {
	DeviceType      string // "mobile" | "tablet" | "desktop" (best effort)
	Platform        string // OS name, e.g. "iOS", "Android", "Windows", "Mac OS", "Linux", "unknown"
	PlatformVersion string
	Browser         string
}

var (
	iosVersionRe     = regexp.MustCompile(`(?i)(?:iphone\s+os|cpu\s+os|cpu\s+iphone\s+os)\s+([\d_]+)`)
	androidVersionRe = regexp.MustCompile(`(?i)android\s+([\d.]+)`)
	windowsVersionRe = regexp.MustCompile(`(?i)windows\s+nt\s+([\d.]+)`)
	macVersionRe     = regexp.MustCompile(`(?i)mac\s+os\s+x\s+([\d_]+)`)
)

func ParseUserAgent(userAgent string) ParsedUserAgent {
	ua := userAgent
	lower := strings.ToLower(ua)

	result := ParsedUserAgent{DeviceType: "desktop", Platform: "unknown"}

	switch {
	case strings.Contains(lower, "ipad"):
		result.DeviceType = "tablet"
		result.Platform = "iOS"
		if m := iosVersionRe.FindStringSubmatch(ua); m != nil {
			result.PlatformVersion = strings.ReplaceAll(m[1], "_", ".")
		}
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipod"):
		result.DeviceType = "mobile"
		result.Platform = "iOS"
		if m := iosVersionRe.FindStringSubmatch(ua); m != nil {
			result.PlatformVersion = strings.ReplaceAll(m[1], "_", ".")
		}
	case strings.Contains(lower, "android"):
		result.Platform = "Android"
		if strings.Contains(lower, "mobile") {
			result.DeviceType = "mobile"
		} else {
			result.DeviceType = "tablet"
		}
		if m := androidVersionRe.FindStringSubmatch(ua); m != nil {
			result.PlatformVersion = m[1]
		}
	case strings.Contains(lower, "windows"):
		result.Platform = "Windows"
		if m := windowsVersionRe.FindStringSubmatch(ua); m != nil {
			result.PlatformVersion = m[1]
		}
	case strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os x"):
		result.Platform = "Mac OS"
		if m := macVersionRe.FindStringSubmatch(ua); m != nil {
			result.PlatformVersion = strings.ReplaceAll(m[1], "_", ".")
		}
	case strings.Contains(lower, "linux"):
		result.Platform = "Linux"
	}

	switch {
	case strings.Contains(lower, "edg/") || strings.Contains(lower, "edge/"):
		result.Browser = "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		result.Browser = "Opera"
	case strings.Contains(lower, "chrome/"):
		result.Browser = "Chrome"
	case strings.Contains(lower, "firefox/"):
		result.Browser = "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome"):
		result.Browser = "Safari"
	default:
		result.Browser = "unknown"
	}

	return result
}

var (
	iosInAppPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)GSA/`),
		regexp.MustCompile(`(?i)Gmail/`),
		regexp.MustCompile(`(?i)FBAN|FBAV`),
		regexp.MustCompile(`(?i)Instagram`),
		regexp.MustCompile(`(?i)Twitter`),
		regexp.MustCompile(`(?i)LinkedIn`),
		regexp.MustCompile(`(?i)MicroMessenger`),
		regexp.MustCompile(`(?i)Outlook`),
		regexp.MustCompile(`(?i)YahooMobile`),
	}
	androidInAppPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)FB_IAB|FBAN|FBAV`),
		regexp.MustCompile(`(?i)Instagram`),
		regexp.MustCompile(`(?i)Line/`),
		regexp.MustCompile(`(?i)KAKAOTALK`),
		regexp.MustCompile(`(?i)Twitter`),
		regexp.MustCompile(`(?i)LinkedIn`),
		regexp.MustCompile(`(?i)MicroMessenger`),
		regexp.MustCompile(`(?i)Outlook-Android`),
		regexp.MustCompile(`(?i)WhatsApp`),
		regexp.MustCompile(`(?i)Pinterest`),
		regexp.MustCompile(`(?i)Telegram`),
		regexp.MustCompile(`(?i)Snapchat`),
		regexp.MustCompile(`\swv\)`), // generic Android WebView marker
	}
)

// IsIOSInAppBrowser mirrors isIOSInAppBrowser().
func IsIOSInAppBrowser(userAgent string) bool {
	for _, p := range iosInAppPatterns {
		if p.MatchString(userAgent) {
			return true
		}
	}
	return false
}

func IsAndroidInAppBrowser(userAgent string) bool {
	for _, p := range androidInAppPatterns {
		if p.MatchString(userAgent) {
			return true
		}
	}
	return false
}

type MobileFallback struct {
	URL    string
	Reason string
}

func PickMobileFallbackURL(device DeviceType, userAgent, iosURL, androidURL, webFallbackURL string) *MobileFallback {
	var inApp bool
	var storeURL, storeReason string
	if device == DeviceIOS {
		inApp = IsIOSInAppBrowser(userAgent)
		storeURL, storeReason = iosURL, "ios_app_store_url"
	} else {
		inApp = IsAndroidInAppBrowser(userAgent)
		storeURL, storeReason = androidURL, "android_app_store_url"
	}

	if inApp {
		if webFallbackURL != "" {
			return &MobileFallback{URL: webFallbackURL, Reason: "web_fallback_url"}
		}
		if storeURL != "" {
			return &MobileFallback{URL: storeURL, Reason: storeReason}
		}
	} else {
		if storeURL != "" {
			return &MobileFallback{URL: storeURL, Reason: storeReason}
		}
		if webFallbackURL != "" {
			return &MobileFallback{URL: webFallbackURL, Reason: "web_fallback_url"}
		}
	}
	return nil
}
