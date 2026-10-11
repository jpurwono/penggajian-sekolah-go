// Package migrations meng-embed seluruh file SQL migrasi ke binary,
// sehingga deployment tidak butuh folder migrasi terpisah.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
