package domain

import (
	"testing"
	"time"
)

var now = time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)

func date(y, m, d int) *time.Time {
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return &t
}

func customer(t *testing.T) *Customer {
	t.Helper()
	c, err := NewCustomer("c1", UserProfile{ID: 7, Name: "Nguyễn Văn An", Active: true}, now)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func goodDoc() IdentityDocument {
	return IdentityDocument{
		Source: "card_ocr", Number: "079200012345", FullName: "NGUYEN VAN AN", DateOfBirth: date(2000, 5, 1),
		ExpiryDate: date(2040, 5, 1), ChecksumValid: true, Confidence: 0.9,
	}
}

func verified(d IdentityDocument) EkycResult { return EkycResult{Document: &d, FaceVerified: true} }

func TestNewCustomerRejectsLockedUser(t *testing.T) {
	if _, err := NewCustomer("c", UserProfile{ID: 1}, now); !IsConflict(err) {
		t.Fatalf("got %v", err)
	}
}

func TestEvaluateKYC(t *testing.T) {
	policy := KYCPolicy{MinAge: 18, MinConfidence: 0.5, RequireFaceMatch: true}
	c := customer(t)
	k := policy.Evaluate(c, verified(goodDoc()), now)
	if k.Status != KYCVerified || k.VerifiedAt == nil || len(k.Reasons) != 0 {
		t.Fatalf("expected verified: %+v", k)
	}
	c.KYC = k
	if e := c.CardEligibility(now); !e.Eligible {
		t.Fatalf("eligible: %+v", e)
	}
	if e := c.CardEligibility(time.Date(2040, 5, 2, 0, 0, 0, 0, time.UTC)); e.Eligible {
		t.Fatal("expired document must not be eligible")
	}

	cases := map[string]func(*IdentityDocument){
		"short number":   func(d *IdentityDocument) { d.Number = "12345" },
		"bad checksum":   func(d *IdentityDocument) { d.ChecksumValid = false },
		"name mismatch":  func(d *IdentityDocument) { d.FullName = "TRAN VAN BINH" },
		"no dob":         func(d *IdentityDocument) { d.DateOfBirth = nil },
		"minor":          func(d *IdentityDocument) { d.DateOfBirth = date(2010, 1, 1) },
		"future dob":     func(d *IdentityDocument) { d.DateOfBirth = date(2030, 1, 1) },
		"expired":        func(d *IdentityDocument) { d.ExpiryDate = date(2026, 10, 6) },
		"low confidence": func(d *IdentityDocument) { d.Confidence = 0.2 },
	}
	for name, mut := range cases {
		d := goodDoc()
		mut(&d)
		if k := policy.Evaluate(customer(t), verified(d), now); k.Status != KYCRejected || len(k.Reasons) == 0 {
			t.Errorf("%s: %+v", name, k)
		}
	}
	d := goodDoc()
	d.ExpiryDate = date(2026, 10, 7)
	if k := policy.Evaluate(customer(t), verified(d), now); k.Status != KYCVerified {
		t.Fatalf("document is valid through its expiry day: %+v", k)
	}
	if k := policy.Evaluate(customer(t), EkycResult{Document: &d}, now); k.Status != KYCPending || k.VerifiedAt != nil {
		t.Fatalf("face match still missing: %+v", k)
	}
	if k := (KYCPolicy{MinAge: 18, MinConfidence: 0.5}).Evaluate(customer(t), EkycResult{Document: &d}, now); k.Status != KYCVerified {
		t.Fatalf("face match not required: %+v", k)
	}
	if k := policy.Evaluate(customer(t), EkycResult{}, now); k.Status != KYCNone {
		t.Fatalf("nothing submitted: %+v", k)
	}
	nfc := goodDoc()
	nfc.Source, nfc.ChecksumValid, nfc.NFCVerified, nfc.Confidence = "nfc_chip", false, true, 0
	if k := policy.Evaluate(customer(t), verified(nfc), now); k.Status != KYCVerified {
		t.Fatalf("nfc chip data is trusted without ocr confidence: %+v", k)
	}
}

func TestSync(t *testing.T) {
	c := customer(t)
	c.KYC = KYCPolicy{MinAge: 18, MinConfidence: 0.5}.Evaluate(c, verified(goodDoc()), now)
	later := now.Add(time.Hour)

	c.Sync(UserProfile{ID: 7, Name: "nguyen van an", Phone: "0901", Active: true}, later)
	if c.KYC.Status != KYCVerified || c.Phone != "0901" || !c.SyncedAt.Equal(later) {
		t.Fatalf("same name keeps kyc: %+v", c)
	}
	c.Sync(UserProfile{ID: 7, Name: "Nguyen Van Binh", Active: false}, later)
	if c.Status != CustomerSuspended || c.KYC.Status != KYCReviewRequired {
		t.Fatalf("locked user and new name: %+v", c)
	}
	c.Sync(UserProfile{ID: 7, Name: "Nguyen Van Binh", Active: true}, later)
	if c.Status != CustomerActive {
		t.Fatalf("unlocked user: %s", c.Status)
	}
	if e := c.CardEligibility(later); e.Eligible {
		t.Fatal("review required must not be eligible")
	}
}

func TestNormalizeName(t *testing.T) {
	if got := NormalizeName("  Đặng   thị  Hồng "); got != "DANG THI HONG" {
		t.Fatalf("got %q", got)
	}
}

func TestParseMRZDate(t *testing.T) {
	if got := FormatDocumentDate(ParseMRZDate("000501", now, false)); got != "01/05/2000" {
		t.Fatalf("dob 2000: %s", got)
	}
	if got := FormatDocumentDate(ParseMRZDate("850312", now, false)); got != "12/03/1985" {
		t.Fatalf("dob 1985: %s", got)
	}
	if got := FormatDocumentDate(ParseMRZDate("400501", now, true)); got != "01/05/2040" {
		t.Fatalf("expiry: %s", got)
	}
	for _, bad := range []string{"", "0005", "001301", "000231", "ab0501"} {
		if ParseMRZDate(bad, now, false) != nil {
			t.Errorf("%q should not parse", bad)
		}
	}
}
