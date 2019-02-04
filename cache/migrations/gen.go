package models

//go:generate rm -f cache.db
//go:generate sql-migrate up -env=cache-development
//go:generate sqlboiler -o ../models/ sqlite3
//go:generate rm cache.db

//-- go:generate sql-migrate up -env=cache-development-pg
//-- go:generate sqlboiler -o ../models/ psql
