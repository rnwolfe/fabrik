// Package migrations embeds the SQL migration files so they can be bundled
// into the server binary and accessed at runtime without the filesystem.
package migrations

import "embed"

// FS is the embedded filesystem containing all migration SQL files.
//
//go:embed *.sql
var FS embed.FS
