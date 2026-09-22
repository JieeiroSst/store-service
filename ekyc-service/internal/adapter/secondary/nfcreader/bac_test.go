package nfcreader

import (
	"encoding/hex"
	"testing"
)

func TestDeriveBACKeys(t *testing.T) {
	// Cross-checked against an independent Python (hashlib.sha1-based)
	// implementation of the same ICAO 9303 Appendix D algorithm for this
	// input - see the derivation script used during development.
	kEnc, kMac, err := DeriveBACKeys("C1234567X", "900115", "300115")
	if err != nil {
		t.Fatalf("DeriveBACKeys: %v", err)
	}

	wantEnc := "9ea438f198f7b964dc4adf58f1136bb0"
	wantMac := "d631e692b38632dc85e37a0d1673f7f1"

	if got := hex.EncodeToString(kEnc); got != wantEnc {
		t.Errorf("KEnc = %s, want %s", got, wantEnc)
	}
	if got := hex.EncodeToString(kMac); got != wantMac {
		t.Errorf("KMac = %s, want %s", got, wantMac)
	}
}

func TestDeriveBACKeys_Properties(t *testing.T) {
	kEnc, kMac, err := DeriveBACKeys("L898902C3", "690806", "940623")
	if err != nil {
		t.Fatalf("DeriveBACKeys: %v", err)
	}

	for name, key := range map[string][]byte{"KEnc": kEnc, "KMac": kMac} {
		if len(key) != 16 {
			t.Errorf("%s length = %d, want 16 (two-key 3DES)", name, len(key))
		}
		for i, b := range key {
			ones := 0
			for v := b; v != 0; v &= v - 1 {
				ones++
			}
			if ones%2 != 1 {
				t.Errorf("%s byte %d (%#02x) does not have odd DES parity", name, i, b)
			}
		}
	}

	kEnc2, _, err := DeriveBACKeys("L898902C3", "690806", "940623")
	if err != nil {
		t.Fatalf("DeriveBACKeys (2nd call): %v", err)
	}
	if hex.EncodeToString(kEnc) != hex.EncodeToString(kEnc2) {
		t.Errorf("DeriveBACKeys is not deterministic")
	}
}
