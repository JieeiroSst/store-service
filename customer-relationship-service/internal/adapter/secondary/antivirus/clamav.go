package antivirus

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"go.uber.org/fx"
)

const chunk = 64 << 10

type clamav struct {
	addr    string
	timeout time.Duration
}

func New(cfg *config.Config) port.Scanner {
	fc := cfg.Files
	if fc.ScanAddress == "" {
		return disabled{}
	}
	timeout := fc.ScanTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &clamav{addr: fc.ScanAddress, timeout: timeout}
}

func (c *clamav) Enabled() bool { return true }

func (c *clamav) Scan(ctx context.Context, _ string, data []byte) (port.ScanResult, error) {
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return port.ScanResult{}, fmt.Errorf("clamd: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()

	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return port.ScanResult{}, fmt.Errorf("clamd: %w", err)
	}
	var size [4]byte
	for off := 0; off < len(data); off += chunk {
		end := min(off+chunk, len(data))
		binary.BigEndian.PutUint32(size[:], uint32(end-off))
		if _, err := conn.Write(size[:]); err != nil {
			return port.ScanResult{}, fmt.Errorf("clamd: %w", err)
		}
		if _, err := conn.Write(data[off:end]); err != nil {
			return port.ScanResult{}, fmt.Errorf("clamd: %w", err)
		}
	}
	binary.BigEndian.PutUint32(size[:], 0)
	if _, err := conn.Write(size[:]); err != nil {
		return port.ScanResult{}, fmt.Errorf("clamd: %w", err)
	}

	reply, err := bufio.NewReader(conn).ReadString(0)
	if err != nil && reply == "" {
		return port.ScanResult{}, fmt.Errorf("clamd: %w", err)
	}
	return parse(reply)
}

func parse(reply string) (port.ScanResult, error) {
	reply = strings.TrimSpace(strings.TrimRight(reply, "\x00"))
	switch {
	case strings.HasSuffix(reply, " OK"):
		return port.ScanResult{Clean: true, Engine: "clamav"}, nil
	case strings.HasSuffix(reply, " FOUND"):
		sig := strings.TrimSuffix(strings.TrimPrefix(reply, "stream:"), " FOUND")
		return port.ScanResult{Signature: strings.TrimSpace(sig), Engine: "clamav"}, nil
	}
	return port.ScanResult{}, fmt.Errorf("clamd: unexpected reply %q", reply)
}

type disabled struct{}

func (disabled) Enabled() bool { return false }
func (disabled) Scan(context.Context, string, []byte) (port.ScanResult, error) {
	return port.ScanResult{}, fmt.Errorf("virus scanning is not configured")
}

var Module = fx.Options(fx.Provide(New))
