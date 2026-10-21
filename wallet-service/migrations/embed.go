package migrations

import "embed"

//go:embed *.sql
var Schema embed.FS

//go:embed seed/*.sql
var Seed embed.FS
