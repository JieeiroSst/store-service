package mrz

import "testing"

func TestCheckDigit_ICAOTestVector(t *testing.T) {
	// ICAO 9303 Part 3 worked example (passport number field): the field
	// "AB2134<<<" with weights 7-3-1 has a documented check digit of 5.
	got := CheckDigit("AB2134<<<")
	if got != 5 {
		t.Fatalf("CheckDigit(AB2134<<<) = %d, want 5", got)
	}
}

func TestCheckDigitMatches(t *testing.T) {
	field := "AB2134<<<"
	check := "5"
	if !CheckDigitMatches(field, check) {
		t.Fatalf("expected CheckDigitMatches(%q, %q) = true", field, check)
	}
	if CheckDigitMatches(field, "9") {
		t.Fatalf("expected CheckDigitMatches(%q, %q) = false", field, "9")
	}
}

func TestParseTD1(t *testing.T) {
	// Build a synthetic but well-formed TD1 MRZ for a fictional card and
	// verify the parser slices fields and validates checksums correctly.
	docNumberField := "C1234567X" // 9 chars
	docCheck := digit(CheckDigit(docNumberField))
	optional1 := padTo("", 15)

	dob := "900115" // 1990-01-15
	dobCheck := digit(CheckDigit(dob))
	sex := "M"
	expiry := "300115"
	expiryCheck := digit(CheckDigit(expiry))
	nationality := "VNM"
	optional2 := padTo("", 11)

	composite := docNumberField + docCheck + optional1 + dob + dobCheck + expiry + expiryCheck + optional2
	compositeCheck := digit(CheckDigit(composite))

	line1 := "ID" + "VNM" + docNumberField + docCheck + optional1
	line2 := dob + dobCheck + sex + expiry + expiryCheck + nationality + optional2 + compositeCheck
	line3 := padTo("NGUYEN<<VAN<A", 30)

	if len(line1) != 30 || len(line2) != 30 {
		t.Fatalf("test fixture malformed: len(line1)=%d len(line2)=%d", len(line1), len(line2))
	}

	fields := ParseTD1(line1, line2, line3)

	if fields.DocumentNumber != "C1234567X" {
		t.Errorf("DocumentNumber = %q, want C1234567X", fields.DocumentNumber)
	}
	if fields.Nationality != "VNM" {
		t.Errorf("Nationality = %q, want VNM", fields.Nationality)
	}
	if fields.Surname != "NGUYEN" || fields.GivenNames != "VAN A" {
		t.Errorf("name = %q / %q, want NGUYEN / VAN A", fields.Surname, fields.GivenNames)
	}
	if !fields.DocNumberValid || !fields.DOBValid || !fields.ExpiryValid || !fields.CompositeValid {
		t.Errorf("expected all checksums valid, got %+v", fields)
	}
}

func digit(n int) string {
	return string(rune('0' + n))
}
