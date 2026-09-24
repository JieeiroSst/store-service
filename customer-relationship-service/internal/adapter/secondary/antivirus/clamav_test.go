package antivirus

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
)

// fakeClamd speaks enough of clamd's INSTREAM protocol to test the client.
type fakeClamd struct {
	addr     string
	received chan []byte
	reply    string
	hang     bool
}

func startClamd(t *testing.T, reply string) *fakeClamd {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeClamd{addr: ln.Addr().String(), received: make(chan []byte, 1), reply: reply}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go f.serve(conn)
		}
	}()
	return f
}

func (f *fakeClamd) serve(conn net.Conn) {
	defer conn.Close()
	cmd := make([]byte, len("zINSTREAM\x00"))
	if _, err := io.ReadFull(conn, cmd); err != nil || string(cmd) != "zINSTREAM\x00" {
		return
	}
	var data []byte
	for {
		var size [4]byte
		if _, err := io.ReadFull(conn, size[:]); err != nil {
			return
		}
		n := binary.BigEndian.Uint32(size[:])
		if n == 0 {
			break
		}
		chunk := make([]byte, n)
		if _, err := io.ReadFull(conn, chunk); err != nil {
			return
		}
		data = append(data, chunk...)
	}
	f.received <- data
	if f.hang {
		time.Sleep(5 * time.Second)
		return
	}
	_, _ = conn.Write([]byte(f.reply + "\x00"))
}

func scanner(addr string, timeout time.Duration) *clamav {
	return New(&config.Config{Files: config.FilesConfig{ScanAddress: addr, ScanTimeout: timeout}}).(*clamav)
}

func TestScan_Clean(t *testing.T) {
	f := startClamd(t, "stream: OK")
	// More than one 64 KB frame, to exercise the framing.
	data := bytes.Repeat([]byte("0123456789"), 20_000)

	res, err := scanner(f.addr, 5*time.Second).Scan(context.Background(), "a.pdf", data)
	if err != nil || !res.Clean || res.Engine != "clamav" {
		t.Fatalf("result %+v, err %v", res, err)
	}
	if got := <-f.received; !bytes.Equal(got, data) {
		t.Fatalf("the daemon received %d bytes, want %d intact", len(got), len(data))
	}
}

func TestScan_Found(t *testing.T) {
	f := startClamd(t, "stream: Win.Test.EICAR_HDB-1 FOUND")
	res, err := scanner(f.addr, 5*time.Second).Scan(context.Background(), "a.pdf", []byte("x"))
	if err != nil || res.Clean || res.Signature != "Win.Test.EICAR_HDB-1" {
		t.Fatalf("result %+v, err %v", res, err)
	}
}

func TestScan_Errors(t *testing.T) {
	for name, reply := range map[string]string{
		"a clamd error":  "INSTREAM size limit exceeded. ERROR",
		"an odd reply":   "what",
		"an empty reply": "",
	} {
		f := startClamd(t, reply)
		if _, err := scanner(f.addr, 5*time.Second).Scan(context.Background(), "a", []byte("x")); err == nil {
			t.Errorf("%s: no error", name)
		}
	}

	// Nothing listening.
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	ln.Close()
	if _, err := scanner(addr, time.Second).Scan(context.Background(), "a", []byte("x")); err == nil {
		t.Error("connection refused: no error")
	}
}

func TestScan_TimesOutAndHonoursCancellation(t *testing.T) {
	f := startClamd(t, "")
	f.hang = true

	start := time.Now()
	if _, err := scanner(f.addr, 300*time.Millisecond).Scan(context.Background(), "a", []byte("x")); err == nil {
		t.Fatal("a daemon that never answers returned no error")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("took %v to give up", d)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(100 * time.Millisecond); cancel() }()
	start = time.Now()
	if _, err := scanner(f.addr, 10*time.Second).Scan(ctx, "a", []byte("x")); err == nil {
		t.Fatal("a cancelled scan returned no error")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("cancellation took %v", d)
	}
}

func TestDisabledWithoutAnAddress(t *testing.T) {
	s := New(&config.Config{})
	if s.Enabled() {
		t.Fatal("enabled without an address")
	}
	if _, err := s.Scan(context.Background(), "a", nil); err == nil {
		t.Fatal("a disabled scanner claimed to scan")
	}
}

func TestParse(t *testing.T) {
	if res, err := parse("stream: OK\x00"); err != nil || !res.Clean {
		t.Fatalf("OK: %+v %v", res, err)
	}
	if res, _ := parse("stream: Some.Virus-1 FOUND\x00"); res.Clean || res.Signature != "Some.Virus-1" {
		t.Fatalf("FOUND: %+v", res)
	}
}

// TestAgainstARealClamd needs a daemon: CLAMAV_TEST_ADDRESS=localhost:3310.
func TestAgainstARealClamd(t *testing.T) {
	addr := os.Getenv("CLAMAV_TEST_ADDRESS")
	if addr == "" {
		t.Skip("CLAMAV_TEST_ADDRESS not set")
	}
	s := scanner(addr, 30*time.Second)
	ctx := context.Background()

	if res, err := s.Scan(ctx, "clean.txt", []byte("just a contract")); err != nil || !res.Clean {
		t.Fatalf("clean file: %+v, %v", res, err)
	}
	// A large file goes through in many frames. (The EICAR signature only
	// matches a file that is the test string itself, so it is not buried here.)
	big := bytes.Repeat([]byte("contract text. "), 200_000)
	if res, err := s.Scan(ctx, "big.txt", big); err != nil || !res.Clean {
		t.Fatalf("big clean file: %+v, %v", res, err)
	}
	// The standard antivirus test string; harmless, every engine flags it.
	eicar := `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`
	res, err := s.Scan(ctx, "eicar.com", []byte(eicar))
	if err != nil || res.Clean || !strings.Contains(strings.ToLower(res.Signature), "eicar") {
		t.Fatalf("EICAR: %+v, %v", res, err)
	}

}
