package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type Postgresql struct {
	DNS string
}

func NewPostgresql(dns string) *pgx.Conn {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dns)
	if err != nil {
		fmt.Println("error error")
	}
	defer conn.Close(ctx)

	return conn
}
