package main

import (
    "context"
    "fmt"
    "os"

    "github.com/jackc/pgx/v5"
)

func main() {
    ctx := context.Background()
    connStr := "postgres://timescaledb:password@localhost:5432/postgres?sslmode=disable"
    conn, err := pgx.Connect(ctx, connStr)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
        os.Exit(1)
    }
    defer conn.Close(ctx)

    var greeting string
    err = conn.QueryRow(ctx, "select 'Hello, Timescale!'").Scan(&greeting)
    if err != nil {
        fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(greeting)
}