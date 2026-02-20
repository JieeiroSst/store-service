package port

import (
	"context"

	"github.com/JIeeiroSst/qr-service/internal/domain/entity"
)

type QRCodeService interface {
	Generate(ctx context.Context, req *GenerateQRRequest) (*QRResponse, error)
	GetByID(ctx context.Context, id string) (*entity.QRCode, error)
	GetByShortCode(ctx context.Context, shortCode string) (*entity.QRCode, error)
	List(ctx context.Context, filter QRCodeFilter) (*QRListResponse, error)
	Update(ctx context.Context, id string, req *UpdateQRRequest) (*entity.QRCode, error)
	UpdateContent(ctx context.Context, id string, req *UpdateContentRequest) (*entity.QRCode, error)
	Delete(ctx context.Context, id string) error
	Redirect(ctx context.Context, shortCode string, meta *ScanMeta) (string, error)
	Regenerate(ctx context.Context, id string) (*QRResponse, error)
}

type ScanHistoryService interface {
	GetHistory(ctx context.Context, qrCodeID string, page, limit int64) (*ScanHistoryResponse, error)
	GetStats(ctx context.Context, qrCodeID string) (*ScanStats, error)
}


type GenerateQRRequest struct {
	Title       string             `json:"title" binding:"required"`
	Type        entity.QRCodeType  `json:"type" binding:"required"`
	Content     string             `json:"content" binding:"required"`
	RedirectURL string             `json:"redirect_url"`
	Size        int                `json:"size"`
	ForeColor   string             `json:"fore_color"`
	BackColor   string             `json:"back_color"`
	ExpiresAt   *string            `json:"expires_at"`
	CreatedBy   string             `json:"created_by"`
}

type UpdateQRRequest struct {
	Title     string             `json:"title"`
	Status    entity.QRCodeStatus `json:"status"`
	ForeColor string             `json:"fore_color"`
	BackColor string             `json:"back_color"`
	ExpiresAt *string            `json:"expires_at"`
}

type UpdateContentRequest struct {
	Content     string `json:"content" binding:"required"`
	RedirectURL string `json:"redirect_url"`
}

type QRResponse struct {
	QRCode      *entity.QRCode `json:"qr_code"`
	ImageBase64 string         `json:"image_base64"`
	ScanURL     string         `json:"scan_url"`
}

type QRListResponse struct {
	Data       []*entity.QRCode `json:"data"`
	Total      int64            `json:"total"`
	Page       int64            `json:"page"`
	Limit      int64            `json:"limit"`
	TotalPages int64            `json:"total_pages"`
}

type ScanHistoryResponse struct {
	Data       []*entity.ScanHistory `json:"data"`
	Total      int64                 `json:"total"`
	Page       int64                 `json:"page"`
	Limit      int64                 `json:"limit"`
	TotalPages int64                 `json:"total_pages"`
}

type ScanMeta struct {
	IPAddress string
	UserAgent string
	Referer   string
}
