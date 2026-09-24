package repository

import (
	"context"
	"strings"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DepartmentRepo struct {
	crud[model.Department]
	db *gorm.DB
}

func NewDepartmentRepo(db *gorm.DB) port.DepartmentRepository {
	return &DepartmentRepo{crud[model.Department]{db, "department_id"}, db}
}

func (r *DepartmentRepo) List(ctx context.Context, page model.Page) ([]model.Department, int64, error) {
	return paginate[model.Department](r.db.WithContext(ctx).Model(&model.Department{}), page, "name")
}

func (r *DepartmentRepo) Stats(ctx context.Context) ([]model.DepartmentStat, error) {
	var out []model.DepartmentStat
	// Two correlated counts avoid the row multiplication of joining both
	// staff and appointments.
	err := r.db.WithContext(ctx).Raw(`
		SELECT d.department_id AS department_id,
		       d.name AS department_name,
		       (SELECT COUNT(*) FROM staff s WHERE s.department_id = d.department_id) AS staff_count,
		       (SELECT COUNT(*) FROM appointments a WHERE a.department_id = d.department_id) AS appointment_count
		FROM departments d
		ORDER BY d.name`).Scan(&out).Error
	return out, err
}

type StaffRepo struct {
	crud[model.Staff]
	db *gorm.DB
}

func NewStaffRepo(db *gorm.DB) port.StaffRepository {
	return &StaffRepo{crud[model.Staff]{db, "staff_id"}, db}
}

func (r *StaffRepo) List(ctx context.Context, departmentID int32, page model.Page) ([]model.Staff, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Staff{})
	if departmentID > 0 {
		q = q.Where("department_id = ?", departmentID)
	}
	return paginate[model.Staff](q, page, "last_name, first_name")
}

type PatientRepo struct {
	crud[model.Patient]
	db *gorm.DB
}

func NewPatientRepo(db *gorm.DB) port.PatientRepository {
	return &PatientRepo{crud[model.Patient]{db, "patient_id"}, db}
}

func (r *PatientRepo) List(ctx context.Context, search string, page model.Page) ([]model.Patient, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Patient{})
	if search != "" {
		like := "%" + escapeLike(strings.ToLower(search)) + "%"
		q = q.Where(`LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?
			OR LOWER(first_name || ' ' || last_name) LIKE ? OR phone LIKE ? OR LOWER(email) LIKE ?`,
			like, like, like, like, like)
	}
	return paginate[model.Patient](q, page, "last_name, first_name, patient_id")
}

// escapeLike neutralises LIKE wildcards in user input.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

type MedicalRecordRepo struct {
	crud[model.MedicalRecord]
	db *gorm.DB
}

func NewMedicalRecordRepo(db *gorm.DB) port.MedicalRecordRepository {
	return &MedicalRecordRepo{crud[model.MedicalRecord]{db, "record_id"}, db}
}

func (r *MedicalRecordRepo) ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.MedicalRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.MedicalRecord{}).Where("patient_id = ?", patientID)
	return paginate[model.MedicalRecord](q, page, "created_at DESC, record_id DESC")
}

type PrescriptionRepo struct {
	crud[model.Prescription]
	db *gorm.DB
}

func NewPrescriptionRepo(db *gorm.DB) port.PrescriptionRepository {
	return &PrescriptionRepo{crud[model.Prescription]{db, "prescription_id"}, db}
}

func (r *PrescriptionRepo) ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.Prescription, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Prescription{}).Where("patient_id = ?", patientID)
	return paginate[model.Prescription](q, page, "start_date DESC, prescription_id DESC")
}

type LabResultRepo struct {
	crud[model.LabResult]
	db *gorm.DB
}

func NewLabResultRepo(db *gorm.DB) port.LabResultRepository {
	return &LabResultRepo{crud[model.LabResult]{db, "result_id"}, db}
}

func (r *LabResultRepo) ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.LabResult, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.LabResult{}).Where("patient_id = ?", patientID)
	return paginate[model.LabResult](q, page, "test_date DESC, result_id DESC")
}

type AppointmentRepo struct {
	crud[model.Appointment]
	db *gorm.DB
}

func NewAppointmentRepo(db *gorm.DB) port.AppointmentRepository {
	return &AppointmentRepo{crud[model.Appointment]{db, "appointment_id"}, db}
}

func (r *AppointmentRepo) List(ctx context.Context, f model.AppointmentFilter, page model.Page) ([]model.Appointment, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Appointment{})
	if f.PatientID > 0 {
		q = q.Where("patient_id = ?", f.PatientID)
	}
	if f.StaffID > 0 {
		q = q.Where("staff_id = ?", f.StaffID)
	}
	if f.From != nil {
		q = q.Where("appointment_date >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("appointment_date <= ?", *f.To)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	return paginate[model.Appointment](q, page, "appointment_date, appointment_id")
}

func (r *AppointmentRepo) HasScheduled(ctx context.Context, staffID int32, at time.Time, excludeID int32) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Appointment{}).
		Where("staff_id = ? AND appointment_date = ? AND status = ? AND appointment_id <> ?",
			staffID, at, model.AppointmentScheduled, excludeID).
		Count(&n).Error
	return n > 0, err
}

func (r *AppointmentRepo) Stats(ctx context.Context, from, to *time.Time, departmentID int32) (*model.AppointmentStats, error) {
	scope := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Appointment{})
		if from != nil {
			q = q.Where("appointment_date >= ?", *from)
		}
		if to != nil {
			q = q.Where("appointment_date <= ?", *to)
		}
		if departmentID > 0 {
			q = q.Where("department_id = ?", departmentID)
		}
		return q
	}

	var byStatus []struct {
		Status model.AppointmentStatus
		N      int32
	}
	if err := scope().Select("status, COUNT(*) AS n").Group("status").Scan(&byStatus).Error; err != nil {
		return nil, err
	}
	out := &model.AppointmentStats{}
	for _, s := range byStatus {
		out.Total += s.N
		switch s.Status {
		case model.AppointmentCompleted:
			out.Completed = s.N
		case model.AppointmentCancelled:
			out.Cancelled = s.N
		case model.AppointmentNoShow:
			out.NoShow = s.N
		}
	}

	err := scope().
		Select("to_char(appointment_date, 'YYYY-MM-DD') AS date, COUNT(*) AS count").
		Group("date").Order("date").Scan(&out.Daily).Error
	return out, err
}

type BillingRepo struct {
	crud[model.Billing]
	db *gorm.DB
}

func NewBillingRepo(db *gorm.DB) port.BillingRepository {
	return &BillingRepo{crud[model.Billing]{db, "bill_id"}, db}
}

func (r *BillingRepo) List(ctx context.Context, patientID int32, status model.BillingStatus, page model.Page) ([]model.Billing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Billing{}).Where("patient_id = ?", patientID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return paginate[model.Billing](q, page, "created_at DESC, bill_id DESC")
}

func (r *BillingRepo) Stats(ctx context.Context, from, to *time.Time) (*model.BillingStats, error) {
	scope := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Billing{}).Where("status <> ?", model.BillingCancelled)
		if from != nil {
			q = q.Where("created_at >= ?", *from)
		}
		if to != nil {
			q = q.Where("created_at <= ?", *to)
		}
		return q
	}

	// Collected = paid bills; outstanding = everything not paid.
	var totals struct {
		TotalBilled, TotalCollected, TotalOutstanding float64
		TotalInsuranceCovered, TotalPatientResp       float64
	}
	err := scope().Select(`
		COALESCE(SUM(amount), 0) AS total_billed,
		COALESCE(SUM(amount) FILTER (WHERE status = 'Paid'), 0) AS total_collected,
		COALESCE(SUM(amount) FILTER (WHERE status <> 'Paid'), 0) AS total_outstanding,
		COALESCE(SUM(insurance_covered), 0) AS total_insurance_covered,
		COALESCE(SUM(patient_responsibility), 0) AS total_patient_resp`).Scan(&totals).Error
	if err != nil {
		return nil, err
	}
	out := &model.BillingStats{
		TotalBilled: totals.TotalBilled, TotalCollected: totals.TotalCollected,
		TotalOutstanding: totals.TotalOutstanding, TotalInsuranceCovered: totals.TotalInsuranceCovered,
		TotalPatientResp: totals.TotalPatientResp,
	}

	err = scope().Select(`
		to_char(created_at, 'YYYY-MM') AS month,
		COALESCE(SUM(amount), 0) AS billed,
		COALESCE(SUM(amount) FILTER (WHERE status = 'Paid'), 0) AS collected`).
		Group("month").Order("month").Scan(&out.Monthly).Error
	return out, err
}

type BillingAccountRepo struct{ db *gorm.DB }

func NewBillingAccountRepo(db *gorm.DB) port.BillingAccountRepository {
	return &BillingAccountRepo{db}
}

func (r *BillingAccountRepo) Get(ctx context.Context, patientID int32) (*model.BillingAccount, error) {
	var a model.BillingAccount
	if err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).First(&a).Error; err != nil {
		return nil, translate(err, false)
	}
	return &a, nil
}

func (r *BillingAccountRepo) Save(ctx context.Context, a *model.BillingAccount) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(a).Error
	return translate(err, false)
}

// accountLockClass namespaces our advisory locks; the patient id is the key.
const accountLockClass = 727002

// WithLock holds a transaction-scoped Postgres advisory lock while fn runs.
// fn does its work on its own connections: the transaction exists only to own
// the lock, which Postgres releases when it ends, even if this process dies.
func (r *BillingAccountRepo) WithLock(ctx context.Context, patientID int32, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", accountLockClass, patientID).Error; err != nil {
			return err
		}
		return fn(ctx)
	})
}
