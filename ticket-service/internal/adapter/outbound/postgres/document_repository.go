package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type DocumentRepository struct{ db *pgxpool.Pool }

var _ outbound.DocumentRepository = (*DocumentRepository)(nil)

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository { return &DocumentRepository{db: db} }

func (r *DocumentRepository) Save(ctx context.Context, d domain.OrderDocument) (bool, error) {
	tag, err := r.db.Exec(ctx, `insert into order_document (order_id, kind, file_id, file_name, size) values ($1, $2, $3, $4, $5)
		on conflict do nothing`, d.OrderID, d.Kind, d.FileID, d.FileName, d.Size)
	return tag.RowsAffected() == 1, mapErr(err)
}

func (r *DocumentRepository) List(ctx context.Context, orderID int64) ([]domain.OrderDocument, error) {
	rows, err := r.db.Query(ctx, `select order_id, kind, file_id, file_name, size, created_at from order_document where order_id = $1 order by kind`, orderID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.OrderDocument{}
	for rows.Next() {
		var d domain.OrderDocument
		if err := rows.Scan(&d.OrderID, &d.Kind, &d.FileID, &d.FileName, &d.Size, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, mapErr(rows.Err())
}

func (r *DocumentRepository) Get(ctx context.Context, orderID int64, kind string) (domain.OrderDocument, error) {
	var d domain.OrderDocument
	err := r.db.QueryRow(ctx, `select order_id, kind, file_id, file_name, size, created_at from order_document where order_id = $1 and kind = $2`,
		orderID, kind).Scan(&d.OrderID, &d.Kind, &d.FileID, &d.FileName, &d.Size, &d.CreatedAt)
	return d, mapErr(err)
}

func (r *DocumentRepository) MarkDone(ctx context.Context, orderID int64) error {
	_, err := r.db.Exec(ctx, `update ticket_order set documents_at = now() where id = $1 and documents_at is null`, orderID)
	return mapErr(err)
}

func (r *DocumentRepository) Pending(ctx context.Context, limit int) ([]int64, error) {
	rows, err := r.db.Query(ctx, `select id from ticket_order where status = $1 and documents_at is null order by id limit $2`, int16(domain.OrderPaid), limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, mapErr(rows.Err())
}
