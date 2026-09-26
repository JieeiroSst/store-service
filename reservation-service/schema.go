// Package reservation exposes the database schema so the postgres adapter can apply it on startup.
package reservation

import _ "embed"

//go:embed schema.sql
var Schema string
