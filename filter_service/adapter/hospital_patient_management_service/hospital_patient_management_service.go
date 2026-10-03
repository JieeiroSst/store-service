package hospital_patient_management_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "hospital-patient-management-service"

const DefaultBaseURL = "http://hospital-patient-management-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) GetDepartment(ctx context.Context, departmentID int, name *string, description *string, createdAt *string, updatedAt *string) (*model.HospitalDepartment, error) {
	path := "/v1/departments/" + url.PathEscape(strconv.Itoa(departmentID))
	q := url.Values{}
	if name != nil {
		q.Set("name", *name)
	}
	if description != nil {
		q.Set("description", *description)
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalDepartment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListDepartments(ctx context.Context, pageSize *int, pageNumber *int) (*model.HospitalListDepartmentsResponse, error) {
	path := "/v1/departments"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	h := http.Header{}
	var out *model.HospitalListDepartmentsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetStaff(ctx context.Context, staffID int, departmentID *int, firstName *string, lastName *string, role *string, specialization *string, email *string, phone *string, hireDate *string, licenseNumber *string, createdAt *string, updatedAt *string) (*model.HospitalStaff, error) {
	path := "/v1/staff/" + url.PathEscape(strconv.Itoa(staffID))
	q := url.Values{}
	if departmentID != nil {
		q.Set("department_id", strconv.Itoa(*departmentID))
	}
	if firstName != nil {
		q.Set("first_name", *firstName)
	}
	if lastName != nil {
		q.Set("last_name", *lastName)
	}
	if role != nil {
		q.Set("role", *role)
	}
	if specialization != nil {
		q.Set("specialization", *specialization)
	}
	if email != nil {
		q.Set("email", *email)
	}
	if phone != nil {
		q.Set("phone", *phone)
	}
	if hireDate != nil {
		q.Set("hire_date", *hireDate)
	}
	if licenseNumber != nil {
		q.Set("license_number", *licenseNumber)
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalStaff
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListStaff(ctx context.Context, pageSize *int, pageNumber *int, departmentID *int) (*model.HospitalListStaffResponse, error) {
	path := "/v1/staff"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	if departmentID != nil {
		q.Set("department_id", strconv.Itoa(*departmentID))
	}
	h := http.Header{}
	var out *model.HospitalListStaffResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetPatient(ctx context.Context, patientID int, firstName *string, lastName *string, dateOfBirth *string, gender *string, bloodType *string, address *string, phone *string, email *string, emergencyContactName *string, emergencyContactPhone *string, insuranceProvider *string, insurancePolicyNumber *string, createdAt *string, updatedAt *string) (*model.HospitalPatient, error) {
	path := "/v1/patients/" + url.PathEscape(strconv.Itoa(patientID))
	q := url.Values{}
	if firstName != nil {
		q.Set("first_name", *firstName)
	}
	if lastName != nil {
		q.Set("last_name", *lastName)
	}
	if dateOfBirth != nil {
		q.Set("date_of_birth", *dateOfBirth)
	}
	if gender != nil {
		q.Set("gender", *gender)
	}
	if bloodType != nil {
		q.Set("blood_type", *bloodType)
	}
	if address != nil {
		q.Set("address", *address)
	}
	if phone != nil {
		q.Set("phone", *phone)
	}
	if email != nil {
		q.Set("email", *email)
	}
	if emergencyContactName != nil {
		q.Set("emergency_contact_name", *emergencyContactName)
	}
	if emergencyContactPhone != nil {
		q.Set("emergency_contact_phone", *emergencyContactPhone)
	}
	if insuranceProvider != nil {
		q.Set("insurance_provider", *insuranceProvider)
	}
	if insurancePolicyNumber != nil {
		q.Set("insurance_policy_number", *insurancePolicyNumber)
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalPatient
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListPatients(ctx context.Context, pageSize *int, pageNumber *int, searchTerm *string) (*model.HospitalListPatientsResponse, error) {
	path := "/v1/patients"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	if searchTerm != nil {
		q.Set("search_term", *searchTerm)
	}
	h := http.Header{}
	var out *model.HospitalListPatientsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetMedicalRecord(ctx context.Context, recordID int, patientID *int, diagnosis *string, treatmentPlan *string, notes *string, createdBy *int, createdAt *string, updatedAt *string) (*model.HospitalMedicalRecord, error) {
	path := "/v1/medical-records/" + url.PathEscape(strconv.Itoa(recordID))
	q := url.Values{}
	if patientID != nil {
		q.Set("patient_id", strconv.Itoa(*patientID))
	}
	if diagnosis != nil {
		q.Set("diagnosis", *diagnosis)
	}
	if treatmentPlan != nil {
		q.Set("treatment_plan", *treatmentPlan)
	}
	if notes != nil {
		q.Set("notes", *notes)
	}
	if createdBy != nil {
		q.Set("created_by", strconv.Itoa(*createdBy))
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalMedicalRecord
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListMedicalRecords(ctx context.Context, patientID int, pageSize *int, pageNumber *int) (*model.HospitalListMedicalRecordsResponse, error) {
	path := "/v1/patients/" + url.PathEscape(strconv.Itoa(patientID)) + "/medical-records"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	h := http.Header{}
	var out *model.HospitalListMedicalRecordsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetAppointment(ctx context.Context, appointmentID int, patientID *int, staffID *int, departmentID *int, appointmentDate *string, status *string, reason *string, notes *string, createdAt *string, updatedAt *string) (*model.HospitalAppointment, error) {
	path := "/v1/appointments/" + url.PathEscape(strconv.Itoa(appointmentID))
	q := url.Values{}
	if patientID != nil {
		q.Set("patient_id", strconv.Itoa(*patientID))
	}
	if staffID != nil {
		q.Set("staff_id", strconv.Itoa(*staffID))
	}
	if departmentID != nil {
		q.Set("department_id", strconv.Itoa(*departmentID))
	}
	if appointmentDate != nil {
		q.Set("appointment_date", *appointmentDate)
	}
	if status != nil {
		q.Set("status", *status)
	}
	if reason != nil {
		q.Set("reason", *reason)
	}
	if notes != nil {
		q.Set("notes", *notes)
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalAppointment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListAppointments(ctx context.Context, pageSize *int, pageNumber *int, patientID *int, staffID *int, dateFrom *string, dateTo *string, status *string) (*model.HospitalListAppointmentsResponse, error) {
	path := "/v1/appointments"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	if patientID != nil {
		q.Set("patient_id", strconv.Itoa(*patientID))
	}
	if staffID != nil {
		q.Set("staff_id", strconv.Itoa(*staffID))
	}
	if dateFrom != nil {
		q.Set("date_from", *dateFrom)
	}
	if dateTo != nil {
		q.Set("date_to", *dateTo)
	}
	if status != nil {
		q.Set("status", *status)
	}
	h := http.Header{}
	var out *model.HospitalListAppointmentsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetPrescription(ctx context.Context, prescriptionID int, patientID *int, prescribedBy *int, medicationName *string, dosage *string, frequency *string, startDate *string, endDate *string, instructions *string, createdAt *string, updatedAt *string) (*model.HospitalPrescription, error) {
	path := "/v1/prescriptions/" + url.PathEscape(strconv.Itoa(prescriptionID))
	q := url.Values{}
	if patientID != nil {
		q.Set("patient_id", strconv.Itoa(*patientID))
	}
	if prescribedBy != nil {
		q.Set("prescribed_by", strconv.Itoa(*prescribedBy))
	}
	if medicationName != nil {
		q.Set("medication_name", *medicationName)
	}
	if dosage != nil {
		q.Set("dosage", *dosage)
	}
	if frequency != nil {
		q.Set("frequency", *frequency)
	}
	if startDate != nil {
		q.Set("start_date", *startDate)
	}
	if endDate != nil {
		q.Set("end_date", *endDate)
	}
	if instructions != nil {
		q.Set("instructions", *instructions)
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalPrescription
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListPrescriptions(ctx context.Context, patientID int, pageSize *int, pageNumber *int) (*model.HospitalListPrescriptionsResponse, error) {
	path := "/v1/patients/" + url.PathEscape(strconv.Itoa(patientID)) + "/prescriptions"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	h := http.Header{}
	var out *model.HospitalListPrescriptionsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetLabResult(ctx context.Context, resultID int, patientID *int, orderedBy *int, testName *string, testDate *string, results *string, normalRange *string, notes *string, createdAt *string, updatedAt *string) (*model.HospitalLabResult, error) {
	path := "/v1/lab-results/" + url.PathEscape(strconv.Itoa(resultID))
	q := url.Values{}
	if patientID != nil {
		q.Set("patient_id", strconv.Itoa(*patientID))
	}
	if orderedBy != nil {
		q.Set("ordered_by", strconv.Itoa(*orderedBy))
	}
	if testName != nil {
		q.Set("test_name", *testName)
	}
	if testDate != nil {
		q.Set("test_date", *testDate)
	}
	if results != nil {
		q.Set("results", *results)
	}
	if normalRange != nil {
		q.Set("normal_range", *normalRange)
	}
	if notes != nil {
		q.Set("notes", *notes)
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalLabResult
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListLabResults(ctx context.Context, patientID int, pageSize *int, pageNumber *int) (*model.HospitalListLabResultsResponse, error) {
	path := "/v1/patients/" + url.PathEscape(strconv.Itoa(patientID)) + "/lab-results"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	h := http.Header{}
	var out *model.HospitalListLabResultsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetBilling(ctx context.Context, billID int, patientID *int, appointmentID *int, amount *float64, insuranceCovered *float64, patientResponsibility *float64, status *string, dueDate *string, paidDate *string, createdAt *string, updatedAt *string) (*model.HospitalBilling, error) {
	path := "/v1/billings/" + url.PathEscape(strconv.Itoa(billID))
	q := url.Values{}
	if patientID != nil {
		q.Set("patient_id", strconv.Itoa(*patientID))
	}
	if appointmentID != nil {
		q.Set("appointment_id", strconv.Itoa(*appointmentID))
	}
	if amount != nil {
		q.Set("amount", strconv.FormatFloat(*amount, 'f', -1, 64))
	}
	if insuranceCovered != nil {
		q.Set("insurance_covered", strconv.FormatFloat(*insuranceCovered, 'f', -1, 64))
	}
	if patientResponsibility != nil {
		q.Set("patient_responsibility", strconv.FormatFloat(*patientResponsibility, 'f', -1, 64))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if dueDate != nil {
		q.Set("due_date", *dueDate)
	}
	if paidDate != nil {
		q.Set("paid_date", *paidDate)
	}
	if createdAt != nil {
		q.Set("created_at", *createdAt)
	}
	if updatedAt != nil {
		q.Set("updated_at", *updatedAt)
	}
	h := http.Header{}
	var out *model.HospitalBilling
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListBillings(ctx context.Context, patientID int, pageSize *int, pageNumber *int, status *string) (*model.HospitalListBillingsResponse, error) {
	path := "/v1/patients/" + url.PathEscape(strconv.Itoa(patientID)) + "/billings"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageNumber != nil {
		q.Set("page_number", strconv.Itoa(*pageNumber))
	}
	if status != nil {
		q.Set("status", *status)
	}
	h := http.Header{}
	var out *model.HospitalListBillingsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetDepartmentStats(ctx context.Context) (*model.HospitalDepartmentStatsResponse, error) {
	path := "/v1/analytics/departments"
	q := url.Values{}
	h := http.Header{}
	var out *model.HospitalDepartmentStatsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetAppointmentStats(ctx context.Context, dateFrom *string, dateTo *string, departmentID *int) (*model.HospitalAppointmentStatsResponse, error) {
	path := "/v1/analytics/appointments"
	q := url.Values{}
	if dateFrom != nil {
		q.Set("date_from", *dateFrom)
	}
	if dateTo != nil {
		q.Set("date_to", *dateTo)
	}
	if departmentID != nil {
		q.Set("department_id", strconv.Itoa(*departmentID))
	}
	h := http.Header{}
	var out *model.HospitalAppointmentStatsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetBillingStats(ctx context.Context, dateFrom *string, dateTo *string) (*model.HospitalBillingStatsResponse, error) {
	path := "/v1/analytics/billings"
	q := url.Values{}
	if dateFrom != nil {
		q.Set("date_from", *dateFrom)
	}
	if dateTo != nil {
		q.Set("date_to", *dateTo)
	}
	h := http.Header{}
	var out *model.HospitalBillingStatsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PatientDocuments(ctx context.Context, id int, limit *int, offset *int) (*model.HospitalDocumentList, error) {
	path := "/v1/patients/" + url.PathEscape(strconv.Itoa(id)) + "/documents"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.HospitalDocumentList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PatientIdentity(ctx context.Context, id int) (*model.HospitalIdentityStatus, error) {
	path := "/v1/patients/" + url.PathEscape(strconv.Itoa(id)) + "/identity"
	q := url.Values{}
	h := http.Header{}
	var out *model.HospitalIdentityStatus
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
