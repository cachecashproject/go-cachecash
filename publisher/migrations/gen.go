package migrations

//go:generate ./gen.sh

import (
	"github.com/gobuffalo/packr"
	migrate "github.com/rubenv/sql-migrate"
)

var (
	Migrations = &migrate.PackrMigrationSource{
		Box: packr.NewBox("."),
	}
)
