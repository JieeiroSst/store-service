package grpcapi

import (
	"time"

	pb "github.com/JIeeiroSst/lib-gateway/hospital-patientm-anagement-service/gateway/hospital-patientm-anagement-service"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const dateLayout = "2006-01-02"

func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func fromTS(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

func fromTSPtr(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	v := t.AsTime()
	return &v
}

func date(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dateLayout)
}

func datePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return date(*t)
}

func parseDate(field, s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, model.Invalid("%s must be formatted as YYYY-MM-DD", field)
	}
	return t, nil
}

func parseDatePtr(field, s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := parseDate(field, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

var genders = map[pb.Gender]model.Gender{
	pb.Gender_GENDER_MALE:   model.GenderMale,
	pb.Gender_GENDER_FEMALE: model.GenderFemale,
	pb.Gender_GENDER_OTHER:  model.GenderOther,
}

var bloodTypes = map[pb.BloodType]model.BloodType{
	pb.BloodType_BLOOD_TYPE_A_POSITIVE:  model.BloodAPositive,
	pb.BloodType_BLOOD_TYPE_A_NEGATIVE:  model.BloodANegative,
	pb.BloodType_BLOOD_TYPE_B_POSITIVE:  model.BloodBPositive,
	pb.BloodType_BLOOD_TYPE_B_NEGATIVE:  model.BloodBNegative,
	pb.BloodType_BLOOD_TYPE_AB_POSITIVE: model.BloodABPositive,
	pb.BloodType_BLOOD_TYPE_AB_NEGATIVE: model.BloodABNegative,
	pb.BloodType_BLOOD_TYPE_O_POSITIVE:  model.BloodOPositive,
	pb.BloodType_BLOOD_TYPE_O_NEGATIVE:  model.BloodONegative,
}

var appointmentStatuses = map[pb.AppointmentStatus]model.AppointmentStatus{
	pb.AppointmentStatus_APPOINTMENT_STATUS_SCHEDULED: model.AppointmentScheduled,
	pb.AppointmentStatus_APPOINTMENT_STATUS_COMPLETED: model.AppointmentCompleted,
	pb.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED: model.AppointmentCancelled,
	pb.AppointmentStatus_APPOINTMENT_STATUS_NO_SHOW:   model.AppointmentNoShow,
}

func genderToPB(g model.Gender) pb.Gender {
	for k, v := range genders {
		if v == g {
			return k
		}
	}
	return pb.Gender_GENDER_UNSPECIFIED
}

func bloodTypeToPB(b *model.BloodType) pb.BloodType {
	if b != nil {
		for k, v := range bloodTypes {
			if v == *b {
				return k
			}
		}
	}
	return pb.BloodType_BLOOD_TYPE_UNSPECIFIED
}

func bloodTypeFromPB(b pb.BloodType) *model.BloodType {
	if v, ok := bloodTypes[b]; ok {
		return &v
	}
	return nil
}

func statusToPB(s model.AppointmentStatus) pb.AppointmentStatus {
	for k, v := range appointmentStatuses {
		if v == s {
			return k
		}
	}
	return pb.AppointmentStatus_APPOINTMENT_STATUS_UNSPECIFIED
}

// statusFromPB returns "" for UNSPECIFIED so callers can treat it as "any".
func statusFromPB(s pb.AppointmentStatus) model.AppointmentStatus {
	return appointmentStatuses[s]
}

func optID(id int32) *int32 {
	if id <= 0 {
		return nil
	}
	return &id
}

func idOf(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

func page(size, number int32) model.Page {
	return model.Page{Number: number, Size: size}
}

// ---- model -> pb ----

func departmentPB(d *model.Department) *pb.Department {
	return &pb.Department{
		DepartmentId: d.ID, Name: d.Name, Description: d.Description,
		CreatedAt: ts(d.CreatedAt), UpdatedAt: ts(d.UpdatedAt),
	}
}

func staffPB(s *model.Staff) *pb.Staff {
	lic := ""
	if s.LicenseNumber != nil {
		lic = *s.LicenseNumber
	}
	return &pb.Staff{
		StaffId: s.ID, DepartmentId: s.DepartmentID, FirstName: s.FirstName, LastName: s.LastName,
		Role: s.Role, Specialization: s.Specialization, Email: s.Email, Phone: s.Phone,
		HireDate: date(s.HireDate), LicenseNumber: lic,
		CreatedAt: ts(s.CreatedAt), UpdatedAt: ts(s.UpdatedAt),
	}
}

func patientPB(p *model.Patient) *pb.Patient {
	return &pb.Patient{
		PatientId: p.ID, FirstName: p.FirstName, LastName: p.LastName,
		DateOfBirth: date(p.DateOfBirth), Gender: genderToPB(p.Gender), BloodType: bloodTypeToPB(p.BloodType),
		Address: p.Address, Phone: p.Phone, Email: p.Email,
		EmergencyContactName: p.EmergencyContactName, EmergencyContactPhone: p.EmergencyContactPhone,
		InsuranceProvider: p.InsuranceProvider, InsurancePolicyNumber: p.InsurancePolicyNumber,
		CreatedAt: ts(p.CreatedAt), UpdatedAt: ts(p.UpdatedAt),
	}
}

func medicalRecordPB(r *model.MedicalRecord) *pb.MedicalRecord {
	return &pb.MedicalRecord{
		RecordId: r.ID, PatientId: r.PatientID, Diagnosis: r.Diagnosis, TreatmentPlan: r.TreatmentPlan,
		Notes: r.Notes, CreatedBy: idOf(r.CreatedBy), CreatedAt: ts(r.CreatedAt), UpdatedAt: ts(r.UpdatedAt),
	}
}

func appointmentPB(a *model.Appointment) *pb.Appointment {
	return &pb.Appointment{
		AppointmentId: a.ID, PatientId: a.PatientID, StaffId: a.StaffID, DepartmentId: a.DepartmentID,
		AppointmentDate: ts(a.AppointmentDate), Status: statusToPB(a.Status), Reason: a.Reason, Notes: a.Notes,
		CreatedAt: ts(a.CreatedAt), UpdatedAt: ts(a.UpdatedAt),
	}
}

func prescriptionPB(p *model.Prescription) *pb.Prescription {
	return &pb.Prescription{
		PrescriptionId: p.ID, PatientId: p.PatientID, PrescribedBy: p.PrescribedBy,
		MedicationName: p.MedicationName, Dosage: p.Dosage, Frequency: p.Frequency,
		StartDate: date(p.StartDate), EndDate: datePtr(p.EndDate), Instructions: p.Instructions,
		CreatedAt: ts(p.CreatedAt), UpdatedAt: ts(p.UpdatedAt),
	}
}

func labResultPB(r *model.LabResult) *pb.LabResult {
	return &pb.LabResult{
		ResultId: r.ID, PatientId: r.PatientID, OrderedBy: r.OrderedBy, TestName: r.TestName,
		TestDate: ts(r.TestDate), Results: r.Results, NormalRange: r.NormalRange, Notes: r.Notes,
		CreatedAt: ts(r.CreatedAt), UpdatedAt: ts(r.UpdatedAt),
	}
}

func billingPB(b *model.Billing) *pb.Billing {
	return &pb.Billing{
		BillId: b.ID, PatientId: b.PatientID, AppointmentId: idOf(b.AppointmentID),
		Amount: float32(b.Amount), InsuranceCovered: float32(b.InsuranceCovered),
		PatientResponsibility: float32(b.PatientResponsibility), Status: string(b.Status),
		DueDate: datePtr(b.DueDate), PaidDate: datePtr(b.PaidDate),
		CreatedAt: ts(b.CreatedAt), UpdatedAt: ts(b.UpdatedAt),
	}
}

func mapAll[M, P any](in []M, f func(*M) P) []P {
	out := make([]P, len(in))
	for i := range in {
		out[i] = f(&in[i])
	}
	return out
}

// ---- pb -> model (create and update requests share these) ----

type staffFields struct {
	DepartmentID                                     int32
	FirstName, LastName, Role, Specialization, Email string
	Phone, HireDate, LicenseNumber                   string
}

func staffModel(id int32, f staffFields) (*model.Staff, error) {
	hire, err := parseDate("hire_date", f.HireDate)
	if err != nil {
		return nil, err
	}
	s := &model.Staff{
		ID: id, DepartmentID: f.DepartmentID, FirstName: f.FirstName, LastName: f.LastName, Role: f.Role,
		Specialization: f.Specialization, Email: f.Email, Phone: f.Phone, HireDate: hire,
	}
	if f.LicenseNumber != "" {
		s.LicenseNumber = &f.LicenseNumber
	}
	return s, nil
}

type patientFields struct {
	FirstName, LastName, DateOfBirth string
	Gender                           pb.Gender
	BloodType                        pb.BloodType
	Address, Phone, Email            string
	EmergencyName, EmergencyPhone    string
	InsuranceProvider, InsurancePol  string
}

func patientModel(id int32, f patientFields) (*model.Patient, error) {
	dob, err := parseDate("date_of_birth", f.DateOfBirth)
	if err != nil {
		return nil, err
	}
	return &model.Patient{
		ID: id, FirstName: f.FirstName, LastName: f.LastName, DateOfBirth: dob,
		Gender: genders[f.Gender], BloodType: bloodTypeFromPB(f.BloodType),
		Address: f.Address, Phone: f.Phone, Email: f.Email,
		EmergencyContactName: f.EmergencyName, EmergencyContactPhone: f.EmergencyPhone,
		InsuranceProvider: f.InsuranceProvider, InsurancePolicyNumber: f.InsurancePol,
	}, nil
}

type prescriptionFields struct {
	PatientID, PrescribedBy          int32
	Medication, Dosage, Frequency    string
	StartDate, EndDate, Instructions string
}

func prescriptionModel(id int32, f prescriptionFields) (*model.Prescription, error) {
	start, err := parseDate("start_date", f.StartDate)
	if err != nil {
		return nil, err
	}
	end, err := parseDatePtr("end_date", f.EndDate)
	if err != nil {
		return nil, err
	}
	return &model.Prescription{
		ID: id, PatientID: f.PatientID, PrescribedBy: f.PrescribedBy, MedicationName: f.Medication,
		Dosage: f.Dosage, Frequency: f.Frequency, StartDate: start, EndDate: end, Instructions: f.Instructions,
	}, nil
}

type billingFields struct {
	PatientID, AppointmentID   int32
	Amount, Insurance, Patient float32
	Status, DueDate            string
}

func billingModel(id int32, f billingFields) (*model.Billing, error) {
	due, err := parseDatePtr("due_date", f.DueDate)
	if err != nil {
		return nil, err
	}
	return &model.Billing{
		ID: id, PatientID: f.PatientID, AppointmentID: optID(f.AppointmentID),
		Amount: float64(f.Amount), InsuranceCovered: float64(f.Insurance), PatientResponsibility: float64(f.Patient),
		Status: model.BillingStatus(f.Status), DueDate: due,
	}, nil
}
