package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type IdentityService struct {
	patients port.PatientRepository
	gateway  port.IdentityGateway
}

func NewIdentityService(patients port.PatientRepository, gateway port.IdentityGateway) port.IdentityUsecase {
	return &IdentityService{patients: patients, gateway: gateway}
}

func (s *IdentityService) LinkUser(ctx context.Context, patientID int32, userID string) (*model.Patient, error) {
	userID = strings.TrimSpace(userID)
	p, err := s.patients.Get(ctx, patientID)
	if err != nil {
		return nil, err
	}
	p.UserID = nil
	if userID != "" {
		p.UserID = &userID
	}
	if err := s.patients.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *IdentityService) userFor(ctx context.Context, patientID int32) (string, error) {
	if !s.gateway.Enabled() {
		return "", fmt.Errorf("%w: identity verification is not configured", model.ErrUpstream)
	}
	p, err := s.patients.Get(ctx, patientID)
	if err != nil {
		return "", err
	}
	if p.UserID == nil {
		return "", model.Conflict("patient %d is not linked to a user account", patientID)
	}
	return *p.UserID, nil
}

func (s *IdentityService) Status(ctx context.Context, patientID int32) (*model.IdentityStatus, error) {
	uid, err := s.userFor(ctx, patientID)
	if err != nil {
		return nil, err
	}
	return s.gateway.Status(ctx, uid)
}

func (s *IdentityService) SubmitCitizenCard(ctx context.Context, patientID int32, front, back []byte) (*model.IdentityStatus, error) {
	if len(front) == 0 || len(back) == 0 {
		return nil, model.Invalid("both the front and back of the card are required")
	}
	uid, err := s.userFor(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if err := s.gateway.SubmitCitizenCard(ctx, uid, front, back); err != nil {
		return nil, err
	}
	return s.gateway.Status(ctx, uid)
}

func (s *IdentityService) SubmitFaceScan(ctx context.Context, patientID int32, frames [][]byte) (*model.IdentityStatus, error) {
	if len(frames) == 0 {
		return nil, model.Invalid("at least one face frame is required")
	}
	uid, err := s.userFor(ctx, patientID)
	if err != nil {
		return nil, err
	}
	if err := s.gateway.SubmitFaceScan(ctx, uid, frames); err != nil {
		return nil, err
	}
	return s.gateway.Status(ctx, uid)
}

func (s *IdentityService) Verify(ctx context.Context, patientID int32) (*model.IdentityStatus, error) {
	uid, err := s.userFor(ctx, patientID)
	if err != nil {
		return nil, err
	}
	verdict, err := s.gateway.Verify(ctx, uid)
	if err != nil {
		return nil, err
	}
	if full, err := s.gateway.Status(ctx, uid); err == nil {
		full.State, full.MatchScore, full.VerifiedAt = verdict.State, verdict.MatchScore, verdict.VerifiedAt
		return full, nil
	}
	return verdict, nil
}
