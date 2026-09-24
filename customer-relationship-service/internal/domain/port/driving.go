package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
)

type Usecase[T any] interface {
	Create(ctx context.Context, entity *T) (*T, error)
	Get(ctx context.Context, id uint) (*T, error)
	List(ctx context.Context, q ListQuery) ([]T, int64, error)
	Update(ctx context.Context, id uint, entity *T) (*T, error)
	Delete(ctx context.Context, id uint) error
}

type ConvertLeadInput struct {
	AccountName            string  `json:"account_name"`
	CreateOpportunity      bool    `json:"create_opportunity"`
	OpportunityDescription string  `json:"opportunity_description"`
	OpportunityAmount      float64 `json:"opportunity_amount"`
}

type ConvertLeadResult struct {
	Lead        *model.Lead        `json:"lead"`
	Account     *model.Account     `json:"account"`
	Contact     *model.Contact     `json:"contact"`
	Opportunity *model.Opportunity `json:"opportunity,omitempty"`
}

type LeadConverter interface {
	Convert(ctx context.Context, leadID uint, in ConvertLeadInput) (*ConvertLeadResult, error)
}

type OpportunityWorkflow interface {
	ChangeStage(ctx context.Context, id uint, stage string) (*model.Opportunity, error)
}

type SignContractInput struct {
	SignedBy    string    `json:"signed_by" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	SigningTime time.Time `json:"signing_time" binding:"required"`
	Algorithm   string    `json:"algorithm"`
	Signature   string    `json:"signature"`
	Certificate string    `json:"certificate"`
	SignedPDF   string    `json:"signed_pdf"`
}

type ContractWorkflow interface {
	Submit(ctx context.Context, id uint) (*model.Contract, error)
	Approve(ctx context.Context, id uint) (*model.Contract, error)
	Reject(ctx context.Context, id uint) (*model.Contract, error)
	Sign(ctx context.Context, id uint, in SignContractInput) (*model.Contract, error)
	SigningPayload(ctx context.Context, id uint, signedBy string, endDate, signingTime time.Time) ([]byte, error)
	Document(ctx context.Context, id uint, signedBy string, endDate, signingTime time.Time) ([]byte, error)
	SignedDocument(ctx context.Context, id uint) ([]byte, error)
}

type ContractLifecycleState struct {
	Exists  bool       `json:"exists"`
	Status  string     `json:"status"`
	EndDate *time.Time `json:"end_date"`
}

type ContractLifecycleUsecase interface {
	State(ctx context.Context, id uint) (ContractLifecycleState, error)
	Expire(ctx context.Context, id uint, at time.Time) (bool, error)
	Remind(ctx context.Context, id uint, remaining time.Duration) error
	Resync(ctx context.Context) (int, error)
}

type FileUpload struct {
	ContractID  uint
	Kind        string
	Name        string
	Description string
	Data        []byte
	ReplacesID  uint
}

type FileListFilter struct {
	Kind        string
	AllVersions bool
	Offset      int
	Limit       int
}

type FileAuditFilter struct {
	FileID uint
	Action string
	Offset int
	Limit  int
}

type FileHistory struct {
	Versions []model.ContractFile      `json:"versions"`
	Events   []model.ContractFileEvent `json:"events"`
}

type ContractFileUsecase interface {
	Upload(ctx context.Context, in FileUpload) (*model.ContractFile, error)
	List(ctx context.Context, contractID uint, f FileListFilter) ([]model.ContractFile, int64, error)
	Get(ctx context.Context, contractID, fileID uint) (*model.ContractFile, error)
	Download(ctx context.Context, contractID, fileID uint) (*model.ContractFile, []byte, error)
	Delete(ctx context.Context, contractID, fileID uint) error
	History(ctx context.Context, contractID, fileID uint) (*FileHistory, error)
	Audit(ctx context.Context, contractID uint, f FileAuditFilter) ([]model.ContractFileEvent, int64, error)
}

type FileSpec struct {
	ContractID  uint
	Kind        string
	Name        string
	Description string
	UploadedBy  string
	UploaderID  string
	Source      string
	Data        []byte
}

type ContractFileKeeper interface {
	Prepare(ctx context.Context, spec FileSpec) (*model.ContractFile, error)
	Persist(ctx context.Context, f *model.ContractFile) error
	Discard(ctx context.Context, f *model.ContractFile)
	Latest(ctx context.Context, contractID uint, kind string) (*model.ContractFile, error)
	Open(ctx context.Context, f *model.ContractFile) ([]byte, error)
}

type ContractExpiryUsecase interface {
	ExpireOverdue(ctx context.Context) (int64, error)
}

type CaseWorkflow interface {
	Close(ctx context.Context, id uint) (*model.Case, error)
}

type AccountOverview struct {
	Account       *model.Account      `json:"account"`
	Contacts      []model.Contact     `json:"contacts"`
	Contracts     []model.Contract    `json:"contracts"`
	Opportunities []model.Opportunity `json:"opportunities"`
}

type AccountOverviewUsecase interface {
	Overview(ctx context.Context, accountID uint) (*AccountOverview, error)
}

type ReportUsecase interface {
	Summary(ctx context.Context) (*model.Summary, error)
}
