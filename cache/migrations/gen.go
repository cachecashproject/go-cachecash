package migrations

//go:generate rm -f cache.db
//go:generate sql-migrate up -env=cache-development
//go:generate sqlboiler -o ../models/ sqlite3
//go:generate mv cache.db ../models/

//-- go:generate sql-migrate up -env=cache-development-pg
//-- go:generate sqlboiler -o ../models/ psql
