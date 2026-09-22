package nfcreader

import (
	"bytes"
	"testing"
)

// encodeTLV is a minimal test-only BER-TLV encoder (short and long-form
// length, single or multi-byte tag) used to build fixtures for the
// decoder under test.
func encodeTLV(tag uint32, value []byte) []byte {
	var out []byte
	out = append(out, encodeTag(tag)...)
	out = append(out, encodeLength(len(value))...)
	out = append(out, value...)
	return out
}

func encodeTag(tag uint32) []byte {
	if tag <= 0xFF {
		return []byte{byte(tag)}
	}
	var b []byte
	for tag > 0 {
		b = append([]byte{byte(tag)}, b...)
		tag >>= 8
	}
	return b
}

func encodeLength(n int) []byte {
	if n < 128 {
		return []byte{byte(n)}
	}
	var lb []byte
	for n > 0 {
		lb = append([]byte{byte(n)}, lb...)
		n >>= 8
	}
	return append([]byte{0x80 | byte(len(lb))}, lb...)
}

func TestParseTLV_Primitive(t *testing.T) {
	data := encodeTLV(0x5F01, []byte("0107"))
	node, consumed, err := ParseTLV(data)
	if err != nil {
		t.Fatalf("ParseTLV: %v", err)
	}
	if consumed != len(data) {
		t.Errorf("consumed = %d, want %d", consumed, len(data))
	}
	if node.Tag != 0x5F01 {
		t.Errorf("Tag = %#x, want %#x", node.Tag, 0x5F01)
	}
	if string(node.Value) != "0107" {
		t.Errorf("Value = %q, want %q", node.Value, "0107")
	}
	if node.IsConstructed() {
		t.Errorf("expected primitive tag, got constructed")
	}
}

func TestParseTLV_ConstructedNested(t *testing.T) {
	inner := encodeTLV(0x5F01, []byte("0107"))
	outer := encodeTLV(0x60, inner) // 0x60 has the constructed bit (0x20) set

	node, _, err := ParseTLV(outer)
	if err != nil {
		t.Fatalf("ParseTLV: %v", err)
	}
	if !node.IsConstructed() {
		t.Fatalf("expected constructed tag")
	}
	child, ok := node.Find(0x5F01)
	if !ok {
		t.Fatalf("expected child tag %#x", 0x5F01)
	}
	if string(child.Value) != "0107" {
		t.Errorf("child Value = %q, want %q", child.Value, "0107")
	}
}

func TestParseTLV_LongFormLength(t *testing.T) {
	value := bytes.Repeat([]byte{0xAB}, 200) // forces long-form length (>127)
	data := encodeTLV(0x53, value)

	node, consumed, err := ParseTLV(data)
	if err != nil {
		t.Fatalf("ParseTLV: %v", err)
	}
	if consumed != len(data) {
		t.Errorf("consumed = %d, want %d", consumed, len(data))
	}
	if len(node.Value) != 200 {
		t.Errorf("len(Value) = %d, want 200", len(node.Value))
	}
}

func TestParseTLV_TruncatedLength(t *testing.T) {
	if _, _, err := ParseTLV([]byte{0x60, 0x05, 0x01, 0x02}); err == nil {
		t.Fatalf("expected error for a length exceeding available data")
	}
}
