package repository

import "github.com/jackc/pgx/v4"

type Book interface {
}

type BookRepo struct {
	conn *pgx.Conn
}

func NewBookRepo(conn *pgx.Conn) *BookRepo {
	return &BookRepo{
		conn: conn,
	}
}
