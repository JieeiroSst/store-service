package application

import "testing"

func TestNewShipIDIsUniqueUnderBurst(t *testing.T) {
	seen := make(map[int]struct{}, 10000)
	for i := 0; i < 10000; i++ {
		id := newShipID()
		if id <= 0 {
			t.Fatalf("newShipID() = %d, want positive", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate ship id %d after %d calls", id, i)
		}
		seen[id] = struct{}{}
	}
}
