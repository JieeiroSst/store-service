package application

import (
	"fmt"
	"io"
)

const (
	maxPANAttempts   = 5
	defaultListLimit = 50
	maxListLimit     = 200
)

func accountLockKey(id string) string { return "account:" + id }

func cardLockKey(id string) string { return "card:" + id }

func customerLockKey(id string) string { return "customer:" + id }

func clampLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultListLimit
	case limit > maxListLimit:
		return maxListLimit
	}
	return limit
}

func newID(r io.Reader) (string, error) {
	var b [16]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
