package http

import (
	"testing"

	"github.com/mileusna/useragent"
)

func TestDetectPlatformAndBots(t *testing.T) {
	cases := []struct {
		name     string
		ua       string
		platform string
		bot      bool
	}{
		{"iphone safari", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1", "ios", false},
		{"ipad desktop mode", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15", "macos", false},
		{"android chrome", "Mozilla/5.0 (Linux; Android 14; SM-S921B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36", "android", false},
		{"zalo in-app android", "Mozilla/5.0 (Linux; Android 13; SM-A536E Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36 Zalo android/12100625 ZaloTheme/light ZaloLanguage/vn", "android", false},
		{"facebook in-app ios", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/460.0.0.36.106;FBBV/580000000;FBDV/iPhone15,2;FBMD/iPhone;FBSN/iOS;FBSV/17.4;FBSS/3;FBID/phone;FBLC/vi_VN;FBOP/5]", "ios", false},
		{"facebook preview", "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)", "", true},
		{"whatsapp preview", "WhatsApp/2.23.20.0", "", true},
		{"telegram preview", "TelegramBot (like TwitterBot)", "", true},
		{"zalo preview", "Mozilla/5.0 (compatible; Zalo-Link-Preview/1.0)", "", true},
		{"empty", "", "", true},
	}
	for _, tc := range cases {
		ua := useragent.Parse(tc.ua)
		if got := isBot(tc.ua, ua); got != tc.bot {
			t.Errorf("%s: isBot = %v, want %v", tc.name, got, tc.bot)
		}
		if tc.platform != "" {
			if got := detectPlatform(ua); got != tc.platform {
				t.Errorf("%s: platform = %s, want %s", tc.name, got, tc.platform)
			}
		}
	}
}
