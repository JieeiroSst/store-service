package web

import "embed"

//go:embed index.html
var Index []byte

//go:embed static
var Static embed.FS
