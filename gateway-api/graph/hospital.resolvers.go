package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy"
)

func (r *mutationResolver) HospitalCreateDepartment(ctx context.Context, input model.HospitalCreateDepartmentRequestInput) (*model.HospitalDepartment, error) {
	return grpcInvoke[model.HospitalDepartment](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/CreateDepartment", "hospital.CreateDepartmentRequest", "hospital.Department", input)
}

func (r *mutationResolver) HospitalUpdateDepartment(ctx context.Context, input model.HospitalDepartmentInput) (*model.HospitalDepartment, error) {
	return grpcInvoke[model.HospitalDepartment](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/UpdateDepartment", "hospital.Department", "hospital.Department", input)
}

func (r *mutationResolver) HospitalDeleteDepartment(ctx context.Context, input model.HospitalDepartmentInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/DeleteDepartment", "hospital.Department", "google.protobuf.Empty", input)
}

func (r *mutationResolver) HospitalCreateStaff(ctx context.Context, input model.HospitalCreateStaffRequestInput) (*model.HospitalStaff, error) {
	return grpcInvoke[model.HospitalStaff](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/CreateStaff", "hospital.CreateStaffRequest", "hospital.Staff", input)
}

func (r *mutationResolver) HospitalUpdateStaff(ctx context.Context, input model.HospitalStaffInput) (*model.HospitalStaff, error) {
	return grpcInvoke[model.HospitalStaff](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/UpdateStaff", "hospital.Staff", "hospital.Staff", input)
}

func (r *mutationResolver) HospitalDeleteStaff(ctx context.Context, input model.HospitalStaffInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/DeleteStaff", "hospital.Staff", "google.protobuf.Empty", input)
}

func (r *mutationResolver) HospitalCreatePatient(ctx context.Context, input model.HospitalCreatePatientRequestInput) (*model.HospitalPatient, error) {
	return grpcInvoke[model.HospitalPatient](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/CreatePatient", "hospital.CreatePatientRequest", "hospital.Patient", input)
}

func (r *mutationResolver) HospitalUpdatePatient(ctx context.Context, input model.HospitalPatientInput) (*model.HospitalPatient, error) {
	return grpcInvoke[model.HospitalPatient](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/UpdatePatient", "hospital.Patient", "hospital.Patient", input)
}

func (r *mutationResolver) HospitalDeletePatient(ctx context.Context, input model.HospitalPatientInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/DeletePatient", "hospital.Patient", "google.protobuf.Empty", input)
}

func (r *mutationResolver) HospitalCreateMedicalRecord(ctx context.Context, input model.HospitalCreateMedicalRecordRequestInput) (*model.HospitalMedicalRecord, error) {
	return grpcInvoke[model.HospitalMedicalRecord](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/CreateMedicalRecord", "hospital.CreateMedicalRecordRequest", "hospital.MedicalRecord", input)
}

func (r *mutationResolver) HospitalUpdateMedicalRecord(ctx context.Context, input model.HospitalMedicalRecordInput) (*model.HospitalMedicalRecord, error) {
	return grpcInvoke[model.HospitalMedicalRecord](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/UpdateMedicalRecord", "hospital.MedicalRecord", "hospital.MedicalRecord", input)
}

func (r *mutationResolver) HospitalDeleteMedicalRecord(ctx context.Context, input model.HospitalMedicalRecordInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/DeleteMedicalRecord", "hospital.MedicalRecord", "google.protobuf.Empty", input)
}

func (r *mutationResolver) HospitalCreateAppointment(ctx context.Context, input model.HospitalCreateAppointmentRequestInput) (*model.HospitalAppointment, error) {
	return grpcInvoke[model.HospitalAppointment](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/CreateAppointment", "hospital.CreateAppointmentRequest", "hospital.Appointment", input)
}

func (r *mutationResolver) HospitalUpdateAppointmentStatus(ctx context.Context, input model.HospitalUpdateAppointmentStatusRequestInput) (*model.HospitalAppointment, error) {
	return grpcInvoke[model.HospitalAppointment](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/UpdateAppointmentStatus", "hospital.UpdateAppointmentStatusRequest", "hospital.Appointment", input)
}

func (r *mutationResolver) HospitalUpdateAppointment(ctx context.Context, input model.HospitalAppointmentInput) (*model.HospitalAppointment, error) {
	return grpcInvoke[model.HospitalAppointment](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/UpdateAppointment", "hospital.Appointment", "hospital.Appointment", input)
}

func (r *mutationResolver) HospitalDeleteAppointment(ctx context.Context, input model.HospitalAppointmentInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/DeleteAppointment", "hospital.Appointment", "google.protobuf.Empty", input)
}

func (r *mutationResolver) HospitalCreatePrescription(ctx context.Context, input model.HospitalCreatePrescriptionRequestInput) (*model.HospitalPrescription, error) {
	return grpcInvoke[model.HospitalPrescription](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/CreatePrescription", "hospital.CreatePrescriptionRequest", "hospital.Prescription", input)
}

func (r *mutationResolver) HospitalUpdatePrescription(ctx context.Context, input model.HospitalPrescriptionInput) (*model.HospitalPrescription, error) {
	return grpcInvoke[model.HospitalPrescription](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/UpdatePrescription", "hospital.Prescription", "hospital.Prescription", input)
}

func (r *mutationResolver) HospitalDeletePrescription(ctx context.Context, input model.HospitalPrescriptionInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/DeletePrescription", "hospital.Prescription", "google.protobuf.Empty", input)
}

func (r *mutationResolver) HospitalCreateLabResult(ctx context.Context, input model.HospitalCreateLabResultRequestInput) (*model.HospitalLabResult, error) {
	return grpcInvoke[model.HospitalLabResult](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/CreateLabResult", "hospital.CreateLabResultRequest", "hospital.LabResult", input)
}

func (r *mutationResolver) HospitalUpdateLabResult(ctx context.Context, input model.HospitalLabResultInput) (*model.HospitalLabResult, error) {
	return grpcInvoke[model.HospitalLabResult](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/UpdateLabResult", "hospital.LabResult", "hospital.LabResult", input)
}

func (r *mutationResolver) HospitalDeleteLabResult(ctx context.Context, input model.HospitalLabResultInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/DeleteLabResult", "hospital.LabResult", "google.protobuf.Empty", input)
}

func (r *mutationResolver) HospitalCreateBilling(ctx context.Context, input model.HospitalCreateBillingRequestInput) (*model.HospitalBilling, error) {
	return grpcInvoke[model.HospitalBilling](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/CreateBilling", "hospital.CreateBillingRequest", "hospital.Billing", input)
}

func (r *mutationResolver) HospitalUpdateBillingStatus(ctx context.Context, input model.HospitalUpdateBillingStatusRequestInput) (*model.HospitalBilling, error) {
	return grpcInvoke[model.HospitalBilling](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/UpdateBillingStatus", "hospital.UpdateBillingStatusRequest", "hospital.Billing", input)
}

func (r *mutationResolver) HospitalUpdateBilling(ctx context.Context, input model.HospitalBillingInput) (*model.HospitalBilling, error) {
	return grpcInvoke[model.HospitalBilling](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/UpdateBilling", "hospital.Billing", "hospital.Billing", input)
}

func (r *mutationResolver) HospitalDeleteBilling(ctx context.Context, input model.HospitalBillingInput) (bool, error) {
	return grpcInvokeEmpty(ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/DeleteBilling", "hospital.Billing", "google.protobuf.Empty", input)
}

func (r *queryResolver) HospitalGetDepartment(ctx context.Context, input model.HospitalDepartmentInput) (*model.HospitalDepartment, error) {
	return grpcInvoke[model.HospitalDepartment](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/GetDepartment", "hospital.Department", "hospital.Department", input)
}

func (r *queryResolver) HospitalListDepartments(ctx context.Context, input model.HospitalListDepartmentsRequestInput) (*model.HospitalListDepartmentsResponse, error) {
	return grpcInvoke[model.HospitalListDepartmentsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/ListDepartments", "hospital.ListDepartmentsRequest", "hospital.ListDepartmentsResponse", input)
}

func (r *queryResolver) HospitalGetStaff(ctx context.Context, input model.HospitalStaffInput) (*model.HospitalStaff, error) {
	return grpcInvoke[model.HospitalStaff](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/GetStaff", "hospital.Staff", "hospital.Staff", input)
}

func (r *queryResolver) HospitalListStaff(ctx context.Context, input model.HospitalListStaffRequestInput) (*model.HospitalListStaffResponse, error) {
	return grpcInvoke[model.HospitalListStaffResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/ListStaff", "hospital.ListStaffRequest", "hospital.ListStaffResponse", input)
}

func (r *queryResolver) HospitalGetPatient(ctx context.Context, input model.HospitalPatientInput) (*model.HospitalPatient, error) {
	return grpcInvoke[model.HospitalPatient](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/GetPatient", "hospital.Patient", "hospital.Patient", input)
}

func (r *queryResolver) HospitalListPatients(ctx context.Context, input model.HospitalListPatientsRequestInput) (*model.HospitalListPatientsResponse, error) {
	return grpcInvoke[model.HospitalListPatientsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/ListPatients", "hospital.ListPatientsRequest", "hospital.ListPatientsResponse", input)
}

func (r *queryResolver) HospitalGetMedicalRecord(ctx context.Context, input model.HospitalMedicalRecordInput) (*model.HospitalMedicalRecord, error) {
	return grpcInvoke[model.HospitalMedicalRecord](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/GetMedicalRecord", "hospital.MedicalRecord", "hospital.MedicalRecord", input)
}

func (r *queryResolver) HospitalListMedicalRecords(ctx context.Context, input model.HospitalListMedicalRecordsRequestInput) (*model.HospitalListMedicalRecordsResponse, error) {
	return grpcInvoke[model.HospitalListMedicalRecordsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.HospotalService/ListMedicalRecords", "hospital.ListMedicalRecordsRequest", "hospital.ListMedicalRecordsResponse", input)
}

func (r *queryResolver) HospitalGetAppointment(ctx context.Context, input model.HospitalAppointmentInput) (*model.HospitalAppointment, error) {
	return grpcInvoke[model.HospitalAppointment](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/GetAppointment", "hospital.Appointment", "hospital.Appointment", input)
}

func (r *queryResolver) HospitalListAppointments(ctx context.Context, input model.HospitalListAppointmentsRequestInput) (*model.HospitalListAppointmentsResponse, error) {
	return grpcInvoke[model.HospitalListAppointmentsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/ListAppointments", "hospital.ListAppointmentsRequest", "hospital.ListAppointmentsResponse", input)
}

func (r *queryResolver) HospitalGetPrescription(ctx context.Context, input model.HospitalPrescriptionInput) (*model.HospitalPrescription, error) {
	return grpcInvoke[model.HospitalPrescription](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/GetPrescription", "hospital.Prescription", "hospital.Prescription", input)
}

func (r *queryResolver) HospitalListPrescriptions(ctx context.Context, input model.HospitalListPrescriptionsRequestInput) (*model.HospitalListPrescriptionsResponse, error) {
	return grpcInvoke[model.HospitalListPrescriptionsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/ListPrescriptions", "hospital.ListPrescriptionsRequest", "hospital.ListPrescriptionsResponse", input)
}

func (r *queryResolver) HospitalGetLabResult(ctx context.Context, input model.HospitalLabResultInput) (*model.HospitalLabResult, error) {
	return grpcInvoke[model.HospitalLabResult](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/GetLabResult", "hospital.LabResult", "hospital.LabResult", input)
}

func (r *queryResolver) HospitalListLabResults(ctx context.Context, input model.HospitalListLabResultsRequestInput) (*model.HospitalListLabResultsResponse, error) {
	return grpcInvoke[model.HospitalListLabResultsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/ListLabResults", "hospital.ListLabResultsRequest", "hospital.ListLabResultsResponse", input)
}

func (r *queryResolver) HospitalGetBilling(ctx context.Context, input model.HospitalBillingInput) (*model.HospitalBilling, error) {
	return grpcInvoke[model.HospitalBilling](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/GetBilling", "hospital.Billing", "hospital.Billing", input)
}

func (r *queryResolver) HospitalListBillings(ctx context.Context, input model.HospitalListBillingsRequestInput) (*model.HospitalListBillingsResponse, error) {
	return grpcInvoke[model.HospitalListBillingsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/ListBillings", "hospital.ListBillingsRequest", "hospital.ListBillingsResponse", input)
}

func (r *queryResolver) HospitalGetDepartmentStats(ctx context.Context) (*model.HospitalDepartmentStatsResponse, error) {
	return grpcInvoke[model.HospitalDepartmentStatsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/GetDepartmentStats", "google.protobuf.Empty", "hospital.DepartmentStatsResponse", nil)
}

func (r *queryResolver) HospitalGetAppointmentStats(ctx context.Context, input model.HospitalAppointmentStatsRequestInput) (*model.HospitalAppointmentStatsResponse, error) {
	return grpcInvoke[model.HospitalAppointmentStatsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/GetAppointmentStats", "hospital.AppointmentStatsRequest", "hospital.AppointmentStatsResponse", input)
}

func (r *queryResolver) HospitalGetBillingStats(ctx context.Context, input model.HospitalBillingStatsRequestInput) (*model.HospitalBillingStatsResponse, error) {
	return grpcInvoke[model.HospitalBillingStatsResponse](ctx, r.Proxy, proxy.HospitalService, "/hospital.AppointmentService/GetBillingStats", "hospital.BillingStatsRequest", "hospital.BillingStatsResponse", input)
}
