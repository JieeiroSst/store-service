package domain

import (
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type CustomerStatus string

const (
	CustomerActive    CustomerStatus = "ACTIVE"
	CustomerSuspended CustomerStatus = "SUSPENDED"
)

type UserProfile struct {
	ID       int64
	Username string
	Email    string
	Name     string
	Phone    string
	Address  string
	Sex      string
	Active   bool
}

type Customer struct {
	ID        string
	UserID    int64
	Username  string
	FullName  string
	Email     string
	Phone     string
	Address   string
	Gender    string
	Status    CustomerStatus
	KYC       KYC
	CreatedAt time.Time
	UpdatedAt time.Time
	SyncedAt  time.Time
}

func NewCustomer(id string, u UserProfile, now time.Time) (*Customer, error) {
	if !u.Active {
		return nil, Conflict("user %d is locked in user-service", u.ID)
	}
	c := &Customer{
		ID:        id,
		UserID:    u.ID,
		Status:    CustomerActive,
		KYC:       KYC{Status: KYCNone},
		CreatedAt: now,
	}
	c.applyProfile(u, now)
	return c, nil
}

func (c *Customer) Sync(u UserProfile, now time.Time) {
	nameChanged := NormalizeName(u.Name) != NormalizeName(c.FullName)
	c.applyProfile(u, now)
	switch {
	case !u.Active:
		c.Status = CustomerSuspended
	case c.Status == CustomerSuspended:
		c.Status = CustomerActive
	}
	if nameChanged && c.KYC.Status == KYCVerified && NormalizeName(c.KYC.FullName) != NormalizeName(c.FullName) {
		c.KYC.Status = KYCReviewRequired
		c.KYC.Reasons = []string{"profile name no longer matches the identity document"}
	}
}

func (c *Customer) applyProfile(u UserProfile, now time.Time) {
	c.Username = u.Username
	c.FullName = strings.TrimSpace(u.Name)
	c.Email = u.Email
	c.Phone = u.Phone
	c.Address = u.Address
	c.Gender = u.Sex
	c.UpdatedAt = now
	c.SyncedAt = now
}

func (c *Customer) Clone() *Customer {
	cp := *c
	cp.KYC.Reasons = append([]string(nil), c.KYC.Reasons...)
	if c.KYC.VerifiedAt != nil {
		t := *c.KYC.VerifiedAt
		cp.KYC.VerifiedAt = &t
	}
	return &cp
}

type Eligibility struct {
	Eligible bool
	Reasons  []string
}

func (c *Customer) CardEligibility(now time.Time) Eligibility {
	var reasons []string
	if c.Status != CustomerActive {
		reasons = append(reasons, "customer is "+string(c.Status))
	}
	if c.KYC.Status != KYCVerified {
		reasons = append(reasons, "kyc is "+string(c.KYC.Status))
	} else if c.KYC.Document.Expired(now) {
		reasons = append(reasons, "identity document expired")
	}
	return Eligibility{Eligible: len(reasons) == 0, Reasons: reasons}
}

func NormalizeName(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case r == 'đ' || r == 'Đ':
			r = 'D'
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
