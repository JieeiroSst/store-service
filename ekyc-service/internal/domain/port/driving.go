package port

import (
	"context"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
)

type EkycUsecase interface {
	SubmitCitizenCard(ctx context.Context, userID string, front, back []byte) (*model.CitizenIdentity, error)
	SubmitFaceScan(ctx context.Context, userID string, frames [][]byte) (*model.FaceBiometric, error)
	SubmitNFCChip(ctx context.Context, userID string, dump ChipDump) (*model.CitizenIdentity, error)
	Verify(ctx context.Context, userID string) (*model.EkycVerification, error)
	GetStatus(ctx context.Context, userID string) (*model.EkycStatus, error)
}
