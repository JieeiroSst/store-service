package file

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type Service struct {
	importer StockImporter
	reports  ReportSource
}

func NewService(importer StockImporter, reports ReportSource) FileService {
	return &Service{importer: importer, reports: reports}
}

func (s *Service) ImportStockCSV(ctx context.Context, in ImportStockInput) (*ImportResult, error) {
	reader := csv.NewReader(in.Reader)
	reader.FieldsPerRecord = -1

	result := &ImportResult{}
	var batch [][2]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			result.Skipped++
			continue
		}
		if len(record) == 0 || record[0] == "" {
			result.Skipped++
			continue
		}
		pin := ""
		if len(record) > 1 {
			pin = record[1]
		}
		batch = append(batch, [2]string{record[0], pin})
	}

	if len(batch) > 0 {
		batchID := shared.NewVoucherID().String()
		imported, err := s.importer.ImportCodes(ctx, in.MerchantID, in.ProductSKU, batch, batchID)
		if err != nil {
			return nil, err
		}
		result.Imported = imported
		result.Skipped += len(batch) - imported
	}
	return result, nil
}

func (s *Service) ExportReportCSV(ctx context.Context, in ExportReportInput, w io.Writer) error {
	if in.ReportType != "redemptions" {
		return fmt.Errorf("unsupported report type %q", in.ReportType)
	}
	rows, err := s.reports.RedemptionRows(ctx, in.From, in.To)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(w)
	defer writer.Flush()
	return writer.WriteAll(rows)
}
