package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) HospitalQuery() generated.HospitalQueryResolver { return &hospitalQueryResolver{r} }

type hospitalQueryResolver struct{ *Resolver }

func (r *hospitalQueryResolver) GetDepartment(ctx context.Context, obj *model.HospitalQuery, departmentID int, name *string, description *string, createdAt *string, updatedAt *string) (*model.HospitalDepartment, error) {
	return r.Clients.HospitalPatientManagementService.GetDepartment(ctx, departmentID, name, description, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListDepartments(ctx context.Context, obj *model.HospitalQuery, pageSize *int, pageNumber *int) (*model.HospitalListDepartmentsResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListDepartments(ctx, pageSize, pageNumber)
}

func (r *hospitalQueryResolver) GetStaff(ctx context.Context, obj *model.HospitalQuery, staffID int, departmentID *int, firstName *string, lastName *string, role *string, specialization *string, email *string, phone *string, hireDate *string, licenseNumber *string, createdAt *string, updatedAt *string) (*model.HospitalStaff, error) {
	return r.Clients.HospitalPatientManagementService.GetStaff(ctx, staffID, departmentID, firstName, lastName, role, specialization, email, phone, hireDate, licenseNumber, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListStaff(ctx context.Context, obj *model.HospitalQuery, pageSize *int, pageNumber *int, departmentID *int) (*model.HospitalListStaffResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListStaff(ctx, pageSize, pageNumber, departmentID)
}

func (r *hospitalQueryResolver) GetPatient(ctx context.Context, obj *model.HospitalQuery, patientID int, firstName *string, lastName *string, dateOfBirth *string, gender *string, bloodType *string, address *string, phone *string, email *string, emergencyContactName *string, emergencyContactPhone *string, insuranceProvider *string, insurancePolicyNumber *string, createdAt *string, updatedAt *string) (*model.HospitalPatient, error) {
	return r.Clients.HospitalPatientManagementService.GetPatient(ctx, patientID, firstName, lastName, dateOfBirth, gender, bloodType, address, phone, email, emergencyContactName, emergencyContactPhone, insuranceProvider, insurancePolicyNumber, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListPatients(ctx context.Context, obj *model.HospitalQuery, pageSize *int, pageNumber *int, searchTerm *string) (*model.HospitalListPatientsResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListPatients(ctx, pageSize, pageNumber, searchTerm)
}

func (r *hospitalQueryResolver) GetMedicalRecord(ctx context.Context, obj *model.HospitalQuery, recordID int, patientID *int, diagnosis *string, treatmentPlan *string, notes *string, createdBy *int, createdAt *string, updatedAt *string) (*model.HospitalMedicalRecord, error) {
	return r.Clients.HospitalPatientManagementService.GetMedicalRecord(ctx, recordID, patientID, diagnosis, treatmentPlan, notes, createdBy, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListMedicalRecords(ctx context.Context, obj *model.HospitalQuery, patientID int, pageSize *int, pageNumber *int) (*model.HospitalListMedicalRecordsResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListMedicalRecords(ctx, patientID, pageSize, pageNumber)
}

func (r *hospitalQueryResolver) GetAppointment(ctx context.Context, obj *model.HospitalQuery, appointmentID int, patientID *int, staffID *int, departmentID *int, appointmentDate *string, status *string, reason *string, notes *string, createdAt *string, updatedAt *string) (*model.HospitalAppointment, error) {
	return r.Clients.HospitalPatientManagementService.GetAppointment(ctx, appointmentID, patientID, staffID, departmentID, appointmentDate, status, reason, notes, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListAppointments(ctx context.Context, obj *model.HospitalQuery, pageSize *int, pageNumber *int, patientID *int, staffID *int, dateFrom *string, dateTo *string, status *string) (*model.HospitalListAppointmentsResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListAppointments(ctx, pageSize, pageNumber, patientID, staffID, dateFrom, dateTo, status)
}

func (r *hospitalQueryResolver) GetPrescription(ctx context.Context, obj *model.HospitalQuery, prescriptionID int, patientID *int, prescribedBy *int, medicationName *string, dosage *string, frequency *string, startDate *string, endDate *string, instructions *string, createdAt *string, updatedAt *string) (*model.HospitalPrescription, error) {
	return r.Clients.HospitalPatientManagementService.GetPrescription(ctx, prescriptionID, patientID, prescribedBy, medicationName, dosage, frequency, startDate, endDate, instructions, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListPrescriptions(ctx context.Context, obj *model.HospitalQuery, patientID int, pageSize *int, pageNumber *int) (*model.HospitalListPrescriptionsResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListPrescriptions(ctx, patientID, pageSize, pageNumber)
}

func (r *hospitalQueryResolver) GetLabResult(ctx context.Context, obj *model.HospitalQuery, resultID int, patientID *int, orderedBy *int, testName *string, testDate *string, results *string, normalRange *string, notes *string, createdAt *string, updatedAt *string) (*model.HospitalLabResult, error) {
	return r.Clients.HospitalPatientManagementService.GetLabResult(ctx, resultID, patientID, orderedBy, testName, testDate, results, normalRange, notes, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListLabResults(ctx context.Context, obj *model.HospitalQuery, patientID int, pageSize *int, pageNumber *int) (*model.HospitalListLabResultsResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListLabResults(ctx, patientID, pageSize, pageNumber)
}

func (r *hospitalQueryResolver) GetBilling(ctx context.Context, obj *model.HospitalQuery, billID int, patientID *int, appointmentID *int, amount *float64, insuranceCovered *float64, patientResponsibility *float64, status *string, dueDate *string, paidDate *string, createdAt *string, updatedAt *string) (*model.HospitalBilling, error) {
	return r.Clients.HospitalPatientManagementService.GetBilling(ctx, billID, patientID, appointmentID, amount, insuranceCovered, patientResponsibility, status, dueDate, paidDate, createdAt, updatedAt)
}

func (r *hospitalQueryResolver) ListBillings(ctx context.Context, obj *model.HospitalQuery, patientID int, pageSize *int, pageNumber *int, status *string) (*model.HospitalListBillingsResponse, error) {
	return r.Clients.HospitalPatientManagementService.ListBillings(ctx, patientID, pageSize, pageNumber, status)
}

func (r *hospitalQueryResolver) GetDepartmentStats(ctx context.Context, obj *model.HospitalQuery) (*model.HospitalDepartmentStatsResponse, error) {
	return r.Clients.HospitalPatientManagementService.GetDepartmentStats(ctx)
}

func (r *hospitalQueryResolver) GetAppointmentStats(ctx context.Context, obj *model.HospitalQuery, dateFrom *string, dateTo *string, departmentID *int) (*model.HospitalAppointmentStatsResponse, error) {
	return r.Clients.HospitalPatientManagementService.GetAppointmentStats(ctx, dateFrom, dateTo, departmentID)
}

func (r *hospitalQueryResolver) GetBillingStats(ctx context.Context, obj *model.HospitalQuery, dateFrom *string, dateTo *string) (*model.HospitalBillingStatsResponse, error) {
	return r.Clients.HospitalPatientManagementService.GetBillingStats(ctx, dateFrom, dateTo)
}

func (r *hospitalQueryResolver) PatientDocuments(ctx context.Context, obj *model.HospitalQuery, id int, limit *int, offset *int) (*model.HospitalDocumentList, error) {
	return r.Clients.HospitalPatientManagementService.PatientDocuments(ctx, id, limit, offset)
}

func (r *hospitalQueryResolver) PatientIdentity(ctx context.Context, obj *model.HospitalQuery, id int) (*model.HospitalIdentityStatus, error) {
	return r.Clients.HospitalPatientManagementService.PatientIdentity(ctx, id)
}
