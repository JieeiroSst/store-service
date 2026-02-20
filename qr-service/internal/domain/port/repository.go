package port

import (
	"context"

	"github.com/JIeeiroSst/qr-service/internal/domain/entity"
)

type QRCodeRepository interface {
	Create(ctx context.Context, qr *entity.QRCode) (*entity.QRCode, error)
	GetByID(ctx context.Context, id string) (*entity.QRCode, error)
	GetByShortCode(ctx context.Context, shortCode string) (*entity.QRCode, error)
	List(ctx context.Context, filter QRCodeFilter) ([]*entity.QRCode, int64, error)
	Update(ctx context.Context, id string, qr *entity.QRCode) (*entity.QRCode, error)
	UpdateContent(ctx context.Context, id string, content string, redirectURL string) (*entity.QRCode, error)
	Delete(ctx context.Context, id string) error
	IncrementScanCount(ctx context.Context, id string) error
}

type ScanHistoryRepository interface {
	Create(ctx context.Context, history *entity.ScanHistory) error
	GetByQRCodeID(ctx context.Context, qrCodeID string, page, limit int64) ([]*entity.ScanHistory, int64, error)
	GetStats(ctx context.Context, qrCodeID string) (*ScanStats, error)
	DeleteByQRCodeID(ctx context.Context, qrCodeID string) error
}

type QRCodeFilter struct {
	Status    string
	Type      string
	Search    string
	Page      int64
	Limit     int64
	CreatedBy string
}

type ScanStats struct {
	TotalScans     int64            `json:"total_scans"`
	UniqueIPs      int64            `json:"unique_ips"`
	DeviceBreakdown map[string]int64 `json:"device_breakdown"`
	OSBreakdown    map[string]int64 `json:"os_breakdown"`
	DailyScans     []DailyScan      `json:"daily_scans"`
}

type DailyScan struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
