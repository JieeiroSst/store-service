package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) QrQuery() generated.QrQueryResolver { return &qrQueryResolver{r} }

type qrQueryResolver struct{ *Resolver }

func (r *qrQueryResolver) QRCodes(ctx context.Context, obj *model.QRQuery, page *int, limit *int, status *string, typeArg *string, search *string, createdBy *string) (*model.QRQRListResponse, error) {
	return r.Clients.QRService.QRCodes(ctx, page, limit, status, typeArg, search, createdBy)
}

func (r *qrQueryResolver) QRCode(ctx context.Context, obj *model.QRQuery, id string) (*model.QRQRCode, error) {
	return r.Clients.QRService.QRCode(ctx, id)
}

func (r *qrQueryResolver) ScanHistory(ctx context.Context, obj *model.QRQuery, id string, page *int, limit *int) (*model.QRScanHistoryResponse, error) {
	return r.Clients.QRService.ScanHistory(ctx, id, page, limit)
}

func (r *qrQueryResolver) ScanStats(ctx context.Context, obj *model.QRQuery, id string) (*model.QRScanStats, error) {
	return r.Clients.QRService.ScanStats(ctx, id)
}
