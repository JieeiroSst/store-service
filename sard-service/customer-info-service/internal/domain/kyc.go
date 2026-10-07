package domain

import (
	"strconv"
	"time"
)

type KYCStatus string

const (
	KYCNone           KYCStatus = "NONE"
	KYCPending        KYCStatus = "PENDING"
	KYCVerified       KYCStatus = "VERIFIED"
	KYCRejected       KYCStatus = "REJECTED"
	KYCReviewRequired KYCStatus = "REVIEW_REQUIRED"
)

const (
	DocumentCCCD     = "CCCD"
	MaxDocumentImage = 10 << 20
	dateLayout       = "02/01/2006"
	minPlausibleYear = 1900
)

type IdentityDocument struct {
	Source        string
	Number        string
	FullName      string
	DateOfBirth   *time.Time
	Gender        string
	Nationality   string
	ExpiryDate    *time.Time
	ChecksumValid bool
	NFCVerified   bool
	Confidence    float64
}

func (d IdentityDocument) Expired(now time.Time) bool {
	return d.ExpiryDate != nil && !now.Before(d.ExpiryDate.AddDate(0, 0, 1))
}

func (d IdentityDocument) Last4() string {
	if len(d.Number) < 4 {
		return d.Number
	}
	return d.Number[len(d.Number)-4:]
}

type EkycResult struct {
	Document     *IdentityDocument
	FaceVerified bool
	FaceScore    float64
}

type KYC struct {
	Status       KYCStatus
	DocumentType string
	DocumentHash string
	Document     IdentityDocument
	FullName     string
	FaceVerified bool
	Reasons      []string
	VerifiedAt   *time.Time
	SubmittedAt  *time.Time
}

type KYCPolicy struct {
	MinAge           int
	MinConfidence    float64
	RequireFaceMatch bool
}

func (p KYCPolicy) Evaluate(c *Customer, res EkycResult, now time.Time) KYC {
	submitted := now
	if res.Document == nil {
		return KYC{Status: KYCNone, Reasons: []string{"no identity document submitted to ekyc-service"}, SubmittedAt: &submitted}
	}
	doc := *res.Document
	var reasons []string
	if n := len(doc.Number); (n != 9 && n != 12) || !isDigits(doc.Number) {
		reasons = append(reasons, "cccd number is missing or malformed")
	}
	if !doc.ChecksumValid && !doc.NFCVerified {
		reasons = append(reasons, "mrz check digits do not match")
	}
	switch {
	case doc.FullName == "":
		reasons = append(reasons, "full name not found on the document")
	case NormalizeName(c.FullName) == "":
		reasons = append(reasons, "user profile has no name to compare with")
	case NormalizeName(doc.FullName) != NormalizeName(c.FullName):
		reasons = append(reasons, "name on the document does not match the user profile")
	}
	switch {
	case doc.DateOfBirth == nil:
		reasons = append(reasons, "date of birth not found on the document")
	case doc.DateOfBirth.Year() < minPlausibleYear || doc.DateOfBirth.After(now):
		reasons = append(reasons, "date of birth is not plausible")
	case doc.DateOfBirth.AddDate(p.MinAge, 0, 0).After(now):
		reasons = append(reasons, "customer is younger than the minimum age")
	}
	if doc.Expired(now) {
		reasons = append(reasons, "identity document expired")
	}
	if !doc.NFCVerified && doc.Confidence < p.MinConfidence {
		reasons = append(reasons, "document scan confidence is too low")
	}

	k := KYC{
		DocumentType: DocumentCCCD,
		Document:     doc,
		FullName:     doc.FullName,
		FaceVerified: res.FaceVerified,
		Reasons:      reasons,
		SubmittedAt:  &submitted,
	}
	switch {
	case len(reasons) > 0:
		k.Status = KYCRejected
	case p.RequireFaceMatch && !res.FaceVerified:
		k.Status = KYCPending
		k.Reasons = []string{"face verification in ekyc-service is not completed"}
	default:
		k.Status = KYCVerified
		k.VerifiedAt = &submitted
	}
	return k
}

func ParseMRZDate(yymmdd string, now time.Time, future bool) *time.Time {
	if len(yymmdd) != 6 || !isDigits(yymmdd) {
		return nil
	}
	yy, _ := strconv.Atoi(yymmdd[:2])
	mm, _ := strconv.Atoi(yymmdd[2:4])
	dd, _ := strconv.Atoi(yymmdd[4:])
	year := 2000 + yy
	if !future && year > now.Year() {
		year -= 100
	}
	t := time.Date(year, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
	if t.Month() != time.Month(mm) || t.Day() != dd {
		return nil
	}
	return &t
}

func FormatDocumentDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateLayout)
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
