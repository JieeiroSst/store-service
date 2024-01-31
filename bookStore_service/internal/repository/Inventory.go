package repository

import "github.com/jackc/pgx/v4"

type Inventory interface {
}

type InventoryRepo struct {
	conn *pgx.Conn
}

func NewInventoryRepo(conn *pgx.Conn) *InventoryRepo {
	return &InventoryRepo{
		conn: conn,
	}
}
