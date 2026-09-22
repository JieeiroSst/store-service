package nfcreader

import (
	"strings"
	"testing"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/mrz"
)

func buildTestTD1MRZ() string {
	docNumberField := "C1234567X"
	docCheck := testDigit(mrz.CheckDigit(docNumberField))
	optional1 := testPadTo("", 15)

	dob := "900115"
	dobCheck := testDigit(mrz.CheckDigit(dob))
	sex := "M"
	expiry := "300115"
	expiryCheck := testDigit(mrz.CheckDigit(expiry))
	nationality := "VNM"
	optional2 := testPadTo("", 11)

	composite := docNumberField + docCheck + optional1 + dob + dobCheck + expiry + expiryCheck + optional2
	compositeCheck := testDigit(mrz.CheckDigit(composite))

	line1 := "ID" + "VNM" + docNumberField + docCheck + optional1
	line2 := dob + dobCheck + sex + expiry + expiryCheck + nationality + optional2 + compositeCheck
	line3 := testPadTo("NGUYEN<<VAN<A", 30)
	return line1 + line2 + line3
}

func testDigit(n int) string {
	return string(rune('0' + n))
}

func testPadTo(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat("<", n-len(s))
}

func TestParseDG1(t *testing.T) {
	mrzText := buildTestTD1MRZ()
	data := encodeTLV(tagDG1, encodeTLV(tagMRZData, []byte(mrzText)))

	fields, rawText, err := ParseDG1(data)
	if err != nil {
		t.Fatalf("ParseDG1: %v", err)
	}
	if rawText != mrzText {
		t.Errorf("raw MRZ text mismatch")
	}
	if fields.DocumentNumber != "C1234567X" {
		t.Errorf("DocumentNumber = %q, want C1234567X", fields.DocumentNumber)
	}
	if fields.Surname != "NGUYEN" || fields.GivenNames != "VAN A" {
		t.Errorf("name = %q / %q, want NGUYEN / VAN A", fields.Surname, fields.GivenNames)
	}
	if !fields.DocNumberValid || !fields.DOBValid || !fields.ExpiryValid || !fields.CompositeValid {
		t.Errorf("expected all checksums valid, got %+v", fields)
	}
}

func TestParseDG1_WrongLength(t *testing.T) {
	data := encodeTLV(tagDG1, encodeTLV(tagMRZData, []byte("too short")))
	if _, _, err := ParseDG1(data); err == nil {
		t.Fatalf("expected error for non-90-character MRZ data")
	}
}
