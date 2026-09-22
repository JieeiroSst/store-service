package nfcreader

import (
	"reflect"
	"testing"
)

func TestParseEFCOM(t *testing.T) {
	inner := append(encodeTLV(tagLDSVersion, []byte("0107")), encodeTLV(tagUnicodeVer, []byte("040000"))...)
	inner = append(inner, encodeTLV(tagDGTagList, []byte{0x61, 0x75})...) // DG1, DG2
	data := encodeTLV(tagEFCOM, inner)

	com, err := ParseEFCOM(data)
	if err != nil {
		t.Fatalf("ParseEFCOM: %v", err)
	}
	if com.LDSVersion != "0107" {
		t.Errorf("LDSVersion = %q, want %q", com.LDSVersion, "0107")
	}
	if com.UnicodeVersion != "040000" {
		t.Errorf("UnicodeVersion = %q, want %q", com.UnicodeVersion, "040000")
	}
	if want := []int{1, 2}; !reflect.DeepEqual(com.PresentDataGroups, want) {
		t.Errorf("PresentDataGroups = %v, want %v", com.PresentDataGroups, want)
	}
}

func TestParseEFCOM_WrongOuterTag(t *testing.T) {
	data := encodeTLV(tagDG1, []byte("not EF.COM"))
	if _, err := ParseEFCOM(data); err == nil {
		t.Fatalf("expected error for wrong outer tag")
	}
}
