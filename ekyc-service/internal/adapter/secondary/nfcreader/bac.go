package nfcreader

import (
	"crypto/sha1"
	"fmt"
	"math/bits"
	"strings"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/mrz"
)

func DeriveBACKeys(documentNumber, dateOfBirth, dateOfExpiry string) (kEnc, kMac []byte, err error) {
	if len(dateOfBirth) != 6 || len(dateOfExpiry) != 6 {
		return nil, nil, fmt.Errorf("nfcreader: dateOfBirth/dateOfExpiry must be 6-digit YYMMDD")
	}

	docField := padMRZField(documentNumber, 9)
	docCheck := mrz.CheckDigit(docField)
	dobCheck := mrz.CheckDigit(dateOfBirth)
	expCheck := mrz.CheckDigit(dateOfExpiry)

	mrzInformation := fmt.Sprintf("%s%d%s%d%s%d", docField, docCheck, dateOfBirth, dobCheck, dateOfExpiry, expCheck)

	seedHash := sha1.Sum([]byte(mrzInformation))
	kSeed := seedHash[:16]

	kEnc = deriveDESKey(kSeed, 1)
	kMac = deriveDESKey(kSeed, 2)
	return kEnc, kMac, nil
}

func deriveDESKey(kSeed []byte, counter uint32) []byte {
	d := make([]byte, 0, len(kSeed)+4)
	d = append(d, kSeed...)
	d = append(d, byte(counter>>24), byte(counter>>16), byte(counter>>8), byte(counter))

	h := sha1.Sum(d)
	ka := adjustDESParity(h[0:8])
	kb := adjustDESParity(h[8:16])
	return append(ka, kb...)
}

func adjustDESParity(in []byte) []byte {
	out := make([]byte, len(in))
	for i, v := range in {
		v &^= 1
		if bits.OnesCount8(v)%2 == 0 {
			v |= 1
		}
		out[i] = v
	}
	return out
}

func padMRZField(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat("<", n-len(s))
}
