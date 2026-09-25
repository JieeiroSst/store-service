// Package renthouse exposes the database schema so the postgres adapter can apply it on startup.
package renthouse

import _ "embed"

//go:embed schema.sql
var Schema string
