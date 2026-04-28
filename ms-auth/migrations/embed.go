package migrations

import "embed"

const Directory = "."

//go:embed *.sql
var Files embed.FS
