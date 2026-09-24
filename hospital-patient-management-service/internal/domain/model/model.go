package model

import "time"

type Gender string

const (
	GenderMale   Gender = "Male"
	GenderFemale Gender = "Female"
	GenderOther  Gender = "Other"
)

func (g Gender) Valid() bool {
	return g == GenderMale || g == GenderFemale || g == GenderOther
}

type BloodType string

const (
	BloodAPositive  BloodType = "A+"
	BloodANegative  BloodType = "A-"
	BloodBPositive  BloodType = "B+"
	BloodBNegative  BloodType = "B-"
	BloodABPositive BloodType = "AB+"
	BloodABNegative BloodType = "AB-"
	BloodOPositive  BloodType = "O+"
	BloodONegative  BloodType = "O-"
)

func (b BloodType) Valid() bool {
	switch b {
	case BloodAPositive, BloodANegative, BloodBPositive, BloodBNegative,
		BloodABPositive, BloodABNegative, BloodOPositive, BloodONegative:
		return true
	}
	return false
}

type AppointmentStatus string

const (
	AppointmentScheduled AppointmentStatus = "Scheduled"
	AppointmentCompleted AppointmentStatus = "Completed"
	AppointmentCancelled AppointmentStatus = "Cancelled"
	AppointmentNoShow    AppointmentStatus = "No-Show"
)

func (s AppointmentStatus) Valid() bool {
	switch s {
	case AppointmentScheduled, AppointmentCompleted, AppointmentCancelled, AppointmentNoShow:
		return true
	}
	return false
}

// CanTransitionTo: only a scheduled appointment can move on; the other
// statuses are final.
func (s AppointmentStatus) CanTransitionTo(next AppointmentStatus) bool {
	return s == AppointmentScheduled && next.Valid() && next != AppointmentScheduled
}

type BillingStatus string

const (
	BillingPending   BillingStatus = "Pending"
	BillingPaid      BillingStatus = "Paid"
	BillingOverdue   BillingStatus = "Overdue"
	BillingCancelled BillingStatus = "Cancelled"
)

func (s BillingStatus) Valid() bool {
	switch s {
	case BillingPending, BillingPaid, BillingOverdue, BillingCancelled:
		return true
	}
	return false
}

// Field tags follow the schema documented in the service Readme.

type Department struct {
	ID          int32  `gorm:"column:department_id;primaryKey"`
	Name        string `gorm:"size:100;not null"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Department) TableName() string { return "departments" }

type Staff struct {
	ID             int32     `gorm:"column:staff_id;primaryKey"`
	DepartmentID   int32     `gorm:"index"`
	FirstName      string    `gorm:"size:50;not null"`
	LastName       string    `gorm:"size:50;not null"`
	Role           string    `gorm:"size:50;not null"`
	Specialization string    `gorm:"size:100"`
	Email          string    `gorm:"size:100;not null;uniqueIndex"`
	Phone          string    `gorm:"size:20"`
	HireDate       time.Time `gorm:"type:date;not null"`
	LicenseNumber  *string   `gorm:"size:50;uniqueIndex"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Staff) TableName() string { return "staff" }

type Patient struct {
	ID                    int32      `gorm:"column:patient_id;primaryKey"`
	FirstName             string     `gorm:"size:50;not null"`
	LastName              string     `gorm:"size:50;not null"`
	DateOfBirth           time.Time  `gorm:"type:date;not null"`
	Gender                Gender     `gorm:"type:gender_type;not null"`
	BloodType             *BloodType `gorm:"type:blood_type"`
	Address               string
	Phone                 string `gorm:"size:20"`
	Email                 string `gorm:"size:100"`
	EmergencyContactName  string `gorm:"size:100"`
	EmergencyContactPhone string `gorm:"size:20"`
	InsuranceProvider     string `gorm:"size:100"`
	InsurancePolicyNumber string `gorm:"size:50"`
	// UserID links the patient to their user-service account. It is set only
	// through the identity endpoints (it is not part of the gRPC contract) and
	// is what wallet payments and eKYC are keyed on.
	UserID    *string `gorm:"size:64;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Patient) TableName() string { return "patients" }

type MedicalRecord struct {
	ID            int32  `gorm:"column:record_id;primaryKey"`
	PatientID     int32  `gorm:"index;not null"`
	Diagnosis     string `gorm:"not null"`
	TreatmentPlan string
	Notes         string
	CreatedBy     *int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (MedicalRecord) TableName() string { return "medical_records" }

type Appointment struct {
	ID              int32             `gorm:"column:appointment_id;primaryKey"`
	PatientID       int32             `gorm:"not null"`
	StaffID         int32             `gorm:"not null;index:idx_appointments_staff_date,priority:1"`
	DepartmentID    int32             `gorm:"not null"`
	AppointmentDate time.Time         `gorm:"not null;index;index:idx_appointments_staff_date,priority:2"`
	Status          AppointmentStatus `gorm:"type:appointment_status;not null;default:Scheduled"`
	Reason          string
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Appointment) TableName() string { return "appointments" }

type Prescription struct {
	ID             int32      `gorm:"column:prescription_id;primaryKey"`
	PatientID      int32      `gorm:"index;not null"`
	PrescribedBy   int32      `gorm:"not null"`
	MedicationName string     `gorm:"size:100;not null"`
	Dosage         string     `gorm:"size:50;not null"`
	Frequency      string     `gorm:"size:50;not null"`
	StartDate      time.Time  `gorm:"type:date;not null"`
	EndDate        *time.Time `gorm:"type:date"`
	Instructions   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Prescription) TableName() string { return "prescriptions" }

type LabResult struct {
	ID          int32     `gorm:"column:result_id;primaryKey"`
	PatientID   int32     `gorm:"index;not null"`
	OrderedBy   int32     `gorm:"not null"`
	TestName    string    `gorm:"size:100;not null"`
	TestDate    time.Time `gorm:"not null"`
	Results     string    `gorm:"not null"`
	NormalRange string
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (LabResult) TableName() string { return "lab_results" }

type Billing struct {
	ID                    int32 `gorm:"column:bill_id;primaryKey"`
	PatientID             int32 `gorm:"index;not null"`
	AppointmentID         *int32
	Amount                float64       `gorm:"type:numeric(10,2);not null"`
	InsuranceCovered      float64       `gorm:"type:numeric(10,2);not null;default:0"`
	PatientResponsibility float64       `gorm:"type:numeric(10,2);not null;default:0"`
	Status                BillingStatus `gorm:"size:50;not null;default:Pending;index"`
	DueDate               *time.Time    `gorm:"type:date"`
	PaidDate              *time.Time    `gorm:"type:date"`
	// InvoiceID is the invoice billing-service issued for what the patient
	// owes; nil while there is nothing to invoice or the sync is pending.
	InvoiceID *int32
	// WalletTransferID is the wallet transfer that paid this bill, kept so a
	// later cancellation can refund it.
	WalletTransferID *string `gorm:"size:64"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (Billing) TableName() string { return "billing" }

// BillingAccount links a patient to their customer and subscription in
// billing-service, which invoices hang off.
type BillingAccount struct {
	PatientID      int32 `gorm:"primaryKey;autoIncrement:false"`
	CustomerID     int32 `gorm:"not null"`
	SubscriptionID int32 `gorm:"not null"`
	CreatedAt      time.Time
}

func (BillingAccount) TableName() string { return "billing_accounts" }

// Models lists every persisted entity in dependency order.
func Models() []any {
	return []any{
		&Department{}, &Staff{}, &Patient{}, &MedicalRecord{},
		&Appointment{}, &Prescription{}, &LabResult{}, &Billing{}, &BillingAccount{},
	}
}

type Page struct {
	Number int32
	Size   int32
}

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Normalized clamps a caller supplied page to sane bounds (1-based).
func (p Page) Normalized() Page {
	if p.Number < 1 {
		p.Number = 1
	}
	if p.Size < 1 {
		p.Size = DefaultPageSize
	}
	if p.Size > MaxPageSize {
		p.Size = MaxPageSize
	}
	return p
}

func (p Page) Offset() int { p = p.Normalized(); return int((p.Number - 1) * p.Size) }
func (p Page) Limit() int  { return int(p.Normalized().Size) }

type AppointmentFilter struct {
	PatientID int32
	StaffID   int32
	From, To  *time.Time
	Status    AppointmentStatus
}

type DepartmentStat struct {
	DepartmentID     int32
	DepartmentName   string
	StaffCount       int32
	AppointmentCount int32
}

type DailyStat struct {
	Date  string
	Count int32
}

type AppointmentStats struct {
	Total, Completed, Cancelled, NoShow int32
	Daily                               []DailyStat
}

type MonthlyBilling struct {
	Month     string
	Billed    float64
	Collected float64
}

type BillingStats struct {
	TotalBilled, TotalCollected, TotalOutstanding float64
	TotalInsuranceCovered, TotalPatientResp       float64
	Monthly                                       []MonthlyBilling
}

// IdentityState is where a patient stands in eKYC.
type IdentityState string

const (
	IdentityNotStarted IdentityState = "not_started"
	IdentityPending    IdentityState = "pending"
	IdentityVerified   IdentityState = "verified"
	IdentityFailed     IdentityState = "failed"
)

// IdentityStatus deliberately carries no card fields or biometrics: the
// hospital needs the outcome, not the identity document.
type IdentityStatus struct {
	State      IdentityState
	MatchScore float64
	VerifiedAt *time.Time
	HasCard    bool
	HasFace    bool
}

// Document is a file attached to a patient (a scan, report or image). Its
// bytes are stored by upload-service; the hospital keeps no copy.
type Document struct {
	ID          string
	PatientID   int32
	FileName    string
	ContentType string
	Size        int64
	CreatedAt   time.Time
}
