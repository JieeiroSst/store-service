package domain

import (
	"strings"
	"testing"
)

func TestUTMNormalizeAndDefaults(t *testing.T) {
	got := UTM{Source: "  Facebook ", Campaign: "Sale_Tet_2026", Content: " Banner A "}.Normalize().
		WithDefaults(UTM{Source: "zalo", Medium: "referral", Campaign: "default", Term: "kw"})

	want := UTM{Source: "facebook", Medium: "referral", Campaign: "sale_tet_2026", Term: "kw", Content: "Banner A"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestUTMNormalizeTruncates(t *testing.T) {
	got := UTM{Campaign: strings.Repeat("a", 300)}.Normalize()
	if len(got.Campaign) != utmMaxLen {
		t.Fatalf("campaign length %d, want %d", len(got.Campaign), utmMaxLen)
	}
	if !(UTM{}).IsZero() || got.IsZero() {
		t.Fatal("IsZero mismatch")
	}
}

func TestSplitRefCode(t *testing.T) {
	cases := map[string][2]string{
		"ABC12345":           {"ABC12345", ""},
		"REF:ABC12345":       {"ABC12345", ""},
		"ABC12345:click-1":   {"ABC12345", "click-1"},
		" REF:ABC12345:c-2 ": {"ABC12345", "c-2"},
	}
	for in, want := range cases {
		code, click := SplitRefCode(in)
		if code != want[0] || click != want[1] {
			t.Errorf("SplitRefCode(%q) = %q, %q; want %q, %q", in, code, click, want[0], want[1])
		}
	}
}
