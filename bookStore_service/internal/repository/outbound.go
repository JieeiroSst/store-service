package repository

import "github.com/jackc/pgx/v4"

type Outbound interface {
}

type OutboundRepo struct {
	conn *pgx.Conn
}

func NewOutboundRepo(conn *pgx.Conn) *OutboundRepo {
	return &OutboundRepo{
		conn: conn,
	}
}
