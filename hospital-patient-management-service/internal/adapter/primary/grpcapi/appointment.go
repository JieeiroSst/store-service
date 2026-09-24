package grpcapi

import (
	"context"

	pb "github.com/JIeeiroSst/lib-gateway/hospital-patientm-anagement-service/gateway/hospital-patientm-anagement-service"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AppointmentHandler struct {
	pb.UnimplementedAppointmentServiceServer

	appointments  port.AppointmentUsecase
	prescriptions port.PrescriptionUsecase
	labResults    port.LabResultUsecase
	billings      port.BillingUsecase
	departments   port.DepartmentUsecase
}

func NewAppointmentHandler(
	appointments port.AppointmentUsecase,
	prescriptions port.PrescriptionUsecase,
	labResults port.LabResultUsecase,
	billings port.BillingUsecase,
	departments port.DepartmentUsecase,
) *AppointmentHandler {
	return &AppointmentHandler{
		appointments: appointments, prescriptions: prescriptions, labResults: labResults,
		billings: billings, departments: departments,
	}
}

// ---- appointments ----

func (h *AppointmentHandler) CreateAppointment(ctx context.Context, in *pb.CreateAppointmentRequest) (*pb.Appointment, error) {
	a, err := h.appointments.Create(ctx, &model.Appointment{
		PatientID: in.PatientId, StaffID: in.StaffId, DepartmentID: in.DepartmentId,
		AppointmentDate: fromTS(in.AppointmentDate), Reason: in.Reason, Notes: in.Notes,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return appointmentPB(a), nil
}

func (h *AppointmentHandler) GetAppointment(ctx context.Context, in *pb.Appointment) (*pb.Appointment, error) {
	a, err := h.appointments.Get(ctx, in.AppointmentId)
	if err != nil {
		return nil, toStatus(err)
	}
	return appointmentPB(a), nil
}

func (h *AppointmentHandler) UpdateAppointmentStatus(ctx context.Context, in *pb.UpdateAppointmentStatusRequest) (*pb.Appointment, error) {
	a, err := h.appointments.UpdateStatus(ctx, in.AppointmentId, statusFromPB(in.Status))
	if err != nil {
		return nil, toStatus(err)
	}
	return appointmentPB(a), nil
}

func (h *AppointmentHandler) ListAppointments(ctx context.Context, in *pb.ListAppointmentsRequest) (*pb.ListAppointmentsResponse, error) {
	items, total, err := h.appointments.List(ctx, model.AppointmentFilter{
		PatientID: in.PatientId, StaffID: in.StaffId,
		From: fromTSPtr(in.DateFrom), To: fromTSPtr(in.DateTo), Status: statusFromPB(in.Status),
	}, page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListAppointmentsResponse{Appointments: mapAll(items, appointmentPB), TotalCount: int32(total)}, nil
}

func (h *AppointmentHandler) UpdateAppointment(ctx context.Context, in *pb.Appointment) (*pb.Appointment, error) {
	a, err := h.appointments.Update(ctx, &model.Appointment{
		ID: in.AppointmentId, PatientID: in.PatientId, StaffID: in.StaffId, DepartmentID: in.DepartmentId,
		AppointmentDate: fromTS(in.AppointmentDate), Reason: in.Reason, Notes: in.Notes,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return appointmentPB(a), nil
}

func (h *AppointmentHandler) DeleteAppointment(ctx context.Context, in *pb.Appointment) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, toStatus(h.appointments.Delete(ctx, in.AppointmentId))
}

// ---- prescriptions ----

func (h *AppointmentHandler) CreatePrescription(ctx context.Context, in *pb.CreatePrescriptionRequest) (*pb.Prescription, error) {
	m, err := prescriptionModel(0, prescriptionFields{
		PatientID: in.PatientId, PrescribedBy: in.PrescribedBy, Medication: in.MedicationName,
		Dosage: in.Dosage, Frequency: in.Frequency, StartDate: in.StartDate, EndDate: in.EndDate,
		Instructions: in.Instructions,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	p, err := h.prescriptions.Create(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return prescriptionPB(p), nil
}

func (h *AppointmentHandler) GetPrescription(ctx context.Context, in *pb.Prescription) (*pb.Prescription, error) {
	p, err := h.prescriptions.Get(ctx, in.PrescriptionId)
	if err != nil {
		return nil, toStatus(err)
	}
	return prescriptionPB(p), nil
}

func (h *AppointmentHandler) ListPrescriptions(ctx context.Context, in *pb.ListPrescriptionsRequest) (*pb.ListPrescriptionsResponse, error) {
	items, total, err := h.prescriptions.ListByPatient(ctx, in.PatientId, page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListPrescriptionsResponse{Prescriptions: mapAll(items, prescriptionPB), TotalCount: int32(total)}, nil
}

func (h *AppointmentHandler) UpdatePrescription(ctx context.Context, in *pb.Prescription) (*pb.Prescription, error) {
	m, err := prescriptionModel(in.PrescriptionId, prescriptionFields{
		PatientID: in.PatientId, PrescribedBy: in.PrescribedBy, Medication: in.MedicationName,
		Dosage: in.Dosage, Frequency: in.Frequency, StartDate: in.StartDate, EndDate: in.EndDate,
		Instructions: in.Instructions,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	p, err := h.prescriptions.Update(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return prescriptionPB(p), nil
}

func (h *AppointmentHandler) DeletePrescription(ctx context.Context, in *pb.Prescription) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, toStatus(h.prescriptions.Delete(ctx, in.PrescriptionId))
}

// ---- lab results ----

func (h *AppointmentHandler) CreateLabResult(ctx context.Context, in *pb.CreateLabResultRequest) (*pb.LabResult, error) {
	r, err := h.labResults.Create(ctx, &model.LabResult{
		PatientID: in.PatientId, OrderedBy: in.OrderedBy, TestName: in.TestName, TestDate: fromTS(in.TestDate),
		Results: in.Results, NormalRange: in.NormalRange, Notes: in.Notes,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return labResultPB(r), nil
}

func (h *AppointmentHandler) GetLabResult(ctx context.Context, in *pb.LabResult) (*pb.LabResult, error) {
	r, err := h.labResults.Get(ctx, in.ResultId)
	if err != nil {
		return nil, toStatus(err)
	}
	return labResultPB(r), nil
}

func (h *AppointmentHandler) ListLabResults(ctx context.Context, in *pb.ListLabResultsRequest) (*pb.ListLabResultsResponse, error) {
	items, total, err := h.labResults.ListByPatient(ctx, in.PatientId, page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListLabResultsResponse{LabResults: mapAll(items, labResultPB), TotalCount: int32(total)}, nil
}

func (h *AppointmentHandler) UpdateLabResult(ctx context.Context, in *pb.LabResult) (*pb.LabResult, error) {
	r, err := h.labResults.Update(ctx, &model.LabResult{
		ID: in.ResultId, PatientID: in.PatientId, OrderedBy: in.OrderedBy, TestName: in.TestName,
		TestDate: fromTS(in.TestDate), Results: in.Results, NormalRange: in.NormalRange, Notes: in.Notes,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return labResultPB(r), nil
}

func (h *AppointmentHandler) DeleteLabResult(ctx context.Context, in *pb.LabResult) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, toStatus(h.labResults.Delete(ctx, in.ResultId))
}

// ---- billing ----

func (h *AppointmentHandler) CreateBilling(ctx context.Context, in *pb.CreateBillingRequest) (*pb.Billing, error) {
	m, err := billingModel(0, billingFields{
		PatientID: in.PatientId, AppointmentID: in.AppointmentId, Amount: in.Amount,
		Insurance: in.InsuranceCovered, Patient: in.PatientResponsibility, Status: in.Status, DueDate: in.DueDate,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	b, err := h.billings.Create(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return billingPB(b), nil
}

func (h *AppointmentHandler) GetBilling(ctx context.Context, in *pb.Billing) (*pb.Billing, error) {
	b, err := h.billings.Get(ctx, in.BillId)
	if err != nil {
		return nil, toStatus(err)
	}
	return billingPB(b), nil
}

func (h *AppointmentHandler) UpdateBillingStatus(ctx context.Context, in *pb.UpdateBillingStatusRequest) (*pb.Billing, error) {
	paid, err := parseDatePtr("paid_date", in.PaidDate)
	if err != nil {
		return nil, toStatus(err)
	}
	b, err := h.billings.UpdateStatus(ctx, in.BillId, model.BillingStatus(in.Status), paid)
	if err != nil {
		return nil, toStatus(err)
	}
	return billingPB(b), nil
}

func (h *AppointmentHandler) ListBillings(ctx context.Context, in *pb.ListBillingsRequest) (*pb.ListBillingsResponse, error) {
	items, total, err := h.billings.List(ctx, in.PatientId, model.BillingStatus(in.Status), page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListBillingsResponse{Billings: mapAll(items, billingPB), TotalCount: int32(total)}, nil
}

func (h *AppointmentHandler) UpdateBilling(ctx context.Context, in *pb.Billing) (*pb.Billing, error) {
	m, err := billingModel(in.BillId, billingFields{
		PatientID: in.PatientId, AppointmentID: in.AppointmentId, Amount: in.Amount,
		Insurance: in.InsuranceCovered, Patient: in.PatientResponsibility, Status: in.Status, DueDate: in.DueDate,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	b, err := h.billings.Update(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return billingPB(b), nil
}

func (h *AppointmentHandler) DeleteBilling(ctx context.Context, in *pb.Billing) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, toStatus(h.billings.Delete(ctx, in.BillId))
}

// ---- analytics ----

func (h *AppointmentHandler) GetDepartmentStats(ctx context.Context, _ *emptypb.Empty) (*pb.DepartmentStatsResponse, error) {
	stats, err := h.departments.Stats(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.DepartmentStatsResponse{}
	for _, s := range stats {
		out.DepartmentStats = append(out.DepartmentStats, &pb.DepartmentStatsResponse_DepartmentStat{
			DepartmentId: s.DepartmentID, DepartmentName: s.DepartmentName,
			StaffCount: s.StaffCount, AppointmentCount: s.AppointmentCount,
		})
	}
	return out, nil
}

func (h *AppointmentHandler) GetAppointmentStats(ctx context.Context, in *pb.AppointmentStatsRequest) (*pb.AppointmentStatsResponse, error) {
	s, err := h.appointments.Stats(ctx, fromTSPtr(in.DateFrom), fromTSPtr(in.DateTo), in.DepartmentId)
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.AppointmentStatsResponse{
		TotalAppointments: s.Total, CompletedAppointments: s.Completed,
		CancelledAppointments: s.Cancelled, NoShowAppointments: s.NoShow,
	}
	for _, d := range s.Daily {
		out.DailyStats = append(out.DailyStats, &pb.AppointmentStatsResponse_DailyStats{Date: d.Date, AppointmentCount: d.Count})
	}
	return out, nil
}

func (h *AppointmentHandler) GetBillingStats(ctx context.Context, in *pb.BillingStatsRequest) (*pb.BillingStatsResponse, error) {
	s, err := h.billings.Stats(ctx, fromTSPtr(in.DateFrom), fromTSPtr(in.DateTo))
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.BillingStatsResponse{
		TotalBilled: float32(s.TotalBilled), TotalCollected: float32(s.TotalCollected),
		TotalOutstanding: float32(s.TotalOutstanding), TotalInsuranceCovered: float32(s.TotalInsuranceCovered),
		TotalPatientResponsibility: float32(s.TotalPatientResp),
	}
	for _, m := range s.Monthly {
		out.MonthlyStats = append(out.MonthlyStats, &pb.BillingStatsResponse_MonthlyStats{
			Month: m.Month, AmountBilled: float32(m.Billed), AmountCollected: float32(m.Collected),
		})
	}
	return out, nil
}
