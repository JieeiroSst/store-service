package postgres

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/jackc/pgx/v5"
)

type reportRepository struct{ db *DB }

func NewReportRepository(db *DB) port.ReportRepository { return &reportRepository{db: db} }

func (r *reportRepository) SalesByProduct(ctx context.Context, machineID string, from, to time.Time) ([]domain.ProductSales, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		SELECT p.product_id, p.name, COUNT(*), COALESCE(SUM(pay.amount_cents), 0), COALESCE(SUM(pay.discount_cents), 0)
		FROM orders o
		JOIN sessions s ON s.session_id = o.session_id
		JOIN payments pay ON pay.payment_id = o.payment_id
		JOIN reservations res ON res.reservation_id = o.reservation_id
		JOIN inventory i ON i.inventory_id = res.inventory_id
		JOIN products p ON p.product_id = i.product_id
		WHERE s.machine_id = $1 AND o.status = 'completed' AND o.fulfilled_at >= $2 AND o.fulfilled_at < $3
		GROUP BY p.product_id, p.name
		ORDER BY SUM(pay.amount_cents) DESC, p.name`, machineID, from, to)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(row pgx.Row) (domain.ProductSales, error) {
		var s domain.ProductSales
		err := row.Scan(&s.ProductID, &s.ProductName, &s.Units, &s.RevenueCents, &s.DiscountCents)
		return s, err
	})
}
