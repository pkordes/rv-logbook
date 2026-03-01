// Package migrations embeds the SQL migration files so they can be used
// by the goose programmatic API in tests and server bootstrap.
package migrations

import "embed"

// FS holds all *.sql migration files embedded at compile time.
// Pass this to goose.UpFS / goose.DownToFS instead of relying on
// a filesystem path at runtime.
//
//go:embed *.sql
var FS embed.FS
