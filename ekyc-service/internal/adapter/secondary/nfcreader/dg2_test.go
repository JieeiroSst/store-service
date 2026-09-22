package nfcreader

import (
	"bytes"
	"testing"
)

func TestParseDG2Photo_ViaTLVNavigation(t *testing.T) {
	fakeCBEFFHeader := []byte("FAC\x00010\x00") // stand-in for the ISO 19794-5 record header this package doesn't parse
	fakeJPEG := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, []byte("...jpeg bytes...")...)
	dataBlock := append(append([]byte{}, fakeCBEFFHeader...), fakeJPEG...)

	bit := encodeTLV(tagBioDataBlock, dataBlock)
	bitTemplate := encodeTLV(tagBIOInfoTmpl, bit)
	bitGroup := encodeTLV(tagBIOInfoGroup, append(encodeTLV(0x02, []byte{0x01}), bitTemplate...))
	dg2 := encodeTLV(tagDG2, bitGroup)

	photo, err := ParseDG2Photo(dg2)
	if err != nil {
		t.Fatalf("ParseDG2Photo: %v", err)
	}
	if !bytes.Equal(photo, fakeJPEG) {
		t.Errorf("extracted photo = %x, want %x", photo, fakeJPEG)
	}
}

func TestParseDG2Photo_FallbackScan(t *testing.T) {
	// Malformed/unexpected outer structure - not a valid DG2 TLV wrapper -
	// but the JPEG marker is still findable by scanning raw bytes.
	fakeJPEG := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, []byte("...jpeg bytes...")...)
	garbage := append([]byte{0x01, 0x02, 0x03}, fakeJPEG...)

	photo, err := ParseDG2Photo(garbage)
	if err != nil {
		t.Fatalf("ParseDG2Photo: %v", err)
	}
	if !bytes.Equal(photo, fakeJPEG) {
		t.Errorf("extracted photo = %x, want %x", photo, fakeJPEG)
	}
}

func TestParseDG2Photo_NoImage(t *testing.T) {
	if _, err := ParseDG2Photo([]byte{0x01, 0x02, 0x03}); err == nil {
		t.Fatalf("expected error when no image marker is present")
	}
}
