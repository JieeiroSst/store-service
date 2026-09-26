package inbound

import (
	"context"
	"io"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type DocumentFile struct {
	Name string
	Size int64
	Body io.ReadCloser
}

type DocumentUseCase interface {
	InvoicePDF(ctx context.Context, actor Principal, orderID int64) ([]byte, error)
	TicketPDF(ctx context.Context, actor Principal, ticketID int64) ([]byte, error)
	List(ctx context.Context, actor Principal, orderID int64) ([]domain.OrderDocument, error)
	Open(ctx context.Context, actor Principal, orderID int64, kind string) (DocumentFile, error)
	Generate(ctx context.Context, orderID int64) (bool, error)
	GenerateMissing(ctx context.Context) (int, error)
}
