package model

import "time"

type Base struct {
	ID        uint      `json:"id" gorm:"column:id;primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (b *Base) SetID(id uint) { b.ID = id }

type Entity[T any] interface {
	*T
	SetID(id uint)
}

type Campaign struct {
	Base
	Name      string    `json:"campaign_name" gorm:"column:campaign_name;size:255" binding:"required"`
	Objective string    `json:"campaign_objectives" gorm:"column:campaign_objectives;size:255"`
	Sponsor   string    `json:"campaign_sponsor" gorm:"column:campaign_sponsor;size:255"`
	StartDate time.Time `json:"campaign_start_date" gorm:"column:campaign_start_date"`
	EndDate   time.Time `json:"campaign_end_date" gorm:"column:campaign_end_date"`
	Details   string    `json:"campaign_other_details" gorm:"column:campaign_other_details;size:255"`
}

func (Campaign) TableName() string { return "crm_campaign" }

type Lead struct {
	Base
	FirstName string `json:"lead_firstname" gorm:"column:lead_firstname;size:255" binding:"required"`
	Surname   string `json:"lead_surname" gorm:"column:lead_surname;size:255" binding:"required"`
	Email     string `json:"email" gorm:"column:email;size:255"`
	Phone     string `json:"phone" gorm:"column:phone;size:32"`
	Source    string `json:"source" gorm:"column:source;size:64;index"`
	Details   string `json:"lead_other_details" gorm:"column:lead_other_details;size:255"`

	Status             string `json:"status" gorm:"column:status;size:32;index"`
	ConvertedAccountID *uint  `json:"converted_account_id" gorm:"column:converted_account_id"`
	ConvertedContactID *uint  `json:"converted_contact_id" gorm:"column:converted_contact_id"`
}

const (
	LeadNew       = "new"
	LeadConverted = "converted"
)

func (Lead) TableName() string { return "crm_lead" }

type Account struct {
	Base
	Name           string `json:"account_name" gorm:"column:account_name;size:255" binding:"required"`
	Description    string `json:"account_description" gorm:"column:account_description;size:255"`
	Phone          string `json:"account_phone" gorm:"column:account_phone;size:20"`
	BillingAddress string `json:"billing_address" gorm:"column:billing_address;size:255"`
	Email          string `json:"email" gorm:"column:email;size:255"`
	// TaxCode is the Vietnamese mã số thuế; Representative and
	// RepresentativeTitle name who signs for the account.
	TaxCode             string `json:"tax_code" gorm:"column:tax_code;size:32"`
	Representative      string `json:"representative" gorm:"column:representative;size:255"`
	RepresentativeTitle string `json:"representative_title" gorm:"column:representative_title;size:255"`
}

func (Account) TableName() string { return "crm_account" }

type Contact struct {
	Base
	AccountID      uint   `json:"account_id" gorm:"column:account_id;index" binding:"required"`
	Name           string `json:"name" gorm:"column:name;size:255"`
	Email          string `json:"email" gorm:"column:email;size:255"`
	Phone          string `json:"phone" gorm:"column:phone;size:32"`
	Address        string `json:"contact_address" gorm:"column:contact_address;size:255"`
	ContactDetails string `json:"contact_contact_details" gorm:"column:contact_contact_details;size:255"`
}

func (Contact) TableName() string { return "crm_contact" }

type CampaignMember struct {
	Base
	CampaignID uint `json:"campaign_id" gorm:"column:campaign_id;index" binding:"required"`
	LeadID     uint `json:"lead_id" gorm:"column:lead_id;index"`
	ContactID  uint `json:"contact_id" gorm:"column:contact_id;index"`
}

func (CampaignMember) TableName() string { return "crm_campaign_member" }

type Case struct {
	Base
	ContactID   uint   `json:"contact_id" gorm:"column:contact_id;index" binding:"required"`
	Subject     string `json:"subject" gorm:"column:subject;size:255"`
	Description string `json:"description" gorm:"column:description;size:1024"`
	Priority    string `json:"priority" gorm:"column:priority;size:16;index"`

	Status   string     `json:"status" gorm:"column:status;size:16;index"`
	ClosedAt *time.Time `json:"closed_at" gorm:"column:closed_at"`
}

const (
	CaseOpen   = "open"
	CaseClosed = "closed"
)

func (Case) TableName() string { return "crm_case" }

type Contract struct {
	Base
	AccountID uint       `json:"account_id" gorm:"column:account_id;index" binding:"required"`
	Number    string     `json:"contract_number" gorm:"column:contract_number;size:64;index"`
	Title     string     `json:"contract_title" gorm:"column:contract_title;size:512"`
	EndDate   *time.Time `json:"end_date" gorm:"column:end_date;index"`

	Status   string     `json:"contract_status" gorm:"column:contract_status;size:32;index"`
	Approval string     `json:"contract_approval" gorm:"column:contract_approval;size:32"`
	SignedBy string     `json:"signed_by" gorm:"column:signed_by;size:255"`
	SignedAt *time.Time `json:"signed_at" gorm:"column:signed_at"`

	SignatureAlgorithm   string `json:"signature_algorithm,omitempty" gorm:"column:signature_algorithm;size:32"`
	Signature            string `json:"signature,omitempty" gorm:"column:signature;type:text"`
	SignerCertificate    string `json:"signer_certificate,omitempty" gorm:"column:signer_certificate;type:text"`
	SignerSubject        string `json:"signer_subject,omitempty" gorm:"column:signer_subject;size:512"`
	SignerSerial         string `json:"signer_serial,omitempty" gorm:"column:signer_serial;size:64"`
	SignerFingerprint    string `json:"signer_fingerprint,omitempty" gorm:"column:signer_fingerprint;size:64"`
	SignaturePayloadHash string `json:"signature_payload_hash,omitempty" gorm:"column:signature_payload_hash;size:64"`
	SignatureFormat      string `json:"signature_format,omitempty" gorm:"column:signature_format;size:32"`
	RevocationStatus     string `json:"revocation_status,omitempty" gorm:"column:revocation_status;size:16"`

	// Bên A's countersignature on the signed PDF (see the sealer adapter).
	CountersignerSubject     string     `json:"countersigner_subject,omitempty" gorm:"column:countersigner_subject;size:512"`
	CountersignerSerial      string     `json:"countersigner_serial,omitempty" gorm:"column:countersigner_serial;size:64"`
	CountersignerFingerprint string     `json:"countersigner_fingerprint,omitempty" gorm:"column:countersigner_fingerprint;size:64"`
	CountersignedAt          *time.Time `json:"countersigned_at,omitempty" gorm:"column:countersigned_at"`

	TimestampToken     string     `json:"timestamp_token,omitempty" gorm:"column:timestamp_token;type:text"`
	TimestampAt        *time.Time `json:"timestamp_at,omitempty" gorm:"column:timestamp_at"`
	TimestampAuthority string     `json:"timestamp_authority,omitempty" gorm:"column:timestamp_authority;size:512"`
}

func (c *Contract) CopySignatureFrom(o *Contract) {
	c.SignedBy, c.SignedAt = o.SignedBy, o.SignedAt
	c.SignatureAlgorithm, c.Signature = o.SignatureAlgorithm, o.Signature
	c.SignerCertificate, c.SignerSubject = o.SignerCertificate, o.SignerSubject
	c.SignerSerial, c.SignerFingerprint = o.SignerSerial, o.SignerFingerprint
	c.SignaturePayloadHash = o.SignaturePayloadHash
	c.SignatureFormat, c.RevocationStatus = o.SignatureFormat, o.RevocationStatus
	c.CountersignerSubject, c.CountersignerSerial = o.CountersignerSubject, o.CountersignerSerial
	c.CountersignerFingerprint, c.CountersignedAt = o.CountersignerFingerprint, o.CountersignedAt
	c.TimestampToken, c.TimestampAt, c.TimestampAuthority = o.TimestampToken, o.TimestampAt, o.TimestampAuthority
}

const (
	ContractDraft    = "draft"
	ContractPending  = "pending"
	ContractApproved = "approved"
	ContractRejected = "rejected"
	ContractExpired  = "expired"
)

func (Contract) TableName() string { return "crm_contract" }

type AccountContactRole struct {
	Base
	ContactID uint `json:"contact_id" gorm:"column:contact_id;index" binding:"required"`
	AccountID uint `json:"account_id" gorm:"column:account_id;index" binding:"required"`
}

func (AccountContactRole) TableName() string { return "crm_account_contact_role" }

type Opportunity struct {
	Base
	AccountID   uint       `json:"account_id" gorm:"column:account_id;index" binding:"required"`
	Description string     `json:"opportunity_description" gorm:"column:opportunity_description;size:255"`
	Details     string     `json:"opportunity_details" gorm:"column:opportunity_details;size:255"`
	Amount      float64    `json:"amount" gorm:"column:amount;type:decimal(15,2)"`
	CloseDate   *time.Time `json:"close_date" gorm:"column:close_date"`

	Stage string `json:"opportunity_stage" gorm:"column:opportunity_stage;size:32;index"`
}

const (
	StageProspecting   = "prospecting"
	StageQualification = "qualification"
	StageProposal      = "proposal"
	StageNegotiation   = "negotiation"
	StageWon           = "won"
	StageLost          = "lost"
)

var Stages = []string{StageProspecting, StageQualification, StageProposal, StageNegotiation, StageWon, StageLost}

func ValidStage(s string) bool {
	for _, st := range Stages {
		if st == s {
			return true
		}
	}
	return false
}

func (Opportunity) TableName() string { return "crm_opportunity" }

type OpportunityContactRole struct {
	Base
	ContactID     uint      `json:"contact_id" gorm:"column:contact_id;index" binding:"required"`
	OpportunityID uint      `json:"opportunity_id" gorm:"column:opportunity_id;index" binding:"required"`
	DateTime      time.Time `json:"date_time" gorm:"column:date_time"`
	Details       string    `json:"other_details" gorm:"column:other_details;size:255"`
}

func (OpportunityContactRole) TableName() string { return "crm_opportunity_contact_role" }

type Searchable interface {
	SearchColumns() []string
}

func (Campaign) SearchColumns() []string { return []string{"campaign_name", "campaign_sponsor"} }
func (Lead) SearchColumns() []string {
	return []string{"lead_firstname", "lead_surname", "email", "phone"}
}
func (Account) SearchColumns() []string     { return []string{"account_name", "account_phone"} }
func (Contact) SearchColumns() []string     { return []string{"name", "email", "phone"} }
func (Case) SearchColumns() []string        { return []string{"subject"} }
func (Opportunity) SearchColumns() []string { return []string{"opportunity_description"} }

func Models() []any {
	return []any{
		&ContractFile{},
		&ContractFileEvent{},
		&Campaign{}, &Lead{}, &Account{}, &Contact{}, &CampaignMember{},
		&Case{}, &Contract{}, &AccountContactRole{}, &Opportunity{},
		&OpportunityContactRole{},
	}
}

type Summary struct {
	Counts        map[string]int64 `json:"counts"`
	LeadsByStatus map[string]int64 `json:"leads_by_status"`
	OpenCases     int64            `json:"open_cases"`
	Pipeline      []StageSummary   `json:"pipeline"`
}

type StageSummary struct {
	Stage  string  `json:"stage"`
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

type Notification struct {
	Recipient string
	Title     string
	Message   string
}
