package port

import (
	"context"
	"io"
	"time"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
)

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type UserClient interface {
	GetUser(ctx context.Context, userID string) (*UserInfo, error)
}

type CardReader interface {
	ReadMRZ(ctx context.Context, backImage []byte) (*model.MRZResult, error)
}

type FaceAnalyzer interface {
	Analyze(ctx context.Context, image []byte) (*model.FaceTemplate, error)
	AnalyzeFrames(ctx context.Context, frames [][]byte) (*model.FaceTemplate, error)
	Compare(a, b []byte) float64
}

type CitizenIdentityRepository interface {
	Create(ctx context.Context, identity *model.CitizenIdentity) error
	GetByUserID(ctx context.Context, userID string) (*model.CitizenIdentity, error)
}

type FaceBiometricRepository interface {
	Upsert(ctx context.Context, face *model.FaceBiometric) error
	GetByUserID(ctx context.Context, userID string) (*model.FaceBiometric, error)
}

type VerificationRepository interface {
	Upsert(ctx context.Context, v *model.EkycVerification) error
	GetByUserID(ctx context.Context, userID string) (*model.EkycVerification, error)
}

type ChipDump struct {
	EFCOM []byte
	DG1   []byte
	DG2   []byte
	EFSOD []byte
}

type NFCReader interface {
	ReadChip(ctx context.Context, dump ChipDump) (*model.ChipReadResult, error)
}

type ObjectStorage interface {
	PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, key string) ([]byte, error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}
