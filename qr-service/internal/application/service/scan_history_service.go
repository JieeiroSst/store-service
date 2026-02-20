package service

import (
	"context"

	"github.com/JIeeiroSst/qr-service/internal/domain/port"
	"go.uber.org/zap"
)

type scanHistoryService struct {
	scanRepo port.ScanHistoryRepository
	logger   *zap.Logger
}

func NewScanHistoryService(
	scanRepo port.ScanHistoryRepository,
	logger *zap.Logger,
) port.ScanHistoryService {
	return &scanHistoryService{
		scanRepo: scanRepo,
		logger:   logger,
	}
}

func (s *scanHistoryService) GetHistory(ctx context.Context, qrCodeID string, page, limit int64) (*port.ScanHistoryResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	data, total, err := s.scanRepo.GetByQRCodeID(ctx, qrCodeID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return &port.ScanHistoryResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *scanHistoryService) GetStats(ctx context.Context, qrCodeID string) (*port.ScanStats, error) {
	return s.scanRepo.GetStats(ctx, qrCodeID)
}
