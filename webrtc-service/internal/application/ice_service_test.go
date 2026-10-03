package application

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
)

func TestICEServersSTUNOnly(t *testing.T) {
	svc := NewICEService(&config.Config{ICE: config.ICEConfig{
		STUNURLs: []string{"stun:stun.example:3478"},
		TURNURLs: []string{"turn:turn.example:3478"},
	}})

	got := svc.Servers("alice")
	if len(got) != 1 || got[0].URLs[0] != "stun:stun.example:3478" || got[0].Username != "" {
		t.Fatalf("Servers = %+v, want STUN only", got)
	}
}

func TestICEServersTURNCredentials(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	svc := NewICEService(&config.Config{ICE: config.ICEConfig{
		TURNURLs:   []string{"turn:turn.example:3478?transport=udp", "turns:turn.example:5349"},
		TURNSecret: "s3cret",
		TURNTTL:    time.Hour,
	}}).(*ICEService)
	svc.now = func() time.Time { return now }

	got := svc.Servers("alice")
	if len(got) != 1 || len(got[0].URLs) != 2 {
		t.Fatalf("Servers = %+v, want one TURN entry with both URLs", got)
	}
	turn := got[0]
	if turn.Username != "1700003600:alice" {
		t.Fatalf("username = %q, want <now+ttl>:alice", turn.Username)
	}
	mac := hmac.New(sha1.New, []byte("s3cret"))
	mac.Write([]byte(turn.Username))
	if want := base64.StdEncoding.EncodeToString(mac.Sum(nil)); turn.Credential != want {
		t.Fatalf("credential = %q, want %q", turn.Credential, want)
	}

	if anon := svc.Servers("bad\nid")[0].Username; !strings.HasSuffix(anon, ":anonymous") {
		t.Fatalf("invalid user id username = %q, want anonymous", anon)
	}
}
