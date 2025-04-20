package http

import (
	"context"

	pb "github.com/JIeeiroSst/lib-gateway/hospital-patientm-anagement-service/gateway/hospital-patientm-anagement-service"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	pb.UnimplementedHospotalServiceServer
}

func (h *Handler) CreateDepartment(ctx context.Context, in *pb.CreateDepartmentRequest) (*pb.Department, error) {
	return nil, nil
}

func (h *Handler) GetDepartment(ctx context.Context, in *pb.Department) (*pb.Department, error) {
	return nil, nil
}

func (h *Handler) ListDepartments(ctx context.Context, in *pb.ListDepartmentsRequest) (*pb.ListDepartmentsResponse, error) {
	return nil, nil
}

func (h *Handler) UpdateDepartment(ctx context.Context, in *pb.Department) (*pb.Department, error) {
	return nil, nil
}

func (h *Handler) DeleteDepartment(ctx context.Context, in *pb.Department) (*emptypb.Empty, error) {
	return nil, nil
}
func (h *Handler) CreateStaff(ctx context.Context, in *pb.CreateStaffRequest) (*pb.Staff, error) {
	return nil, nil
}

func (h *Handler) GetStaff(ctx context.Context, in *pb.Staff) (*pb.Staff, error) {
	return nil, nil
}

func (h *Handler) ListStaff(ctx context.Context, in *pb.ListStaffRequest) (*pb.ListStaffResponse, error) {
	return nil, nil
}
func (h *Handler) UpdateStaff(ctx context.Context, in *pb.Staff) (*pb.Staff, error) {
	return nil, nil
}

func (h *Handler) DeleteStaff(ctx context.Context, in *pb.Staff) (*emptypb.Empty, error) {
	return nil, nil
}

func (h *Handler) CreatePatient(ctx context.Context, in *pb.CreatePatientRequest) (*pb.Patient, error) {
	return nil, nil
}

func (h *Handler) GetPatient(ctx context.Context, in *pb.Patient) (*pb.Patient, error) {
	return nil, nil
}

func (h *Handler) ListPatients(ctx context.Context, in *pb.ListPatientsRequest) (*pb.ListPatientsResponse, error) {
	return nil, nil
}

func (h *Handler) UpdatePatient(ctx context.Context, in *pb.Patient) (*pb.Patient, error) {
	return nil, nil
}

func (h *Handler) DeletePatient(ctx context.Context, in *pb.Patient) (*emptypb.Empty, error) {
	return nil, nil
}

func (h *Handler) CreateMedicalRecord(ctx context.Context, in *pb.CreateMedicalRecordRequest) (*pb.MedicalRecord, error) {
	return nil, nil
}

func (h *Handler) GetMedicalRecord(ctx context.Context, in *pb.MedicalRecord) (*pb.MedicalRecord, error) {
	return nil, nil
}

func (h *Handler) ListMedicalRecords(ctx context.Context, in *pb.ListMedicalRecordsRequest) (*pb.ListMedicalRecordsResponse, error) {
	return nil, nil
}

func (h *Handler) UpdateMedicalRecord(ctx context.Context, in *pb.MedicalRecord) (*pb.MedicalRecord, error) {
	return nil, nil
}

func (h *Handler) DeleteMedicalRecord(ctx context.Context, in *pb.MedicalRecord) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
