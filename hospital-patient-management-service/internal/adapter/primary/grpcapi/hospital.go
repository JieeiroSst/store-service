package grpcapi

import (
	"context"

	pb "github.com/JIeeiroSst/lib-gateway/hospital-patientm-anagement-service/gateway/hospital-patientm-anagement-service"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"google.golang.org/protobuf/types/known/emptypb"
)

// HospitalHandler adapts the HospotalService gRPC contract (departments,
// staff, patients, medical records) onto the application usecases.
type HospitalHandler struct {
	pb.UnimplementedHospotalServiceServer

	departments    port.DepartmentUsecase
	staff          port.StaffUsecase
	patients       port.PatientUsecase
	medicalRecords port.MedicalRecordUsecase
}

func NewHospitalHandler(
	departments port.DepartmentUsecase,
	staff port.StaffUsecase,
	patients port.PatientUsecase,
	medicalRecords port.MedicalRecordUsecase,
) *HospitalHandler {
	return &HospitalHandler{departments: departments, staff: staff, patients: patients, medicalRecords: medicalRecords}
}

// ---- departments ----

func (h *HospitalHandler) CreateDepartment(ctx context.Context, in *pb.CreateDepartmentRequest) (*pb.Department, error) {
	d, err := h.departments.Create(ctx, &model.Department{Name: in.Name, Description: in.Description})
	if err != nil {
		return nil, toStatus(err)
	}
	return departmentPB(d), nil
}

func (h *HospitalHandler) GetDepartment(ctx context.Context, in *pb.Department) (*pb.Department, error) {
	d, err := h.departments.Get(ctx, in.DepartmentId)
	if err != nil {
		return nil, toStatus(err)
	}
	return departmentPB(d), nil
}

func (h *HospitalHandler) ListDepartments(ctx context.Context, in *pb.ListDepartmentsRequest) (*pb.ListDepartmentsResponse, error) {
	items, total, err := h.departments.List(ctx, page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListDepartmentsResponse{Departments: mapAll(items, departmentPB), TotalCount: int32(total)}, nil
}

func (h *HospitalHandler) UpdateDepartment(ctx context.Context, in *pb.Department) (*pb.Department, error) {
	d, err := h.departments.Update(ctx, &model.Department{ID: in.DepartmentId, Name: in.Name, Description: in.Description})
	if err != nil {
		return nil, toStatus(err)
	}
	return departmentPB(d), nil
}

func (h *HospitalHandler) DeleteDepartment(ctx context.Context, in *pb.Department) (*emptypb.Empty, error) {
	if err := h.departments.Delete(ctx, in.DepartmentId); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

// ---- staff ----

func (h *HospitalHandler) CreateStaff(ctx context.Context, in *pb.CreateStaffRequest) (*pb.Staff, error) {
	m, err := staffModel(0, staffFields{
		DepartmentID: in.DepartmentId, FirstName: in.FirstName, LastName: in.LastName, Role: in.Role,
		Specialization: in.Specialization, Email: in.Email, Phone: in.Phone, HireDate: in.HireDate,
		LicenseNumber: in.LicenseNumber,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	s, err := h.staff.Create(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return staffPB(s), nil
}

func (h *HospitalHandler) GetStaff(ctx context.Context, in *pb.Staff) (*pb.Staff, error) {
	s, err := h.staff.Get(ctx, in.StaffId)
	if err != nil {
		return nil, toStatus(err)
	}
	return staffPB(s), nil
}

func (h *HospitalHandler) ListStaff(ctx context.Context, in *pb.ListStaffRequest) (*pb.ListStaffResponse, error) {
	items, total, err := h.staff.List(ctx, in.DepartmentId, page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListStaffResponse{Staff: mapAll(items, staffPB), TotalCount: int32(total)}, nil
}

func (h *HospitalHandler) UpdateStaff(ctx context.Context, in *pb.Staff) (*pb.Staff, error) {
	m, err := staffModel(in.StaffId, staffFields{
		DepartmentID: in.DepartmentId, FirstName: in.FirstName, LastName: in.LastName, Role: in.Role,
		Specialization: in.Specialization, Email: in.Email, Phone: in.Phone, HireDate: in.HireDate,
		LicenseNumber: in.LicenseNumber,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	s, err := h.staff.Update(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return staffPB(s), nil
}

func (h *HospitalHandler) DeleteStaff(ctx context.Context, in *pb.Staff) (*emptypb.Empty, error) {
	if err := h.staff.Delete(ctx, in.StaffId); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

// ---- patients ----

func (h *HospitalHandler) CreatePatient(ctx context.Context, in *pb.CreatePatientRequest) (*pb.Patient, error) {
	m, err := patientModel(0, patientFields{
		FirstName: in.FirstName, LastName: in.LastName, DateOfBirth: in.DateOfBirth, Gender: in.Gender,
		BloodType: in.BloodType, Address: in.Address, Phone: in.Phone, Email: in.Email,
		EmergencyName: in.EmergencyContactName, EmergencyPhone: in.EmergencyContactPhone,
		InsuranceProvider: in.InsuranceProvider, InsurancePol: in.InsurancePolicyNumber,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	p, err := h.patients.Create(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return patientPB(p), nil
}

func (h *HospitalHandler) GetPatient(ctx context.Context, in *pb.Patient) (*pb.Patient, error) {
	p, err := h.patients.Get(ctx, in.PatientId)
	if err != nil {
		return nil, toStatus(err)
	}
	return patientPB(p), nil
}

func (h *HospitalHandler) ListPatients(ctx context.Context, in *pb.ListPatientsRequest) (*pb.ListPatientsResponse, error) {
	items, total, err := h.patients.List(ctx, in.SearchTerm, page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListPatientsResponse{Patients: mapAll(items, patientPB), TotalCount: int32(total)}, nil
}

func (h *HospitalHandler) UpdatePatient(ctx context.Context, in *pb.Patient) (*pb.Patient, error) {
	m, err := patientModel(in.PatientId, patientFields{
		FirstName: in.FirstName, LastName: in.LastName, DateOfBirth: in.DateOfBirth, Gender: in.Gender,
		BloodType: in.BloodType, Address: in.Address, Phone: in.Phone, Email: in.Email,
		EmergencyName: in.EmergencyContactName, EmergencyPhone: in.EmergencyContactPhone,
		InsuranceProvider: in.InsuranceProvider, InsurancePol: in.InsurancePolicyNumber,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	p, err := h.patients.Update(ctx, m)
	if err != nil {
		return nil, toStatus(err)
	}
	return patientPB(p), nil
}

func (h *HospitalHandler) DeletePatient(ctx context.Context, in *pb.Patient) (*emptypb.Empty, error) {
	if err := h.patients.Delete(ctx, in.PatientId); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

// ---- medical records ----

func (h *HospitalHandler) CreateMedicalRecord(ctx context.Context, in *pb.CreateMedicalRecordRequest) (*pb.MedicalRecord, error) {
	r, err := h.medicalRecords.Create(ctx, &model.MedicalRecord{
		PatientID: in.PatientId, Diagnosis: in.Diagnosis, TreatmentPlan: in.TreatmentPlan,
		Notes: in.Notes, CreatedBy: optID(in.CreatedBy),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return medicalRecordPB(r), nil
}

func (h *HospitalHandler) GetMedicalRecord(ctx context.Context, in *pb.MedicalRecord) (*pb.MedicalRecord, error) {
	r, err := h.medicalRecords.Get(ctx, in.RecordId)
	if err != nil {
		return nil, toStatus(err)
	}
	return medicalRecordPB(r), nil
}

func (h *HospitalHandler) ListMedicalRecords(ctx context.Context, in *pb.ListMedicalRecordsRequest) (*pb.ListMedicalRecordsResponse, error) {
	items, total, err := h.medicalRecords.ListByPatient(ctx, in.PatientId, page(in.PageSize, in.PageNumber))
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ListMedicalRecordsResponse{Records: mapAll(items, medicalRecordPB), TotalCount: int32(total)}, nil
}

func (h *HospitalHandler) UpdateMedicalRecord(ctx context.Context, in *pb.MedicalRecord) (*pb.MedicalRecord, error) {
	r, err := h.medicalRecords.Update(ctx, &model.MedicalRecord{
		ID: in.RecordId, PatientID: in.PatientId, Diagnosis: in.Diagnosis, TreatmentPlan: in.TreatmentPlan,
		Notes: in.Notes, CreatedBy: optID(in.CreatedBy),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return medicalRecordPB(r), nil
}

func (h *HospitalHandler) DeleteMedicalRecord(ctx context.Context, in *pb.MedicalRecord) (*emptypb.Empty, error) {
	if err := h.medicalRecords.Delete(ctx, in.RecordId); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}
