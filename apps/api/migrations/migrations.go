// Package migrations carries the migrations inside the binary: numbered SQL
// files, applied by the `migrate` command and never edited once they have run
// outside a development machine. Rules: Database in CODING_STANDARDS.md.
package migrations

import "embed"

// Files holds every migration, each named by its zero-padded number.
//
//go:embed *.sql
var Files embed.FS
