package application

import (
	"context"
	"fmt"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type DocumentService struct {
	patients port.PatientRepository
	gateway  port.DocumentGateway
}

func NewDocumentService(patients port.PatientRepository, gateway port.DocumentGateway) port.DocumentUsecase {
	return &DocumentService{patients: patients, gateway: gateway}
}

func (s *DocumentService) check(ctx context.Context, patientID int32) error {
	if !s.gateway.Enabled() {
		return fmt.Errorf("%w: document storage is not configured", model.ErrUpstream)
	}
	if err := positive("patient_id", patientID); err != nil {
		return err
	}
	_, err := s.patients.Get(ctx, patientID)
	return err
}

func (s *DocumentService) Upload(ctx context.Context, patientID int32, in port.DocumentUpload) (*model.Document, error) {
	if err := s.check(ctx, patientID); err != nil {
		return nil, err
	}
	return s.gateway.Upload(ctx, patientID, in)
}

func (s *DocumentService) List(ctx context.Context, patientID int32, limit, offset int) ([]model.Document, int64, error) {
	if err := s.check(ctx, patientID); err != nil {
		return nil, 0, err
	}
	return s.gateway.List(ctx, patientID, limit, offset)
}

func (s *DocumentService) Open(ctx context.Context, patientID int32, docID string) (*port.DocumentDownload, error) {
	if err := s.check(ctx, patientID); err != nil {
		return nil, err
	}
	return s.gateway.Open(ctx, patientID, docID)
}

func (s *DocumentService) Delete(ctx context.Context, patientID int32, docID string) error {
	if err := s.check(ctx, patientID); err != nil {
		return err
	}
	return s.gateway.Delete(ctx, patientID, docID)
}
