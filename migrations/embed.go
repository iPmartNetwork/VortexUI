// Package migrations embeds the SQL migration files so the panel binary can
// apply them on startup — no separate goose binary or mounted files needed in
// production images.
package migrations

import "embed"

// FS holds every migration, embedded at build time.
//
//go:embed *.sql
var FS embed.FS
