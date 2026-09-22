package mrz

import "strings"

func CheckDigit(s string) int {
	weights := [3]int{7, 3, 1}
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += CharValue(rune(s[i])) * weights[i%3]
	}
	return sum % 10
}

func CharValue(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0')
	case r >= 'A' && r <= 'Z':
		return int(r-'A') + 10
	default:
		return 0
	}
}

func CheckDigitMatches(field, check string) bool {
	if len(check) != 1 || check[0] < '0' || check[0] > '9' {
		return false
	}
	want := int(check[0] - '0')
	return CheckDigit(field) == want
}

type TD1Fields struct {
	DocumentNumber string
	Surname        string
	GivenNames     string
	Nationality    string
	DateOfBirth    string
	Sex            string
	DateOfExpiry   string
	DocNumberValid bool
	DOBValid       bool
	ExpiryValid    bool
	CompositeValid bool
}

func ParseTD1(line1, line2, line3 string) TD1Fields {
	line1 = padTo(line1, 30)
	line2 = padTo(line2, 30)
	line3 = padTo(line3, 30)

	// Line 1: doc type(2) + issuing country(3) + doc number(9) + check(1) + optional(15)
	docNumberField := line1[5:14]
	docNumberCheck := line1[14:15]
	optional1 := line1[15:30]

	// Line 2: DOB(6) + check(1) + sex(1) + expiry(6) + check(1) + nationality(3) + optional(11) + composite check(1)
	dob := line2[0:6]
	dobCheck := line2[6:7]
	sex := line2[7:8]
	expiry := line2[8:14]
	expiryCheck := line2[14:15]
	nationality := line2[15:18]
	optional2 := line2[18:29]
	compositeCheck := line2[29:30]

	surname, givenNames := parseName(line3)

	f := TD1Fields{
		DocumentNumber: trimFiller(docNumberField),
		Surname:        surname,
		GivenNames:     givenNames,
		Nationality:    trimFiller(nationality),
		DateOfBirth:    dob,
		Sex:            sex,
		DateOfExpiry:   expiry,
	}

	f.DocNumberValid = CheckDigitMatches(docNumberField, docNumberCheck)
	f.DOBValid = CheckDigitMatches(dob, dobCheck)
	f.ExpiryValid = CheckDigitMatches(expiry, expiryCheck)

	composite := docNumberField + docNumberCheck + optional1 + dob + dobCheck + expiry + expiryCheck + optional2
	f.CompositeValid = CheckDigitMatches(composite, compositeCheck)

	return f
}

func parseName(line3 string) (surname, givenNames string) {
	parts := strings.SplitN(line3, "<<", 2)
	surname = trimFiller(strings.ReplaceAll(parts[0], "<", " "))
	if len(parts) == 2 {
		givenNames = trimFiller(strings.ReplaceAll(parts[1], "<", " "))
	}
	return strings.TrimSpace(surname), strings.TrimSpace(givenNames)
}

func trimFiller(s string) string {
	return strings.Trim(s, "< ")
}

func padTo(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat("<", n-len(s))
}
