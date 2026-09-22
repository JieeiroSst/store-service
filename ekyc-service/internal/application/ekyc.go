package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/ekyc-service/config"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type ekycService struct {
	cfg           *config.Config
	users         port.UserClient
	cardReader    port.CardReader
	faceAnalyzer  port.FaceAnalyzer
	nfcReader     port.NFCReader
	storage       port.ObjectStorage
	identities    port.CitizenIdentityRepository
	faces         port.FaceBiometricRepository
	verifications port.VerificationRepository
}

func NewEkycService(
	cfg *config.Config,
	users port.UserClient,
	cardReader port.CardReader,
	faceAnalyzer port.FaceAnalyzer,
	nfcReader port.NFCReader,
	storage port.ObjectStorage,
	identities port.CitizenIdentityRepository,
	faces port.FaceBiometricRepository,
	verifications port.VerificationRepository,
) port.EkycUsecase {
	return &ekycService{
		cfg:           cfg,
		users:         users,
		cardReader:    cardReader,
		faceAnalyzer:  faceAnalyzer,
		nfcReader:     nfcReader,
		storage:       storage,
		identities:    identities,
		faces:         faces,
		verifications: verifications,
	}
}

func (s *ekycService) SubmitCitizenCard(ctx context.Context, userID string, front, back []byte) (*model.CitizenIdentity, error) {
	if _, err := s.users.GetUser(ctx, userID); err != nil {
		return nil, err
	}

	mrz, err := s.cardReader.ReadMRZ(ctx, back)
	if err != nil {
		return nil, err
	}
	if mrz.Confidence < s.cfg.Ekyc.MRZMinConfidence {
		return nil, port.ErrMRZNotReadable
	}

	frontKey, err := s.putImage(ctx, userID, "front.jpg", front)
	if err != nil {
		return nil, err
	}
	backKey, err := s.putImage(ctx, userID, "back.jpg", back)
	if err != nil {
		return nil, err
	}

	identity := &model.CitizenIdentity{
		ID:             uuid.NewString(),
		UserID:         userID,
		Source:         model.IdentitySourceCardOCR,
		DocumentNumber: mrz.DocumentNumber,
		Surname:        mrz.Surname,
		GivenNames:     mrz.GivenNames,
		Nationality:    mrz.Nationality,
		DateOfBirth:    mrz.DateOfBirth,
		Sex:            mrz.Sex,
		DateOfExpiry:   mrz.DateOfExpiry,
		MRZLine1:       mrz.Line1,
		MRZLine2:       mrz.Line2,
		MRZLine3:       mrz.Line3,
		ChecksumValid:  mrz.ChecksumValid,
		Confidence:     mrz.Confidence,
		FrontImageKey:  frontKey,
		BackImageKey:   backKey,
	}
	if err := s.identities.Create(ctx, identity); err != nil {
		return nil, fmt.Errorf("store citizen identity: %w", err)
	}
	return identity, nil
}

func (s *ekycService) SubmitNFCChip(ctx context.Context, userID string, dump port.ChipDump) (*model.CitizenIdentity, error) {
	if _, err := s.users.GetUser(ctx, userID); err != nil {
		return nil, err
	}

	chip, err := s.nfcReader.ReadChip(ctx, dump)
	if err != nil {
		return nil, err
	}

	identity := &model.CitizenIdentity{
		ID:             uuid.NewString(),
		UserID:         userID,
		Source:         model.IdentitySourceNFCChip,
		DocumentNumber: chip.DocumentNumber,
		Surname:        chip.Surname,
		GivenNames:     chip.GivenNames,
		Nationality:    chip.Nationality,
		DateOfBirth:    chip.DateOfBirth,
		Sex:            chip.Sex,
		DateOfExpiry:   chip.DateOfExpiry,
		ChecksumValid:  chip.MRZChecksumValid,
		Confidence:     1.0,
		NFCVerified:    chip.AllHashesValid() && chip.SODSignatureSelfConsistent,
	}

	if len(chip.FaceImage) > 0 {
		key, err := s.putImage(ctx, userID, "chip-face.jpg", chip.FaceImage)
		if err != nil {
			return nil, err
		}
		identity.ChipFaceImageKey = key
	}

	if err := s.identities.Create(ctx, identity); err != nil {
		return nil, fmt.Errorf("store citizen identity: %w", err)
	}
	return identity, nil
}

func (s *ekycService) SubmitFaceScan(ctx context.Context, userID string, frames [][]byte) (*model.FaceBiometric, error) {
	if _, err := s.users.GetUser(ctx, userID); err != nil {
		return nil, err
	}

	tmpl, err := s.faceAnalyzer.AnalyzeFrames(ctx, frames)
	if err != nil {
		return nil, err
	}
	if tmpl.QualityScore < s.cfg.Ekyc.FaceMinQuality {
		return nil, fmt.Errorf("face scan quality %.2f below minimum %.2f: capture in better lighting/focus", tmpl.QualityScore, s.cfg.Ekyc.FaceMinQuality)
	}

	imageKey, err := s.putImage(ctx, userID, "face-scan.jpg", tmpl.AlignedImage)
	if err != nil {
		return nil, err
	}

	face := &model.FaceBiometric{
		ID:           uuid.NewString(),
		UserID:       userID,
		Template:     tmpl.Template,
		FaceImageKey: imageKey,
		QualityScore: tmpl.QualityScore,
	}
	if err := s.faces.Upsert(ctx, face); err != nil {
		return nil, fmt.Errorf("store face biometric: %w", err)
	}
	return face, nil
}

func (s *ekycService) Verify(ctx context.Context, userID string) (*model.EkycVerification, error) {
	identity, err := s.identities.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	face, err := s.faces.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	portraitKey := identity.FrontImageKey
	if identity.Source == model.IdentitySourceNFCChip && identity.ChipFaceImageKey != "" {
		portraitKey = identity.ChipFaceImageKey
	}
	if portraitKey == "" {
		return nil, fmt.Errorf("no reference portrait on file for user %s", userID)
	}

	portraitImage, err := s.storage.GetObject(ctx, portraitKey)
	if err != nil {
		return nil, fmt.Errorf("fetch reference portrait: %w", err)
	}
	portrait, err := s.faceAnalyzer.Analyze(ctx, portraitImage)
	if err != nil {
		return nil, fmt.Errorf("analyze reference portrait: %w", err)
	}

	score := s.faceAnalyzer.Compare(face.Template, portrait.Template)
	status := model.VerificationFailed
	var verifiedAt *time.Time
	if score >= s.cfg.Ekyc.FaceMatchThreshold && identity.ChecksumValid {
		status = model.VerificationVerified
		now := time.Now()
		verifiedAt = &now
	}

	v := &model.EkycVerification{
		ID:         uuid.NewString(),
		UserID:     userID,
		Status:     status,
		MatchScore: score,
		VerifiedAt: verifiedAt,
	}
	if err := s.verifications.Upsert(ctx, v); err != nil {
		return nil, fmt.Errorf("store verification: %w", err)
	}
	return v, nil
}

func (s *ekycService) GetStatus(ctx context.Context, userID string) (*model.EkycStatus, error) {
	status := &model.EkycStatus{UserID: userID}

	identity, err := s.identities.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, port.ErrIdentityNotFound) {
		return nil, err
	}
	status.Identity = identity

	face, err := s.faces.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, port.ErrFaceNotFound) {
		return nil, err
	}
	status.Face = face

	verification, err := s.verifications.GetByUserID(ctx, userID)
	if err != nil && !errors.Is(err, port.ErrVerificationNotFound) {
		return nil, err
	}
	status.Verification = verification

	return status, nil
}

func (s *ekycService) putImage(ctx context.Context, userID, name string, data []byte) (string, error) {
	key := fmt.Sprintf("citizen/%s/%s", userID, name)
	if err := s.storage.PutObject(ctx, key, bytes.NewReader(data), int64(len(data)), "image/jpeg"); err != nil {
		return "", fmt.Errorf("upload %s: %w", name, err)
	}
	return key, nil
}
