package migrations

import "embed"

// FS contains SQL migrations bundled into the backend binary.
//
//go:embed *.sql
var FS embed.FS
